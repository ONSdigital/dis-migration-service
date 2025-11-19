package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/config"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/migrator"
	auth "github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// MigrationAPI provides a struct to wrap the api around
type MigrationAPI struct {
	JobService application.JobService
	Migrator   migrator.Migrator
	Paginator  *Paginator
	Router     *mux.Router
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, cfg *config.Config, router *mux.Router, jobService application.JobService, dataMigrator migrator.Migrator, authMiddleware auth.Middleware) *MigrationAPI {
	paginator := NewPaginator(cfg.DefaultLimit, cfg.DefaultOffset, cfg.DefaultMaxLimit)

	api := &MigrationAPI{
		Migrator:   dataMigrator,
		Router:     router,
		JobService: jobService,
		Paginator:  paginator,
	}

	api.get("/v1/migration-jobs",
		paginator.Paginate(api.getJobs),
	)

	api.post(
		"/v1/migration-jobs",
		authMiddleware.Require("migrations:create", api.createJob),
	)

	api.get(
		fmt.Sprintf("/v1/migration-jobs/{%s}", PathParameterJobID),
		authMiddleware.Require("migrations:read", api.getJob),
	)

	return api
}

// get registers a GET http.HandlerFunc.
func (api *MigrationAPI) get(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodGet)
}

// post registers a POST http.HandlerFunc.
func (api *MigrationAPI) post(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodPost)
}

// put registers a PUT http.HandlerFunc.
// nolint:unused // putting in now for completeness.
func (api *MigrationAPI) put(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodPut)
}

// delete registers a DELETE http.HandlerFunc.
// nolint:unused // putting in now for completeness.
func (api *MigrationAPI) delete(path string, handler http.HandlerFunc) {
	api.Router.HandleFunc(path, handler).Methods(http.MethodDelete)
}

// handleError deals with all errors from the API layer
func handleError(ctx context.Context, w http.ResponseWriter, r *http.Request, errors ...error) {
	var errList appErrors.ErrorList
	var statusCode = 0

	for i := range errors {
		redactedErr := appErrors.New(errors[i])

		if redactedErr.Code > statusCode {
			statusCode = redactedErr.Code
		}

		if redactedErr.Code > 499 {
			log.Error(ctx, "handling server error", errors[i])
		}

		errList.Errors = append(errList.Errors, redactedErr)
	}

	responseBody, err := json.Marshal(errList)
	if err != nil {
		log.Error(ctx, "failed to encode error list", err)
		http.Error(w, appErrors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(responseBody); err != nil {
		log.Error(ctx, "failed to write response body", err)
		http.Error(w, appErrors.ErrInternalServerError.Error(), http.StatusInternalServerError)
	}
}

// handleSuccess sets the status code and body for a response
func handleSuccess(ctx context.Context, w http.ResponseWriter, r *http.Request, statusCode int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if body != nil {
		if _, err := w.Write(body); err != nil {
			log.Error(ctx, "failed to encode response body", err)
			handleError(ctx, w, r, appErrors.ErrInternalServerError)
			return
		}
	}
}
