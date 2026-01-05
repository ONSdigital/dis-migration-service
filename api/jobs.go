package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/statemachine"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// StateChangeRequest represents the payload used to request a job state change.
type StateChangeRequest struct {
	State domain.State `json:"state"`
}

// getJob is an implementation for retrieving a migration job.
func (api *MigrationAPI) getJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobNumberStr := vars[PathParameterJobNumber]
	jobNumber, err := strconv.Atoi(jobNumberStr)
	if err != nil {
		log.Info(ctx, "failed to get job - job number must be an int")
		handleError(ctx, w, r, appErrors.ErrJobNumberMustBeInt)
		return
	}

	job, err := api.JobService.GetJob(ctx, jobNumber)
	if err != nil {
		if !errors.Is(err, appErrors.ErrJobNotFound) {
			log.Error(ctx, "failed to get job with job number: "+strconv.Itoa(jobNumber), err)
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
	statesParam := r.URL.Query()["state"] // supports ?state=a&state=b and ?state=a,b
	states := make([]domain.State, 0, len(statesParam))

	for _, s := range statesParam {
		for _, p := range strings.Split(s, ",") {
			state := domain.State(strings.TrimSpace(p))
			if !domain.IsValidState(state) {
				return nil, 0, appErrors.ErrJobStateInvalid
			}
			states = append(states, state)
		}
	}
	return api.JobService.GetJobs(r.Context(), states, limit, offset)
}

// getJobTasks is an implementation of PaginatedHandler for retrieving
// job tasks.
func (api *MigrationAPI) getJobTasks(w http.ResponseWriter, r *http.Request, limit, offset int) (items interface{}, totalCount int, err error) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobNumberStr := vars[PathParameterJobNumber]

	if jobNumberStr == "" {
		return nil, 0, appErrors.ErrJobNumberNotProvided
	}
	jobNumber, err := strconv.Atoi(jobNumberStr)
	if err != nil {
		log.Error(ctx, "failed to get job -  job number must be an int", err)
		handleError(ctx, w, r, err)
		return &[]domain.Task{}, 0, err
	}

	// Ensure job exists -> return 404 if not found
	if _, err := api.JobService.GetJob(ctx, jobNumber); err != nil {
		return nil, 0, err // This will return 404 if job not found
	}

	statesParam := r.URL.Query()["state"]
	states := make([]domain.State, 0, len(statesParam))

	for _, s := range statesParam {
		for _, p := range strings.Split(s, ",") {
			state := domain.State(strings.TrimSpace(p))
			if !domain.IsValidState(state) {
				return nil, 0, appErrors.ErrTaskStateInvalid
			}
			states = append(states, state)
		}
	}

	// Fetch tasks for the job
	tasks, totalCount, err := api.JobService.GetJobTasks(ctx, states, jobNumber, limit, offset)
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
}

// getJobEvents is an implementation for retrieving job events.
func (api *MigrationAPI) getJobEvents(w http.ResponseWriter, r *http.Request, limit, offset int) (items interface{}, totalCount int, err error) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobNumberStr := vars[PathParameterJobNumber]

	if jobNumberStr == "" {
		return nil, 0, appErrors.ErrJobNumberNotProvided
	}

	jobNumber, err := strconv.Atoi(jobNumberStr)
	if err != nil {
		log.Error(ctx, "failed to get job -  job number must be an int", err)
		handleError(ctx, w, r, err)
		return
	}

	// Ensure job exists -> return 404 if not found
	if _, err := api.JobService.GetJob(ctx, jobNumber); err != nil {
		return nil, 0, err
	}

	// Fetch events for the job
	events, totalCount, err := api.JobService.GetJobEvents(ctx, jobNumber, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return events, totalCount, nil
}

// updateJobState handles requests to update the state of a migration job.
func (api *MigrationAPI) updateJobState(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	vars := mux.Vars(r)
	jobNumberStr := vars[PathParameterJobNumber]

	if jobNumberStr == "" {
		handleError(ctx, w, r, appErrors.ErrJobNumberNotProvided)
	}

	jobNumber, err := strconv.Atoi(jobNumberStr)
	if err != nil {
		log.Error(ctx, "failed to get job -  job number must be an int", err)
		handleError(ctx, w, r, err)
		return
	}

	// Base log context reused throughout the handler
	logData := log.Data{
		"job_number": jobNumber,
	}

	// Read request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Info(ctx, "unable to read request body", logData)
		handleError(ctx, w, r, appErrors.ErrUnableToParseBody)
		return
	}

	var req StateChangeRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Info(ctx, "failed to decode request body", logData)
		handleError(ctx, w, r, appErrors.ErrUnableToParseBody)
		return
	}

	// Add requested state to log context once parsed
	logData["requested_state"] = req.State

	// Validate state is known
	if !domain.IsValidState(req.State) {
		log.Info(ctx, "invalid state provided", logData)
		handleError(ctx, w, r, appErrors.ErrJobStateInvalid)
		return
	}

	// Validate state is allowed for this endpoint
	if !statemachine.IsAllowedStateForJobUpdate(req.State) {
		log.Info(ctx, "state not allowed for this endpoint", logData)
		handleError(ctx, w, r, appErrors.ErrJobStateNotAllowed)
		return
	}

	// Get current job
	job, err := api.JobService.GetJob(ctx, jobNumber)
	if err != nil {
		if !errors.Is(err, appErrors.ErrJobNotFound) {
			log.Error(ctx, "failed to get job", err, logData)
		}
		handleError(ctx, w, r, err)
		return
	}

	// Add current state once job is known
	logData["current_state"] = job.State

	// Attempt state transition
	err = api.JobService.UpdateJobState(ctx, jobNumber, req.State)
	if err != nil {
		var transitionErr *statemachine.TransitionError
		if errors.As(err, &transitionErr) {
			log.Info(ctx, "invalid state transition", logData)
			handleError(ctx, w, r, appErrors.ErrJobStateTransitionNotAllowed)
			return
		}

		log.Error(ctx, "failed to update job state", err, logData)
		handleError(ctx, w, r, err)
		return
	}

	log.Info(ctx, "job state updated successfully", logData)
	w.WriteHeader(http.StatusNoContent)
}
