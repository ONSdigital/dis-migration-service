package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	apiErrors "github.com/ONSdigital/dis-migration-service/api/errors"
	"github.com/ONSdigital/dis-migration-service/migrator"
	"github.com/ONSdigital/dis-migration-service/store"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// MigrationAPI provides a struct to wrap the api around
type MigrationAPI struct {
	Router   *mux.Router
	Store    *store.Datastore
	Migrator migrator.Migrator
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, router *mux.Router, dataStore *store.Datastore, dataMigrator migrator.Migrator) *MigrationAPI {
	api := &MigrationAPI{
		Router:   router,
		Store:    dataStore,
		Migrator: dataMigrator,
	}

	api.post(
		"/v1/migration-jobs",
		api.createJob,
	)

	api.get(
		fmt.Sprintf("/v1/migration-jobs/{%s}", PathParameterJobID),
		api.getJob,
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
func handleError(ctx context.Context, w http.ResponseWriter, r *http.Request, errors ...apiErrors.APIError) {
	var errList apiErrors.ErrorList
	var statusCode = 0

	for i := range errors {
		if errors[i].Code > statusCode {
			statusCode = errors[i].Code
		}
		redactedError := apiErrors.Error{
			Code:        errors[i].Code,
			Description: errors[i].Error(),
		}
		errList.Errors = append(errList.Errors, redactedError)
	}

	responseBody, err := json.Marshal(errList)
	if err != nil {
		log.Error(ctx, "failed to encode error list", err)
		http.Error(w, apiErrors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err := w.Write(responseBody); err != nil {
		log.Error(ctx, "failed to write response body", err)
		http.Error(w, apiErrors.ErrInternalServerError.Error(), http.StatusInternalServerError)
	}
}

// handleSuccess sets the status code and body for a response
func handleSuccess(ctx context.Context, w http.ResponseWriter, r *http.Request, statusCode int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if body != nil {
		if _, err := w.Write(body); err != nil {
			log.Error(ctx, "failed to encode response body", err)
			handleError(ctx, w, r, apiErrors.ErrInternalServerError)
			return
		}
	}
}
