package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	domainMocks "github.com/ONSdigital/dis-migration-service/domain/mock"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/store"
	storeMocks "github.com/ONSdigital/dis-migration-service/store/mock"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	testHost  = "http://localhost:8080"
	testJobID = "test-job-id"
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

func TestGetJob(t *testing.T) {
	Convey("Given a job service and a store that has a job for the requested id", t, func() {
		expectedJob := &domain.Job{
			ID: "job-1",
		}

		mockMongo := &storeMocks.MongoDBMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return expectedJob, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()

		Convey("When GetJob is called", func() {
			job, err := jobService.GetJob(ctx, expectedJob.ID)

			Convey("Then the store GetJob should be called and the job returned", func() {
				So(len(mockMongo.GetJobCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobCalls()[0].JobID, ShouldEqual, expectedJob.ID)

				So(err, ShouldBeNil)
				So(job, ShouldResemble, expectedJob)
			})
		})
	})

	Convey("Given a job service and a store that returns not found", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return nil, appErrors.ErrJobNotFound
			},
		}

		missingID := "missing-job"

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()

		Convey("When GetJob is called for a missing job", func() {
			job, err := jobService.GetJob(ctx, missingID)

			Convey("Then nil job and ErrJobNotFound should be returned", func() {
				So(len(mockMongo.GetJobCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobCalls()[0].JobID, ShouldEqual, missingID)

				So(job, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, appErrors.ErrJobNotFound)
			})
		})
	})

	Convey("Given a job service and a store that returns an internal error", t, func() {
		fakeErr := errors.New("fake store error")

		mockMongo := &storeMocks.MongoDBMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return nil, fakeErr
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()
		jobID := "job-error"

		Convey("When GetJob is called and the store errors", func() {
			job, err := jobService.GetJob(ctx, jobID)

			Convey("Then nil job and the underlying error should be returned", func() {
				So(len(mockMongo.GetJobCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobCalls()[0].JobID, ShouldEqual, jobID)

				So(job, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, fakeErr.Error())
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

func TestGetJobTasks(t *testing.T) {
	Convey("Given a job service and store that has stored tasks for a job", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobTasksFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{
					{
						ID:          "task1",
						JobID:       testJobID,
						LastUpdated: time.Now().UTC(),
						Source: &domain.TaskMetadata{
							ID:    "source-id-1",
							Label: "Source Dataset 1",
							URI:   "/data/source1",
						},
						Target: &domain.TaskMetadata{
							ID:    "target-id-1",
							Label: "Target Dataset 1",
							URI:   "/data/target1",
						},
						State: domain.JobStateMigrating,
						Type:  domain.MigrationTaskTypeDataset,
						Links: domain.TaskLinks{
							Self: &domain.LinkObject{HRef: "http://localhost:8080/v1/migration-jobs/test-job-id/tasks/task1"},
							Job:  &domain.LinkObject{HRef: "http://localhost:8080/v1/migration-jobs/test-job-id"},
						},
					},
					{
						ID:          "task2",
						JobID:       testJobID,
						LastUpdated: time.Now().UTC(),
						Source: &domain.TaskMetadata{
							ID:    "source-id-2",
							Label: "Source Dataset 2",
							URI:   "/data/source2",
						},
						Target: &domain.TaskMetadata{
							ID:    "target-id-2",
							Label: "Target Dataset 2",
							URI:   "/data/target2",
						},
						State: domain.JobStatePublishing,
						Type:  domain.MigrationTaskTypeDatasetEdition,
						Links: domain.TaskLinks{
							Self: &domain.LinkObject{HRef: "http://localhost:8080/v1/migration-jobs/test-job-id/tasks/task2"},
							Job:  &domain.LinkObject{HRef: "http://localhost:8080/v1/migration-jobs/test-job-id"},
						},
					},
				}, 2, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()
		jobID := testJobID

		Convey("When job tasks are retrieved", func() {
			tasks, totalCount, err := jobService.GetJobTasks(ctx, jobID, 10, 0)

			Convey("Then the store should be called to get job tasks", func() {
				So(len(mockMongo.GetJobTasksCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobTasksCalls()[0].JobID, ShouldEqual, jobID)
				So(mockMongo.GetJobTasksCalls()[0].Limit, ShouldEqual, 10)
				So(mockMongo.GetJobTasksCalls()[0].Offset, ShouldEqual, 0)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And the tasks should be returned", func() {
						So(len(tasks), ShouldEqual, 2)
						So(totalCount, ShouldEqual, 2)
						So(tasks[0].ID, ShouldEqual, "task1")
						So(tasks[0].Source.ID, ShouldEqual, "source-id-1")
						So(tasks[0].Target.ID, ShouldEqual, "target-id-1")
						So(tasks[0].State, ShouldEqual, domain.JobStateMigrating)
						So(tasks[0].Type, ShouldEqual, domain.MigrationTaskTypeDataset)
						So(tasks[1].ID, ShouldEqual, "task2")
						So(tasks[1].State, ShouldEqual, domain.JobStatePublishing)
						So(tasks[1].Type, ShouldEqual, domain.MigrationTaskTypeDatasetEdition)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that has no tasks for a job", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobTasksFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{}, 0, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()
		jobID := testJobID

		Convey("When job tasks are retrieved", func() {
			tasks, totalCount, err := jobService.GetJobTasks(ctx, jobID, 10, 0)

			Convey("Then the store should be called to get job tasks", func() {
				So(len(mockMongo.GetJobTasksCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobTasksCalls()[0].JobID, ShouldEqual, jobID)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And an empty task list should be returned", func() {
						So(len(tasks), ShouldEqual, 0)
						So(totalCount, ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that returns an error when getting job tasks", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobTasksFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Task, int, error) {
				return nil, 0, errors.New("fake error for testing")
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()
		jobID := testJobID

		Convey("When job tasks are retrieved", func() {
			tasks, totalCount, err := jobService.GetJobTasks(ctx, jobID, 10, 0)

			Convey("Then the store should be called to get job tasks", func() {
				So(len(mockMongo.GetJobTasksCalls()), ShouldEqual, 1)

				Convey("Then an error should be returned", func() {
					So(tasks, ShouldBeNil)
					So(totalCount, ShouldEqual, 0)
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}

func TestCountTasksByJobID(t *testing.T) {
	Convey("Given a job service and store that has tasks for a job", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			CountTasksByJobIDFunc: func(ctx context.Context, jobID string) (int, error) {
				return 5, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()
		jobID := testJobID

		Convey("When task count is retrieved", func() {
			count, err := jobService.CountTasksByJobID(ctx, jobID)

			Convey("Then the store should be called to count tasks", func() {
				So(len(mockMongo.CountTasksByJobIDCalls()), ShouldEqual, 1)
				So(mockMongo.CountTasksByJobIDCalls()[0].JobID, ShouldEqual, jobID)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And the task count should be returned", func() {
						So(count, ShouldEqual, 5)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that has no tasks for a job", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			CountTasksByJobIDFunc: func(ctx context.Context, jobID string) (int, error) {
				return 0, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()
		jobID := testJobID

		Convey("When task count is retrieved", func() {
			count, err := jobService.CountTasksByJobID(ctx, jobID)

			Convey("Then the store should be called to count tasks", func() {
				So(len(mockMongo.CountTasksByJobIDCalls()), ShouldEqual, 1)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And zero count should be returned", func() {
						So(count, ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that returns an error when counting tasks", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			CountTasksByJobIDFunc: func(ctx context.Context, jobID string) (int, error) {
				return 0, errors.New("fake error for testing")
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients, testHost)

		ctx := context.Background()
		jobID := testJobID

		Convey("When task count is retrieved", func() {
			count, err := jobService.CountTasksByJobID(ctx, jobID)

			Convey("Then the store should be called to count tasks", func() {
				So(len(mockMongo.CountTasksByJobIDCalls()), ShouldEqual, 1)

				Convey("Then an error should be returned", func() {
					So(count, ShouldEqual, 0)
					So(err, ShouldNotBeNil)
				})
			})
		})
	})
}
