package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dis-migration-service/api"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/migrator"
	"github.com/ONSdigital/dis-migration-service/store"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

// Service contains all the configs, server and clients to run the API
type Service struct {
	API         *api.MigrationAPI
	Config      *config.Config
	HealthCheck HealthChecker
	Router      *mux.Router
	Server      HTTPServer
	ServiceList *ExternalServiceList
	mongoDB     store.MongoDB
	migrator    migrator.Migrator
	clients     *clients.ClientList
}

type MigrationServiceStore struct {
	store.MongoDB
}

func New(cfg *config.Config, serviceList *ExternalServiceList) *Service {
	svc := &Service{
		Config:      cfg,
		ServiceList: serviceList,
	}
	return svc
}

// Run the service
func (svc *Service) Run(ctx context.Context, buildTime, gitCommit, version string, svcErrors chan error) (err error) {
	log.Info(ctx, "running service")
	log.Info(ctx, "using service configuration", log.Data{"config": svc.Config})

	// Get MongoDB client
	//mongoDB, err := serviceList.GetMongoDB(ctx, cfg.MongoConfig)
	//if err != nil {
	//	log.Fatal(ctx, "failed to initialise mongo DB", err)
	//	return nil, err
	//}

	// Get MongoDB client
	svc.mongoDB, err = svc.ServiceList.GetMongoDB(ctx, svc.Config.MongoConfig)
	if err != nil {
		log.Fatal(ctx, "failed to initialise mongo DB", err)
		return err
	}

	// Get Datastore
	datastore := store.Datastore{Backend: MigrationServiceStore{svc.mongoDB}}

	// Get app clients
	svc.clients = svc.ServiceList.GetAppClients(ctx, svc.Config)

	// Get Migrator
	//migr, err := serviceList.GetMigrator(ctx)
	svc.migrator, err = svc.ServiceList.GetMigrator(ctx, datastore, svc.clients)
	if err != nil {
		log.Fatal(ctx, "failed to initialise migrator", err)
		return err
	}

	// Setup healthcheck
	svc.HealthCheck, err = svc.ServiceList.GetHealthCheck(svc.Config, buildTime, gitCommit, version)
	if err != nil {
		log.Fatal(ctx, "could not instantiate healthcheck", err)
		return err
	}

	if err := registerCheckers(ctx, svc.HealthCheck, svc.mongoDB); err != nil {
		return errors.Wrap(err, "unable to register checkers")
	}

	// Initialise the router
	r := mux.NewRouter()

	r.StrictSlash(true).Path("/health").HandlerFunc(svc.HealthCheck.Handler)
	svc.HealthCheck.Start(ctx)

	if svc.Config.OtelEnabled {
		r.Use(otelmux.Middleware(svc.Config.OTServiceName))
		// TODO: Any middleware will require 'otelhttp.NewMiddleware(cfg.OTServiceName),' included for Open Telemetry
	}

	middleware := createMiddleware(svc.HealthCheck)
	svc.Server = svc.ServiceList.GetHTTPServer(svc.Config.BindAddr, middleware.Then(r))

	// Set up the API
	svc.API = api.Setup(ctx, r, &datastore, svc.migrator)

	// Run the http server in a new go-routine
	go func() {
		if err := svc.Server.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	//return &Service{
	//	Config:      cfg,
	//	Router:      r,
	//	API:         a,
	//	HealthCheck: hc,
	//	ServiceList: serviceList,
	//	Server:      s,
	//	mongoDB:     mongoDB,
	//	migrator:    migr,
	//}, nil
	return nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.Config.GracefulShutdownTimeout
	log.Info(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout})
	shutdownContext, cancel := context.WithTimeout(ctx, timeout)

	// track shutown gracefully closes up
	var hasShutdownError bool

	go func() {
		defer cancel()

		// stop healthcheck, as it depends on everything else
		if svc.ServiceList.HealthCheck {
			svc.HealthCheck.Stop()
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.Server.Shutdown(shutdownContext); err != nil {
			log.Error(ctx, "failed to shutdown http server", err)
			hasShutdownError = true
		}

		// Close Migrator
		if svc.ServiceList.Migrator {
			if err := svc.migrator.Shutdown(shutdownContext); err != nil {
				log.Error(shutdownContext, "failed to close migrator", err)
				hasShutdownError = true
			}
		}

		// Close MongoDB (if it exists)
		if svc.ServiceList.MongoDB {
			if err := svc.mongoDB.Close(shutdownContext); err != nil {
				log.Error(shutdownContext, "failed to close mongo db session", err)
				hasShutdownError = true
			}
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-shutdownContext.Done()

	// timeout expired
	if shutdownContext.Err() == context.DeadlineExceeded {
		log.Error(shutdownContext, "shutdown timed out", ctx.Err())
		return shutdownContext.Err()
	}

	// other error
	if hasShutdownError {
		err := errors.New("failed to shutdown gracefully")
		log.Error(shutdownContext, "failed to shutdown gracefully ", err)
		return err
	}

	log.Info(shutdownContext, "graceful shutdown was successful")
	return nil
}

// CreateMiddleware creates an Alice middleware chain of handlers
// to forward collectionID from cookie from header
func createMiddleware(hc HealthChecker) alice.Chain {
	// healthcheck
	healthcheckHandler := healthcheckMiddleware(hc.Handler, "/health")
	middleware := alice.New(healthcheckHandler)

	return middleware
}

// healthcheckMiddleware creates a new http.Handler to intercept /health requests.
func healthcheckMiddleware(healthcheckHandler func(http.ResponseWriter, *http.Request), path string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if (req.Method == "GET" || req.Method == "HEAD") && req.URL.Path == path {
				healthcheckHandler(w, req)
				return
			}

			h.ServeHTTP(w, req)
		})
	}
}

func registerCheckers(ctx context.Context, hc HealthChecker, mongoCli store.MongoDB) (err error) {
	hasErrors := false

	if err = hc.AddCheck("Mongo DB", mongoCli.Checker); err != nil {
		hasErrors = true
		log.Error(ctx, "error adding check for mongo db", err)
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
