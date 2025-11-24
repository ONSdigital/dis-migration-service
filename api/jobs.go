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

// getJobs is an implementation of PaginatedHandler for retrieving jobs.
func (api *MigrationAPI) getJobs(w http.ResponseWriter, r *http.Request, limit, offset int) (items interface{}, totalCount int, err error) {
	return api.JobService.GetJobs(r.Context(), limit, offset)
}

// getJobTasks is an implementation of PaginatedHandler for retrieving
// job tasks.
func (api *MigrationAPI) getJobTasks(w http.ResponseWriter, r *http.Request, limit, offset int) (items interface{}, totalCount int, err error) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobID := vars[PathParameterJobID]

	if jobID == "" {
		return nil, 0, appErrors.ErrJobIDNotProvided
	}

	// Ensure job exists -> return 404 if not found
	if _, err := api.JobService.GetJob(ctx, jobID); err != nil {
		return nil, 0, err // This will return 404 if job not found
	}

	// Fetch tasks for the job
	tasks, totalCount, err := api.JobService.GetJobTasks(ctx, jobID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return tasks, totalCount, nil
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
