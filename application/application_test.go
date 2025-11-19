package application

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	domainMocks "github.com/ONSdigital/dis-migration-service/domain/mock"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/store"
	storeMocks "github.com/ONSdigital/dis-migration-service/store/mock"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	testHost = "http://localhost:8080"
)

func TestCreateJob(t *testing.T) {
	Convey("Given a job service and store that has no stored jobs and a valid job config", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobsByConfigAndStateFunc: func(ctx context.Context, jc *domain.JobConfig, states []domain.JobState, limit, offset int) ([]*domain.Job, error) {
				return nil, nil
			},
			CreateJobFunc: func(ctx context.Context, job *domain.Job) error {
				return nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockValidator := &domainMocks.JobValidatorMock{
			ValidateSourceIDWithExternalFunc: func(ctx context.Context, sourceID string, appClients *clients.ClientList) error {
				return nil
			},
			ValidateTargetIDWithExternalFunc: func(ctx context.Context, targetID string, appClients *clients.ClientList) error {
				return nil
			},
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)
		jobConfig := domain.JobConfig{
			SourceID:  "/source-id",
			TargetID:  "target-id",
			Type:      domain.JobTypeStaticDataset,
			Validator: mockValidator,
		}

		ctx := context.Background()

		Convey("When a job is created", func() {
			job, err := jobService.CreateJob(ctx, &jobConfig)

			Convey("Then the store should be checked for matching jobs", func() {
				So(len(mockMongo.GetJobsByConfigAndStateCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobsByConfigAndStateCalls()[0].Jc.SourceID, ShouldEqual, jobConfig.SourceID)
				So(mockMongo.GetJobsByConfigAndStateCalls()[0].Jc.TargetID, ShouldEqual, jobConfig.TargetID)
				So(mockMongo.GetJobsByConfigAndStateCalls()[0].Jc.Type, ShouldEqual, jobConfig.Type)
				So(mockMongo.GetJobsByConfigAndStateCalls()[0].Limit, ShouldEqual, 1)
				So(mockMongo.GetJobsByConfigAndStateCalls()[0].Offset, ShouldEqual, 0)
				So(mockMongo.GetJobsByConfigAndStateCalls()[0].States, ShouldEqual, domain.GetNonCancelledStates())

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And a job should be created in the store", func() {
						So(job, ShouldNotEqual, &domain.Job{})
						So(len(mockMongo.CreateJobCalls()), ShouldEqual, 1)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that has a matching stored job and a valid job config", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobsByConfigAndStateFunc: func(ctx context.Context, jc *domain.JobConfig, states []domain.JobState, limit, offset int) ([]*domain.Job, error) {
				return []*domain.Job{
					{
						Config: jc,
						State:  domain.JobStateSubmitted,
					},
				}, nil
			},
			CreateJobFunc: func(ctx context.Context, job *domain.Job) error {
				return nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockValidator := &domainMocks.JobValidatorMock{
			ValidateSourceIDWithExternalFunc: func(ctx context.Context, sourceID string, appClients *clients.ClientList) error {
				return nil
			},
			ValidateTargetIDWithExternalFunc: func(ctx context.Context, targetID string, appClients *clients.ClientList) error {
				return nil
			},
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)
		jobConfig := domain.JobConfig{
			SourceID:  "/source-id",
			TargetID:  "target-id",
			Type:      domain.JobTypeStaticDataset,
			Validator: mockValidator,
		}

		ctx := context.Background()

		Convey("When a job is created", func() {
			job, err := jobService.CreateJob(ctx, &jobConfig)

			Convey("Then the store should be checked for matching jobs", func() {
				So(len(mockMongo.GetJobsByConfigAndStateCalls()), ShouldEqual, 1)

				Convey("Then an error should be returned", func() {
					So(job, ShouldEqual, &domain.Job{})
					So(err, ShouldNotBeNil)
					So(err, ShouldEqual, appErrors.ErrJobAlreadyRunning)

					Convey("And the store should not be called to create a job", func() {
						So(len(mockMongo.CreateJobCalls()), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that returns an error when checking jobs and a valid job config", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobsByConfigAndStateFunc: func(ctx context.Context, jc *domain.JobConfig, states []domain.JobState, limit, offset int) ([]*domain.Job, error) {
				return nil, errors.New("fake error for testing")
			},
			CreateJobFunc: func(ctx context.Context, job *domain.Job) error {
				return nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		host := "http://localhost:8080"

		mockValidator := &domainMocks.JobValidatorMock{
			ValidateSourceIDWithExternalFunc: func(ctx context.Context, sourceID string, appClients *clients.ClientList) error {
				return nil
			},
			ValidateTargetIDWithExternalFunc: func(ctx context.Context, targetID string, appClients *clients.ClientList) error {
				return nil
			},
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, host)
		jobConfig := domain.JobConfig{
			SourceID:  "/source-id",
			TargetID:  "target-id",
			Type:      domain.JobTypeStaticDataset,
			Validator: mockValidator,
		}

		ctx := context.Background()

		Convey("When a job is created", func() {
			job, err := jobService.CreateJob(ctx, &jobConfig)

			Convey("Then the store should be checked for matching jobs", func() {
				So(len(mockMongo.GetJobsByConfigAndStateCalls()), ShouldEqual, 1)

				Convey("Then an error should be returned", func() {
					So(job, ShouldEqual, &domain.Job{})
					So(err, ShouldNotBeNil)
					So(err, ShouldEqual, appErrors.ErrInternalServerError)

					Convey("And the store should not be called to create a job", func() {
						So(len(mockMongo.CreateJobCalls()), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that returns an error when creating a job and a valid job config", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobsByConfigAndStateFunc: func(ctx context.Context, jc *domain.JobConfig, states []domain.JobState, limit, offset int) ([]*domain.Job, error) {
				return nil, nil
			},
			CreateJobFunc: func(ctx context.Context, job *domain.Job) error {
				return errors.New("fake error for testing")
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockValidator := &domainMocks.JobValidatorMock{
			ValidateSourceIDWithExternalFunc: func(ctx context.Context, sourceID string, appClients *clients.ClientList) error {
				return nil
			},
			ValidateTargetIDWithExternalFunc: func(ctx context.Context, targetID string, appClients *clients.ClientList) error {
				return nil
			},
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)
		jobConfig := domain.JobConfig{
			SourceID:  "/source-id",
			TargetID:  "target-id",
			Type:      domain.JobTypeStaticDataset,
			Validator: mockValidator,
		}

		ctx := context.Background()

		Convey("When a job is created", func() {
			job, err := jobService.CreateJob(ctx, &jobConfig)

			Convey("Then the store should be called to create a job", func() {
				So(len(mockMongo.GetJobsByConfigAndStateCalls()), ShouldEqual, 1)
				So(len(mockMongo.CreateJobCalls()), ShouldEqual, 1)

				Convey("Then an error should be returned", func() {
					So(job, ShouldEqual, &domain.Job{})
					So(err, ShouldNotBeNil)
					So(err, ShouldEqual, appErrors.ErrInternalServerError)
				})
			})
		})
	})
}

func TestGetJobs(t *testing.T) {
	Convey("Given a job service and store that has stored jobs", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobsFunc: func(ctx context.Context, limit, offset int) ([]*domain.Job, int, error) {
				return []*domain.Job{
					{ID: "job1"},
					{ID: "job2"},
				}, 2, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()

		Convey("When jobs are retrieved", func() {
			jobs, totalCount, err := jobService.GetJobs(ctx, 10, 0)

			Convey("Then the store should be called to get jobs", func() {
				So(len(mockMongo.GetJobsCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobsCalls()[0].Limit, ShouldEqual, 10)
				So(mockMongo.GetJobsCalls()[0].Offset, ShouldEqual, 0)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And the jobs should be returned", func() {
						So(len(jobs), ShouldEqual, 2)
						So(totalCount, ShouldEqual, 2)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that returns an error when getting jobs", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobsFunc: func(ctx context.Context, limit, offset int) ([]*domain.Job, int, error) {
				return nil, 0, errors.New("fake error for testing")
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()

		Convey("When jobs are retrieved", func() {
			jobs, totalCount, err := jobService.GetJobs(ctx, 10, 0)

			Convey("Then the store should be called to get jobs", func() {
				So(len(mockMongo.GetJobsCalls()), ShouldEqual, 1)

				Convey("Then an error should be returned", func() {
					So(jobs, ShouldBeNil)
					So(totalCount, ShouldEqual, 0)
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}
