package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

func (api *MigrationAPI) getJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars[PathParameterJobID]

	job, err := api.JobService.GetJob(ctx, jobID)
	if err != nil {
		if !errors.Is(err, appErrors.ErrJobNotFound) {
			log.Error(ctx, "failed to get job", err)
		}
		handleError(ctx, w, r, err)
		return
	}

	bytes, err := json.Marshal(job)
	if err != nil {
		log.Error(ctx, "failed to marshal response", err)
		handleError(ctx, w, r, err)
		return
	}

	handleSuccess(ctx, w, r, http.StatusOK, bytes)
}

func (api *MigrationAPI) createJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	jobConfigBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Info(ctx, "unable to read body")
		handleError(ctx, w, r, appErrors.ErrUnableToParseBody)
		return
	}

	var jobConfig *domain.JobConfig

	err = json.Unmarshal(jobConfigBytes, &jobConfig)
	if err != nil {
		log.Info(ctx, "failed to unmarshal job config")
		handleError(ctx, w, r, appErrors.ErrUnableToParseBody)
		return
	}

	errs := jobConfig.ValidateInternal()
	if errs != nil {
		log.Info(ctx, "failed to validate job config")
		handleError(ctx, w, r, errs...)
		return
	}

	job, err := api.JobService.CreateJob(ctx, jobConfig)
	if err != nil {
		handleError(ctx, w, r, err)
		return
	}

	bytes, err := json.Marshal(job)
	if err != nil {
		log.Error(ctx, "failed to marshal response", err)
		handleError(ctx, w, r, err)
		return
	}

	handleSuccess(ctx, w, r, http.StatusAccepted, bytes)

	api.Migrator.Migrate(context.Background(), job)
}
