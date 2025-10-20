package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	apiErrors "github.com/ONSdigital/dis-migration-service/api/errors"
	"github.com/ONSdigital/dis-migration-service/domain"
	storeErrors "github.com/ONSdigital/dis-migration-service/store/errors"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

func (api *MigrationAPI) getJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars[PathParameterJobID]

	job, err := api.Store.GetJob(ctx, jobID)
	if err != nil {
		if err == storeErrors.ErrJobNotFound {
			handleError(ctx, w, r, apiErrors.ErrJobNotFound)
		} else {
			log.Error(ctx, "failed to get job", err)
			handleError(ctx, w, r, apiErrors.ErrInternalServerError)
		}
		return
	}

	bytes, err := json.Marshal(job)
	if err != nil {
		log.Error(ctx, "failed to marshal response", err)
		handleError(ctx, w, r, apiErrors.ErrInternalServerError)
		return
	}

	handleSuccess(ctx, w, r, http.StatusOK, bytes)
}

func (api *MigrationAPI) createJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	jobConfigBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Info(ctx, "unable to read body")
		handleError(ctx, w, r, apiErrors.ErrUnableToParseBody)
		return
	}

	var jobConfig *domain.JobConfig

	err = json.Unmarshal(jobConfigBytes, &jobConfig)
	if err != nil {
		log.Info(ctx, "failed to unmarshal job config")
		handleError(ctx, w, r, apiErrors.ErrUnableToParseBody)
		return
	}

	errs := validateJobConfig(jobConfig)
	if errs != nil {
		log.Info(ctx, "failed to validate job config")
		handleError(ctx, w, r, errs...)
		return
	}

	job := domain.NewJob(jobConfig)

	storedJob, err := api.Store.CreateJob(ctx, &job)
	if err != nil {
		log.Error(ctx, "failed to create job", err)
		handleError(ctx, w, r, apiErrors.ErrInternalServerError)
		return
	}

	bytes, err := json.Marshal(storedJob)
	if err != nil {
		log.Error(ctx, "failed to marshal response", err)
		handleError(ctx, w, r, apiErrors.ErrInternalServerError)
		return
	}

	handleSuccess(ctx, w, r, http.StatusAccepted, bytes)

	api.Migrator.Migrate(context.Background(), storedJob)
}
