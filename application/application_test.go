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
			GetJobsFunc: func(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error) {
				jobs := []*domain.Job{
					{ID: "job1", State: domain.JobStateSubmitted},
					{ID: "job2", State: domain.JobStateApproved},
					{ID: "job3", State: domain.JobStateCompleted},
				}
				return jobs, len(jobs), nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}
		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()

		Convey("When GetJobs is called with valid parameters", func() {
			states := []domain.JobState{
				domain.JobStateSubmitted,
				domain.JobStateApproved,
			}

			jobs, total, err := jobService.GetJobs(ctx, states, 20, 5)

			Convey("Then the store should be called with correct parameters", func() {
				So(len(mockMongo.GetJobsCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobsCalls()[0].States, ShouldResemble, states)
				So(mockMongo.GetJobsCalls()[0].Limit, ShouldEqual, 20)
				So(mockMongo.GetJobsCalls()[0].Offset, ShouldEqual, 5)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And the jobs should be returned", func() {
						So(total, ShouldEqual, 3)
						So(len(jobs), ShouldEqual, 3)
						So(jobs[0].ID, ShouldEqual, "job1")
						So(jobs[1].ID, ShouldEqual, "job2")
						So(jobs[2].ID, ShouldEqual, "job3")
					})
				})
			})
		})

		Convey("When GetJobs is called with nil states", func() {
			jobs, total, err := jobService.GetJobs(ctx, nil, 10, 0)

			Convey("Then the store should be called with nil states", func() {
				So(len(mockMongo.GetJobsCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobsCalls()[0].States, ShouldBeNil)
				So(mockMongo.GetJobsCalls()[0].Limit, ShouldEqual, 10)
				So(mockMongo.GetJobsCalls()[0].Offset, ShouldEqual, 0)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And the jobs should be returned", func() {
						So(total, ShouldEqual, 3)
						So(len(jobs), ShouldEqual, 3)
					})
				})
			})
		})

		Convey("When GetJobs is called with empty states slice", func() {
			emptyStates := []domain.JobState{}
			jobs, total, err := jobService.GetJobs(ctx, emptyStates, 10, 0)

			Convey("Then the store should be called with empty states", func() {
				So(len(mockMongo.GetJobsCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobsCalls()[0].States, ShouldResemble, emptyStates)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And the jobs should be returned", func() {
						So(total, ShouldEqual, 3)
						So(len(jobs), ShouldEqual, 3)
					})
				})
			})
		})
	})

	Convey("Given a job service and store that returns an error when getting jobs", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobsFunc: func(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error) {
				return nil, 0, errors.New("fake error for testing")
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()

		Convey("When GetJobs is called", func() {
			jobs, totalCount, err := jobService.GetJobs(ctx, nil, 10, 0)

			Convey("Then the store should be called to get jobs", func() {
				So(len(mockMongo.GetJobsCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobsCalls()[0].Limit, ShouldEqual, 10)
				So(mockMongo.GetJobsCalls()[0].Offset, ShouldEqual, 0)

				Convey("Then an error should be returned", func() {
					So(jobs, ShouldBeNil)
					So(totalCount, ShouldEqual, 0)
					So(err, ShouldNotBeNil)
				})
			})
		})
	})

	Convey("Given a job service and store that returns no jobs", t, func() {
		mockMongo := &storeMocks.MongoDBMock{
			GetJobsFunc: func(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error) {
				return []*domain.Job{}, 0, nil
			},
		}

		mockStore := store.Datastore{
			Backend: mockMongo,
		}

		mockClients := clients.ClientList{}

		jobService := Setup(&mockStore, &mockClients)

		ctx := context.Background()

		Convey("When GetJobs is called", func() {
			jobs, totalCount, err := jobService.GetJobs(ctx, nil, 10, 0)

			Convey("Then the store should be called", func() {
				So(len(mockMongo.GetJobsCalls()), ShouldEqual, 1)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And an empty list should be returned", func() {
						So(len(jobs), ShouldEqual, 0)
						So(totalCount, ShouldEqual, 0)
					})
				})
			})
		})
	})
}

func TestCreateTask(t *testing.T) {
	Convey("Given a job service and store", t, func() {
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

					Convey("Then no error should be returned", func() {
						So(err, ShouldBeNil)

						Convey("And the task should be returned", func() {
							So(createdTask, ShouldNotBeNil)
							So(createdTask.ID, ShouldEqual, "task-123")
							So(createdTask.Type, ShouldEqual, domain.TaskTypeDataset)
							So(createdTask.State, ShouldEqual, domain.JobStateSubmitted)
						})
					})
				})
			})
		})
	})

	Convey("Given a job service and store where a job does not exist", t, func() {
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

		Convey("When a task is created for that job", func() {
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

	Convey("Given a job service and store that returns an error when creating a task", t, func() {
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

					Convey("Then no error should be returned", func() {
						So(err, ShouldBeNil)

						Convey("And the event should be returned with job ID set", func() {
							So(createdEvent, ShouldNotBeNil)
							So(createdEvent.ID, ShouldEqual, "event-123")
							So(createdEvent.Action, ShouldEqual, string(domain.JobStateApproved))
							So(createdEvent.RequestedBy.ID, ShouldEqual, "user-123")
							So(createdEvent.RequestedBy.Email, ShouldEqual, "publisher@ons.gov.uk")
						})
					})
				})
			})
		})
	})

	Convey("Given a job service and store where a job does not exist", t, func() {
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

		Convey("When an event is created for that job", func() {
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

	Convey("Given a job service and store that returns an error when creating an event", t, func() {
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

	Convey("Given a job service and store with events where some users have emails and some do not", t, func() {
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

		Convey("When an event is created without an email", func() {
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
	Convey("Given the GetJobEvents service method", t, func() {
		ctx := context.Background()
		jobID := testJobID

		Convey("When GetJobEvents is called with default pagination", func() {
			mockEvents := []*domain.Event{
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
			}

			mockMongo := &storeMocks.MongoDBMock{
				GetJobEventsFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
					return mockEvents, len(mockEvents), nil
				},
			}

			mockStore := store.Datastore{Backend: mockMongo}
			mockClients := clients.ClientList{}
			jobService := Setup(&mockStore, &mockClients)

			events, total, err := jobService.GetJobEvents(ctx, jobID, 10, 0)

			Convey("Then the store GetJobEvents method should be called with correct parameters", func() {
				So(len(mockMongo.GetJobEventsCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobEventsCalls()[0].Ctx, ShouldEqual, ctx)
				So(mockMongo.GetJobEventsCalls()[0].JobID, ShouldEqual, jobID)
				So(mockMongo.GetJobEventsCalls()[0].Limit, ShouldEqual, 10)
				So(mockMongo.GetJobEventsCalls()[0].Offset, ShouldEqual, 0)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And the events should be returned", func() {
						So(events, ShouldEqual, mockEvents)
						So(total, ShouldEqual, 2)
					})
				})
			})
		})

		Convey("When GetJobEvents is called with custom pagination", func() {
			customLimit := 5
			customOffset := 10

			mockMongo := &storeMocks.MongoDBMock{
				GetJobEventsFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
					return []*domain.Event{{ID: "event-1", JobID: jobID}}, 20, nil
				},
			}

			mockStore := store.Datastore{Backend: mockMongo}
			mockClients := clients.ClientList{}
			jobService := Setup(&mockStore, &mockClients)

			events, total, err := jobService.GetJobEvents(ctx, jobID, customLimit, customOffset)

			Convey("Then the store should be called with the custom pagination parameters", func() {
				So(len(mockMongo.GetJobEventsCalls()), ShouldEqual, 1)
				So(mockMongo.GetJobEventsCalls()[0].Limit, ShouldEqual, customLimit)
				So(mockMongo.GetJobEventsCalls()[0].Offset, ShouldEqual, customOffset)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And the paginated result should be returned", func() {
						So(len(events), ShouldEqual, 1)
						So(total, ShouldEqual, 20)
						So(events[0].ID, ShouldEqual, "event-1")
					})
				})
			})
		})

		Convey("When the store returns an error", func() {
			expectedErr := errors.New("database connection failed")

			mockMongo := &storeMocks.MongoDBMock{
				GetJobEventsFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
					return nil, 0, expectedErr
				},
			}

			mockStore := store.Datastore{Backend: mockMongo}
			mockClients := clients.ClientList{}
			jobService := Setup(&mockStore, &mockClients)

			events, total, err := jobService.GetJobEvents(ctx, jobID, 10, 0)

			Convey("Then the store should be called", func() {
				So(len(mockMongo.GetJobEventsCalls()), ShouldEqual, 1)

				Convey("Then an error should be returned", func() {
					So(events, ShouldBeNil)
					So(total, ShouldEqual, 0)
					So(err, ShouldEqual, expectedErr)
				})
			})
		})

		Convey("When the store returns an empty list", func() {
			mockMongo := &storeMocks.MongoDBMock{
				GetJobEventsFunc: func(ctx context.Context, jobID string, limit, offset int) ([]*domain.Event, int, error) {
					return []*domain.Event{}, 0, nil
				},
			}

			mockStore := store.Datastore{Backend: mockMongo}
			mockClients := clients.ClientList{}
			jobService := Setup(&mockStore, &mockClients)

			events, total, err := jobService.GetJobEvents(ctx, jobID, 10, 0)

			Convey("Then the store should be called", func() {
				So(len(mockMongo.GetJobEventsCalls()), ShouldEqual, 1)

				Convey("Then no error should be returned", func() {
					So(err, ShouldBeNil)

					Convey("And an empty result should be returned", func() {
						So(len(events), ShouldEqual, 0)
						So(total, ShouldEqual, 0)
					})
				})
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
