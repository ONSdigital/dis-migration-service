package application

import (
	"context"
	"errors"
	"fmt"
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
	testJobID        = "test-job-id"
	nonExistentJobID = "non-existent-job-id"
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

		jobService := Setup(&mockStore, &mockClients)
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

		jobService := Setup(&mockStore, &mockClients)
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

		mockValidator := &domainMocks.JobValidatorMock{
			ValidateSourceIDWithExternalFunc: func(ctx context.Context, sourceID string, appClients *clients.ClientList) error {
				return nil
			},
			ValidateTargetIDWithExternalFunc: func(ctx context.Context, targetID string, appClients *clients.ClientList) error {
				return nil
			},
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)
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

		jobService := Setup(&mockStore, &mockClients)
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

		jobService := Setup(&mockStore, &mockClients)

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

		jobService := Setup(&mockStore, &mockClients)

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

		jobService := Setup(&mockStore, &mockClients)

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

		jobService := Setup(&mockStore, &mockClients)

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

		jobService := Setup(&mockStore, &mockClients)

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

func TestCreateTask(t *testing.T) {
	Convey("Given a job service and store that has a stored job in submitted state", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return &domain.Job{
					ID:    jobID,
					State: domain.JobStateSubmitted,
				}, nil
			},
			CreateTaskFunc: func(ctx context.Context, task *domain.Task) error {
				return nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When a task is created", func() {
			task := &domain.Task{
				ID:    "task-123",
				Type:  domain.TaskTypeDataset,
				State: domain.JobStateSubmitted,
				Source: &domain.TaskMetadata{
					ID:    "source-1",
					Label: "Source Dataset",
					URI:   "/data/source",
				},
				Target: &domain.TaskMetadata{
					ID:    "target-1",
					Label: "Target Dataset",
					URI:   "/data/target",
				},
			}

			createdTask, err := jobService.CreateTask(ctx, jobID, task)

			Convey("Then the store should be called to get the job", func() {
				So(len(mockMongo.GetJobCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobCalls()[0].JobID, ShouldEqual, jobID)

				Convey("Then the store should be called to create the task", func() {
					So(len(mockMongo.CreateTaskCalls()), ShouldEqual, 1)
					So(mockMongo.CreateTaskCalls()[0].Task.JobID, ShouldEqual, jobID)

					Convey("Then no error should be returned", func() {
						So(err, ShouldBeNil)

						Convey("And the task should be returned", func() {
							So(createdTask, ShouldNotBeNil)
							So(createdTask.ID, ShouldEqual, "task-123")
							So(createdTask.JobID, ShouldEqual, jobID)
							So(createdTask.Type, ShouldEqual, domain.TaskTypeDataset)
							So(createdTask.State, ShouldEqual, domain.JobStateSubmitted)
						})
					})
				})
			})
		})
	})

	Convey("Given a job service and store where job does not exist", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return nil, appErrors.ErrJobNotFound
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := nonExistentJobID

		Convey("When a task is created", func() {
			task := &domain.Task{
				ID:   "task-123",
				Type: domain.TaskTypeDataset,
				Source: &domain.TaskMetadata{
					ID:  "source-1",
					URI: "/data/source",
				},
				Target: &domain.TaskMetadata{
					ID:  "target-1",
					URI: "/data/target",
				},
			}

			createdTask, err := jobService.CreateTask(ctx, jobID, task)

			Convey("Then an error should be returned", func() {
				So(createdTask, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, appErrors.ErrJobNotFound)

				Convey("And the store should not be called to create the task", func() {
					So(len(mockMongo.CreateTaskCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a job service and store that returns an error when creating task", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return &domain.Job{
					ID:    jobID,
					State: domain.JobStateSubmitted,
				}, nil
			},
			CreateTaskFunc: func(ctx context.Context, task *domain.Task) error {
				return errors.New("fake error for testing")
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When a task is created", func() {
			task := &domain.Task{
				ID:   "task-123",
				Type: domain.TaskTypeDataset,
				Source: &domain.TaskMetadata{
					ID:  "source-1",
					URI: "/data/source",
				},
				Target: &domain.TaskMetadata{
					ID:  "target-1",
					URI: "/data/target",
				},
			}

			createdTask, err := jobService.CreateTask(ctx, jobID, task)

			Convey("Then an error should be returned", func() {
				So(createdTask, ShouldBeNil)
				So(err, ShouldNotBeNil)
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
						Type:  domain.TaskTypeDataset,
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
						Type:  domain.TaskTypeDatasetEdition,
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

		jobService := Setup(&mockStore, &mockClients)

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
						So(tasks[0].Type, ShouldEqual, domain.TaskTypeDataset)
						So(tasks[1].ID, ShouldEqual, "task2")
						So(tasks[1].State, ShouldEqual, domain.JobStatePublishing)
						So(tasks[1].Type, ShouldEqual, domain.TaskTypeDatasetEdition)
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

		jobService := Setup(&mockStore, &mockClients)

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

		jobService := Setup(&mockStore, &mockClients)

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

		jobService := Setup(&mockStore, &mockClients)

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

		jobService := Setup(&mockStore, &mockClients)

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

		jobService := Setup(&mockStore, &mockClients)

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

func TestCreateEvent(t *testing.T) {
	Convey("Given a job service and store that has a stored job", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return &domain.Job{
					ID:    jobID,
					State: domain.JobStateSubmitted,
				}, nil
			},
			CreateEventFunc: func(ctx context.Context, event *domain.Event) error {
				return nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When an event is created", func() {
			createdAtTime := time.Now().UTC()
			createdAtStr := createdAtTime.Format(time.RFC3339)

			event := &domain.Event{
				ID:        "event-123",
				Action:    "approved",
				CreatedAt: createdAtStr,
				RequestedBy: &domain.User{
					ID:    "user-123",
					Email: "publisher@ons.gov.uk",
				},
			}

			createdEvent, err := jobService.CreateEvent(ctx, jobID, event)

			Convey("Then the store should be called to get the job", func() {
				So(len(mockMongo.GetJobCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobCalls()[0].JobID, ShouldEqual, jobID)

				Convey("Then the store should be called to create the event", func() {
					So(len(mockMongo.CreateEventCalls()), ShouldEqual, 1)
					So(mockMongo.CreateEventCalls()[0].Event.JobID, ShouldEqual, jobID)

					Convey("Then no error should be returned", func() {
						So(err, ShouldBeNil)

						Convey("And the event should be returned with job ID set", func() {
							So(createdEvent, ShouldNotBeNil)
							So(createdEvent.ID, ShouldEqual, "event-123")
							So(createdEvent.JobID, ShouldEqual, jobID)
							So(createdEvent.Action, ShouldEqual, string(domain.JobStateApproved))
							So(createdEvent.RequestedBy.ID, ShouldEqual, "user-123")
							So(createdEvent.RequestedBy.Email, ShouldEqual, "publisher@ons.gov.uk")
						})
					})
				})
			})
		})
	})

	Convey("Given a job service and store where job does not exist", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return nil, appErrors.ErrJobNotFound
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := nonExistentJobID

		Convey("When an event is created", func() {
			event := &domain.Event{
				ID:     "event-123",
				Action: "approved",
				RequestedBy: &domain.User{
					ID:    "user-123",
					Email: "publisher@ons.gov.uk",
				},
			}

			createdEvent, err := jobService.CreateEvent(ctx, jobID, event)

			Convey("Then an error should be returned", func() {
				So(createdEvent, ShouldBeNil)
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, appErrors.ErrJobNotFound)

				Convey("And the store should not be called to create the event", func() {
					So(len(mockMongo.CreateEventCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a job service and store that returns an error when creating event", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return &domain.Job{
					ID:    jobID,
					State: domain.JobStateSubmitted,
				}, nil
			},
			CreateEventFunc: func(ctx context.Context, event *domain.Event) error {
				return errors.New("fake error for testing")
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When an event is created", func() {
			event := &domain.Event{
				ID:     "event-123",
				Action: "approved",
				RequestedBy: &domain.User{
					ID:    "user-123",
					Email: "publisher@ons.gov.uk",
				},
			}

			createdEvent, err := jobService.CreateEvent(ctx, jobID, event)

			Convey("Then an error should be returned", func() {
				So(createdEvent, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a job service and store with event containing optional email field", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return &domain.Job{
					ID:    jobID,
					State: domain.JobStateSubmitted,
				}, nil
			},
			CreateEventFunc: func(ctx context.Context, event *domain.Event) error {
				return nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When an event is created without email in requested_by", func() {
			event := &domain.Event{
				ID:     "event-123",
				Action: "migrating",
				RequestedBy: &domain.User{
					ID: "user-456",
				},
			}

			createdEvent, err := jobService.CreateEvent(ctx, jobID, event)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And the event should be created successfully", func() {
					So(createdEvent, ShouldNotBeNil)
					So(createdEvent.RequestedBy.ID, ShouldEqual, "user-456")
					So(createdEvent.RequestedBy.Email, ShouldEqual, "")
				})
			})
		})
	})

	Convey("Given a job service and store with different action states", t, func() {
		testCases := []struct {
			action string
			name   string
		}{
			{"submitted", "submitted"},
			{"approved", "approved"},
			{"migrating", "migrating"},
			{"publishing", "publishing"},
			{"completed", "completed"},
			{"failed_migration", "failed_migration"},
			{"cancelled", "cancelled"},
		}

		for _, tc := range testCases {
			// Capture tc in a local variable to avoid closure issues
			testCase := tc

			Convey(fmt.Sprintf("When an event with action '%s' is created", testCase.name), func() {
				mockMongo := &storeMocks.MongoDBMock{
					GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
						return &domain.Job{
							ID:    jobID,
							State: domain.JobStateSubmitted,
						}, nil
					},
					CreateEventFunc: func(ctx context.Context, event *domain.Event) error {
						return nil
					},
				}

				mockStore := store.Datastore{
					Backend: mockMongo,
				}

				mockClients := clients.ClientList{}

				jobService := Setup(&mockStore, &mockClients)

				ctx := context.Background()
				jobID := testJobID

				event := &domain.Event{
					ID:        fmt.Sprintf("event-%s", testCase.name),
					Action:    testCase.action, // Use the correct action from test case
					CreatedAt: time.Now().UTC().Format(time.RFC3339),
					RequestedBy: &domain.User{
						ID:    "user-123",
						Email: "publisher@ons.gov.uk",
					},
				}

				createdEvent, err := jobService.CreateEvent(ctx, jobID, event)

				Convey("Then the event should be created with the correct action", func() {
					So(err, ShouldBeNil)
					So(createdEvent.Action, ShouldEqual, testCase.action) // Compare with string directly
				})
			})
		}
	})
}

func TestGetJobEvents(t *testing.T) {
	Convey("Given a job service and store that has stored events for a job", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobEventsFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
				return []*domain.Event{
					{
						ID:        "event-1",
						JobID:     jobID,
						CreatedAt: "2025-11-19T13:30:00Z",
						Action:    "submitted",
						RequestedBy: &domain.User{
							ID:    "user-1",
							Email: "user1@ons.gov.uk",
						},
					},
					{
						ID:        "event-2",
						JobID:     jobID,
						CreatedAt: "2025-11-19T13:35:00Z",
						Action:    "approved",
						RequestedBy: &domain.User{
							ID:    "user-2",
							Email: "user2@ons.gov.uk",
						},
					},
				}, 2, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When job events are retrieved", func() {
			events, totalCount, err := jobService.GetJobEvents(ctx, jobID, 10, 0)

			Convey("Then the store should be called to get job events", func() {
				So(len(mockMongo.GetJobEventsCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobEventsCalls()[0].JobID, ShouldEqual, jobID)
				So(mockMongo.GetJobEventsCalls()[0].Limit, ShouldEqual, 10)
				So(mockMongo.GetJobEventsCalls()[0].Offset, ShouldEqual, 0)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And the events should be returned", func() {
						So(len(events), ShouldEqual, 2)
						So(totalCount, ShouldEqual, 2)
						So(events[0].ID, ShouldEqual, "event-1")
						So(events[0].JobID, ShouldEqual, jobID)
						So(events[0].Action, ShouldEqual, "submitted")
						So(events[1].ID, ShouldEqual, "event-2")
						So(events[1].Action, ShouldEqual, "approved")
					})
				})
			})
		})
	})

	Convey("Given a job service and store that has no events for a job", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobEventsFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
				return []*domain.Event{}, 0, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When job events are retrieved", func() {
			events, totalCount, err := jobService.GetJobEvents(ctx, jobID, 10, 0)

			Convey("Then the store should be called to get job events", func() {
				So(len(mockMongo.GetJobEventsCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobEventsCalls()[0].JobID, ShouldEqual, jobID)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And an empty event list should be returned", func() {
						So(len(events), ShouldEqual, 0)
						So(totalCount, ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that returns an error when getting job events", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobEventsFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
				return nil, 0, errors.New("fake error for testing")
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When job events are retrieved", func() {
			events, totalCount, err := jobService.GetJobEvents(ctx, jobID, 10, 0)

			Convey("Then the store should be called to get job events", func() {
				So(len(mockMongo.GetJobEventsCalls()), ShouldEqual, 1)

				Convey("Then an error should be returned", func() {
					So(events, ShouldBeNil)
					So(totalCount, ShouldEqual, 0)
					So(err, ShouldNotBeNil)
				})
			})
		})
	})

	Convey("Given a job service and store with pagination parameters", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobEventsFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
				return []*domain.Event{
					{
						ID:        "event-2",
						JobID:     jobID,
						CreatedAt: "2025-11-19T13:35:00Z",
						Action:    "approved",
						RequestedBy: &domain.User{
							ID:    "user-2",
							Email: "user2@ons.gov.uk",
						},
					},
				}, 5, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When job events are retrieved with limit and offset", func() {
			events, totalCount, err := jobService.GetJobEvents(ctx, jobID, 1, 1)

			Convey("Then the store should be called with correct pagination parameters", func() {
				So(len(mockMongo.GetJobEventsCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobEventsCalls()[0].Limit, ShouldEqual, 1)
				So(mockMongo.GetJobEventsCalls()[0].Offset, ShouldEqual, 1)

				Convey("Then the paginated events should be returned", func() {
					So(err, ShouldBeNil)
					So(len(events), ShouldEqual, 1)
					So(totalCount, ShouldEqual, 5)
					So(events[0].ID, ShouldEqual, "event-2")
				})
			})
		})
	})

	Convey("Given a job service and store with multiple events and pagination", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobEventsFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
				if limit == 2 && offset == 0 {
					return []*domain.Event{
						{
							ID:        "event-1",
							JobID:     jobID,
							CreatedAt: "2025-11-19T13:30:00Z",
							Action:    "submitted",
							RequestedBy: &domain.User{
								ID: "user-1",
							},
						},
						{
							ID:        "event-2",
							JobID:     jobID,
							CreatedAt: "2025-11-19T13:35:00Z",
							Action:    "approved",
							RequestedBy: &domain.User{
								ID: "user-2",
							},
						},
					}, 4, nil
				}
				if limit == 2 && offset == 2 {
					return []*domain.Event{
						{
							ID:        "event-3",
							JobID:     jobID,
							CreatedAt: "2025-11-19T13:40:00Z",
							Action:    "migrating",
							RequestedBy: &domain.User{
								ID: "user-3",
							},
						},
						{
							ID:        "event-4",
							JobID:     jobID,
							CreatedAt: "2025-11-19T13:45:00Z",
							Action:    "completed",
							RequestedBy: &domain.User{
								ID: "user-4",
							},
						},
					}, 4, nil
				}
				return []*domain.Event{}, 0, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When first page of events is retrieved", func() {
			events, totalCount, err := jobService.GetJobEvents(ctx, jobID, 2, 0)

			Convey("Then first page events should be returned", func() {
				So(err, ShouldBeNil)
				So(len(events), ShouldEqual, 2)
				So(totalCount, ShouldEqual, 4)
				So(events[0].ID, ShouldEqual, "event-1")
				So(events[1].ID, ShouldEqual, "event-2")
			})
		})

		Convey("When second page of events is retrieved", func() {
			events, totalCount, err := jobService.GetJobEvents(ctx, jobID, 2, 2)

			Convey("Then second page events should be returned", func() {
				So(err, ShouldBeNil)
				So(len(events), ShouldEqual, 2)
				So(totalCount, ShouldEqual, 4)
				So(events[0].ID, ShouldEqual, "event-3")
				So(events[1].ID, ShouldEqual, "event-4")
			})
		})
	})

	Convey("Given a job service and store with events from different users", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobEventsFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
				return []*domain.Event{
					{
						ID:        "event-human",
						JobID:     jobID,
						CreatedAt: "2025-11-19T13:30:00Z",
						Action:    "approved",
						RequestedBy: &domain.User{
							ID:    "human-user",
							Email: "human@ons.gov.uk",
						},
					},
					{
						ID:        "event-service",
						JobID:     jobID,
						CreatedAt: "2025-11-19T13:35:00Z",
						Action:    "migrating",
						RequestedBy: &domain.User{
							ID: "service-account",
						},
					},
				}, 2, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When job events are retrieved", func() {
			events, totalCount, err := jobService.GetJobEvents(ctx, jobID, 10, 0)

			Convey("Then events from both human and service users should be returned", func() {
				So(err, ShouldBeNil)
				So(len(events), ShouldEqual, 2)
				So(totalCount, ShouldEqual, 2)
				So(events[0].RequestedBy.Email, ShouldEqual, "human@ons.gov.uk")
				So(events[1].RequestedBy.Email, ShouldEqual, "")
			})
		})
	})
}

func TestCountEventsByJobID(t *testing.T) {
	Convey("Given a job service and store that has events for a job", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			CountEventsByJobIDFunc: func(ctx context.Context, jobID string) (int, error) {
				return 5, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When event count is retrieved", func() {
			count, err := jobService.CountEventsByJobID(ctx, jobID)

			Convey("Then the store should be called to count events", func() {
				So(len(mockMongo.CountEventsByJobIDCalls()), ShouldEqual, 1)
				So(mockMongo.CountEventsByJobIDCalls()[0].JobID, ShouldEqual, jobID)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And the event count should be returned", func() {
						So(count, ShouldEqual, 5)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that has no events for a job", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			CountEventsByJobIDFunc: func(ctx context.Context, jobID string) (int, error) {
				return 0, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When event count is retrieved", func() {
			count, err := jobService.CountEventsByJobID(ctx, jobID)

			Convey("Then the store should be called to count events", func() {
				So(len(mockMongo.CountEventsByJobIDCalls()), ShouldEqual, 1)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And zero count should be returned", func() {
						So(count, ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that returns an error when counting events", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			CountEventsByJobIDFunc: func(ctx context.Context, jobID string) (int, error) {
				return 0, errors.New("fake error for testing")
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When event count is retrieved", func() {
			count, err := jobService.CountEventsByJobID(ctx, jobID)

			Convey("Then the store should be called to count events", func() {
				So(len(mockMongo.CountEventsByJobIDCalls()), ShouldEqual, 1)

				Convey("Then an error should be returned", func() {
					So(count, ShouldEqual, 0)
					So(err, ShouldNotBeNil)
				})
			})
		})
	})

	Convey("Given a job service and store with large event count", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			CountEventsByJobIDFunc: func(ctx context.Context, jobID string) (int, error) {
				return 1000, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()
		jobID := testJobID

		Convey("When event count is retrieved for job with many events", func() {
			count, err := jobService.CountEventsByJobID(ctx, jobID)

			Convey("Then the large count should be returned", func() {
				So(err, ShouldBeNil)
				So(count, ShouldEqual, 1000)
			})
		})
	})

	Convey("Given a job service and store with multiple jobs", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			CountEventsByJobIDFunc: func(ctx context.Context, jobID string) (int, error) {
				if jobID == "job-1" {
					return 3, nil
				}
				if jobID == "job-2" {
					return 7, nil
				}
				return 0, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()

		Convey("When event count is retrieved for different jobs", func() {
			count1, err1 := jobService.CountEventsByJobID(ctx, "job-1")
			count2, err2 := jobService.CountEventsByJobID(ctx, "job-2")

			Convey("Then correct counts should be returned for each job", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(count1, ShouldEqual, 3)
				So(count2, ShouldEqual, 7)
			})
		})
	})
}
