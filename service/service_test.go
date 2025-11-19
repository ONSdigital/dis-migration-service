package service_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	authorisationMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/migrator"
	migratorMock "github.com/ONSdigital/dis-migration-service/migrator/mock"
	"github.com/ONSdigital/dis-migration-service/service"
	"github.com/ONSdigital/dis-migration-service/service/mock"
	"github.com/ONSdigital/dis-migration-service/store"

	storeMock "github.com/ONSdigital/dis-migration-service/store/mock"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "BuildTime"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
	errServer     = errors.New("HTTP Server error")
)

var (
	errHealthcheck = errors.New("healthCheck error")
)

var funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
	return nil, errHealthcheck
}

var funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) service.HTTPServer {
	return nil
}

func TestRun(t *testing.T) {
	Convey("Having a set of mocked dependencies", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)

		authorisationMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		hcMock := &mock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		serverWg := &sync.WaitGroup{}
		serverMock := &mock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return nil
			},
		}

		failingServerMock := &mock.HTTPServerMock{
			ListenAndServeFunc: func() error {
				serverWg.Done()
				return errServer
			},
		}

		funcDoGetHealthcheckOk := func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return serverMock
		}

		funcDoGetFailingHTTPServer := func(bindAddr string, router http.Handler) service.HTTPServer {
			return failingServerMock
		}

		funcDoGetMongoDBOk := func(ctx context.Context, cfg config.MongoConfig) (store.MongoDB, error) {
			return &storeMock.MongoDBMock{}, nil
		}

		funcDoGetMigrator := func(ctx context.Context, jobService application.JobService, clientList *clients.ClientList) (migrator.Migrator, error) {
			return &migratorMock.MigratorMock{}, nil
		}

		funcDoGetAppClientsOk := func(context.Context, *config.Config) *clients.ClientList {
			return &clients.ClientList{}
		}

		funcDoGetAuthOk := func(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
			return authorisationMiddleware, nil
		}

		Convey("Given that initialising healthcheck returns an error", func() {
			// setup (run before each `Convey` at this scope / indentation):
			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc:  funcDoGetHTTPServerNil,
				DoGetHealthCheckFunc: funcDoGetHealthcheckErr,
				DoGetMigratorFunc:    funcDoGetMigrator,
				DoGetMongoDBFunc:     funcDoGetMongoDBOk,
				DoGetAppClientsFunc:  funcDoGetAppClientsOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)

			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.HealthCheck, ShouldBeFalse)
			})

			Reset(func() {
				// This reset is run after each `Convey` at the same scope (indentation)
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {
			// setup (run before each `Convey` at this scope / indentation):
			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc:              funcDoGetHTTPServer,
				DoGetHealthCheckFunc:             funcDoGetHealthcheckOk,
				DoGetMigratorFunc:                funcDoGetMigrator,
				DoGetMongoDBFunc:                 funcDoGetMongoDBOk,
				DoGetAppClientsFunc:              funcDoGetAppClientsOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)

			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.HealthCheck, ShouldBeTrue)
			})

			Convey("The checkers are registered and the healthcheck and http server started", func() {
				So(len(hcMock.AddCheckCalls()), ShouldEqual, 1)
				So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, "localhost:30100")
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				//!!! a call needed to stop the server, maybe ?
				serverWg.Wait() // Wait for HTTP server go-routine to finish
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})

			Reset(func() {
				// This reset is run after each `Convey` at the same scope (indentation)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {
			// setup (run before each `Convey` at this scope / indentation):
			errAddCheckFail := errors.New("Error(s) registering checkers for healthcheck")
			hcMockAddFail := &mock.HealthCheckerMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddCheckFail },
				StartFunc:    func(ctx context.Context) {},
			}

			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: funcDoGetHTTPServerNil,
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMockAddFail, nil
				},
				DoGetMongoDBFunc:                 funcDoGetMongoDBOk,
				DoGetMigratorFunc:                funcDoGetMigrator,
				DoGetAppClientsFunc:              funcDoGetAppClientsOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails, but all checks try to register", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, fmt.Sprintf("unable to register checkers: %s", errAddCheckFail.Error()))
				So(svcList.HealthCheck, ShouldBeTrue)
				So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 1)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Mongo DB")
			})
			Reset(func() {
				// This reset is run after each `Convey` at the same scope (indentation)
			})
		})

		Convey("Given that all dependencies are successfully initialised but the http server fails", func() {
			// setup (run before each `Convey` at this scope / indentation):
			initMock := &mock.InitialiserMock{
				DoGetHealthCheckFunc:             funcDoGetHealthcheckOk,
				DoGetHTTPServerFunc:              funcDoGetFailingHTTPServer,
				DoGetMigratorFunc:                funcDoGetMigrator,
				DoGetMongoDBFunc:                 funcDoGetMongoDBOk,
				DoGetAppClientsFunc:              funcDoGetAppClientsOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			serverWg.Add(1)

			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			So(err, ShouldBeNil)

			Convey("Then the error is returned in the error channel", func() {
				sErr := <-svcErrors
				So(sErr.Error(), ShouldResemble, fmt.Sprintf("failure in http listen and serve: %s", errServer.Error()))
				So(len(failingServerMock.ListenAndServeCalls()), ShouldEqual, 1)
			})

			Reset(func() {
				// This reset is run after each `Convey` at the same scope (indentation)
			})
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Having a correctly initialised service", t, func() {
		cfg, cfgErr := config.Get()
		So(cfgErr, ShouldBeNil)

		hcStopped := false
		serverStopped := false

		authorisationMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &mock.HealthCheckerMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &mock.HTTPServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				serverStopped = true
				return nil
			},
		}

		funcClose := func(context.Context) error {
			if !hcStopped {
				return errors.New("Dependency was closed before healthcheck")
			}
			if !serverStopped {
				return errors.New("Dependency was closed before http server")
			}
			return nil
		}

		funcDoGetMigrator := func(ctx context.Context, jobService application.JobService, clientList *clients.ClientList) (migrator.Migrator, error) {
			return &migratorMock.MigratorMock{
				ShutdownFunc: funcClose,
			}, nil
		}

		funcDoGetMongoDBOk := func(context.Context, config.MongoConfig) (store.MongoDB, error) {
			return &storeMock.MongoDBMock{
				CloseFunc: funcClose,
			}, nil
		}

		funcDoGetAppClientsOk := func(context.Context, *config.Config) *clients.ClientList {
			return &clients.ClientList{}
		}

		funcDoGetAuthOk := func(ctx context.Context, authorisationConfig *authorisation.Config) (authorisation.Middleware, error) {
			return authorisationMiddleware, nil
		}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {
			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return serverMock },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetMigratorFunc:                funcDoGetMigrator,
				DoGetMongoDBFunc:                 funcDoGetMongoDBOk,
				DoGetAppClientsFunc:              funcDoGetAppClientsOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)

			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {
			failingServerMock := &mock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.HTTPServer { return failingServerMock },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.HealthChecker, error) {
					return hcMock, nil
				},
				DoGetMigratorFunc:                funcDoGetMigrator,
				DoGetMongoDBFunc:                 funcDoGetMongoDBOk,
				DoGetAppClientsFunc:              funcDoGetAppClientsOk,
				DoGetAuthorisationMiddlewareFunc: funcDoGetAuthOk,
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)

			svc := service.New(cfg, svcList)
			err := svc.Run(ctx, testBuildTime, testGitCommit, testVersion, svcErrors)

			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "failed to shutdown gracefully")
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(failingServerMock.ShutdownCalls()), ShouldEqual, 1)
		})

		Convey("If service times out while shutting down, the Close operation fails with the expected error", func() {
			cfg.GracefulShutdownTimeout = 1 * time.Millisecond
			timeoutServerMock := &mock.HTTPServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					time.Sleep(2 * time.Millisecond)
					return nil
				},
			}

			svcList := service.NewServiceList(nil)
			svcList.HealthCheck = true
			svc := service.Service{
				Config:      cfg,
				ServiceList: svcList,
				Server:      timeoutServerMock,
				HealthCheck: hcMock,
			}

			err := svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldResemble, "context deadline exceeded")
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(timeoutServerMock.ShutdownCalls()), ShouldEqual, 1)
		})
	})
}
