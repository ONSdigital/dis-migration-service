package application

import (
	"context"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"

	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/store"
	"github.com/ONSdigital/log.go/v2/log"
)

// JobService defines the contract for job-related operations
//
//go:generate moq -out mock/jobservice.go -pkg mock . JobService
type JobService interface {
	CreateJob(ctx context.Context, jobConfig *domain.JobConfig) (*domain.Job, error)
	GetJob(ctx context.Context, jobID string) (*domain.Job, error)
}

type jobService struct {
	store   *store.Datastore
	host    string
	clients *clients.ClientList
}

func Setup(datastore *store.Datastore, appClients *clients.ClientList, host string) JobService {
	return &jobService{
		store:   datastore,
		host:    host,
		clients: appClients,
	}
}

func (js *jobService) CreateJob(ctx context.Context, jobConfig *domain.JobConfig) (*domain.Job, error) {
	err := jobConfig.ValidateExternal(ctx, *js.clients)
	if err != nil {
		return &domain.Job{}, err
	}

	job := domain.NewJob(jobConfig, js.host)

	foundJobs, err := js.store.GetJobsByConfigAndState(ctx, job.Config, domain.GetNonCancelledStates(), 0, 1)
	if err != nil {
		log.Error(ctx, "failed to validate job creation", err)
		return &domain.Job{}, appErrors.ErrInternalServerError
	}

	if len(foundJobs) > 0 {
		log.Error(ctx, "found running jobs with this config", err, log.Data{
			"job_config": job.Config,
		})
		return &domain.Job{}, appErrors.ErrJobAlreadyRunning
	}

	err = js.store.CreateJob(ctx, &job)
	if err != nil {
		log.Error(ctx, "failed to create job", err)
		return &domain.Job{}, appErrors.ErrInternalServerError
	}
	return &job, nil
}

func (js *jobService) GetJob(ctx context.Context, jobID string) (*domain.Job, error) {
	return js.store.GetJob(ctx, jobID)
}
