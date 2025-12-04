package migrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dis-migration-service/application"
	applicationMocks "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/executor"
	executorMocks "github.com/ONSdigital/dis-migration-service/executor/mock"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	fakeJobType domain.JobType = "fake-job-type"
	fakeJobID   string         = "fake-job-id"
)

func TestMigratorExecuteJob(t *testing.T) {
	Convey("Given a migrator with test executors", t, func() {
		mockJobExecutor := &executorMocks.JobExecutorMock{
			MigrateFunc: func(ctx context.Context, job *domain.Job) error {
				return nil
			},
		}

		getJobExecutors = func(_ application.JobService, _ *clients.ClientList) map[domain.JobType]executor.JobExecutor {
			return map[domain.JobType]executor.JobExecutor{
				fakeJobType: mockJobExecutor,
			}
		}

		mockJobService := &applicationMocks.JobServiceMock{}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()

		Convey("When a job in state migrating is executed", func() {
			job := &domain.Job{
				ID: fakeJobID,
				Config: &domain.JobConfig{
					Type: fakeJobType,
				},
				State: domain.JobStateMigrating,
			}

			mig.executeJob(ctx, job)
			mig.wg.Wait()

			Convey("Then the executor is called to migrate", func() {
				So(len(mockJobExecutor.MigrateCalls()), ShouldEqual, 1)
				So(mockJobExecutor.MigrateCalls()[0].Job.ID, ShouldEqual, fakeJobID)
			})
		})

		Convey("When a job in an unknown state is executed", func() {
			job := &domain.Job{
				ID: fakeJobID,
				Config: &domain.JobConfig{
					Type: fakeJobType,
				},
				State: "unknown-state",
			}

			mig.executeJob(ctx, job)
			mig.wg.Wait()

			Convey("Then the executor is not called to migrate", func() {
				So(len(mockJobExecutor.MigrateCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Given a migrator with no executor for a job type", t, func() {
		getJobExecutors = func(_ application.JobService, _ *clients.ClientList) map[domain.JobType]executor.JobExecutor {
			return map[domain.JobType]executor.JobExecutor{}
		}

		mockJobService := &applicationMocks.JobServiceMock{
			UpdateJobStateFunc: func(ctx context.Context, job *domain.Job, state domain.JobState) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()

		Convey("When a job is executed", func() {
			job := &domain.Job{
				ID: fakeJobID,
				Config: &domain.JobConfig{
					Type: fakeJobType,
				},
				State: domain.JobStateMigrating,
			}

			mig.executeJob(ctx, job)
			mig.wg.Wait()

			Convey("Then the job is marked as failed", func() {
				So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
				So(mockJobService.UpdateJobStateCalls()[0].Job.ID, ShouldEqual, fakeJobID)
				So(mockJobService.UpdateJobStateCalls()[0].NewState, ShouldEqual, domain.JobStateFailedMigration)
			})
		})
	})

	Convey("Given a migrator with an executor that fails to execute the job", t, func() {
		mockJobExecutor := &executorMocks.JobExecutorMock{
			MigrateFunc: func(ctx context.Context, job *domain.Job) error {
				return errors.New("migration error")
			},
		}

		getJobExecutors = func(_ application.JobService, _ *clients.ClientList) map[domain.JobType]executor.JobExecutor {
			return map[domain.JobType]executor.JobExecutor{
				fakeJobType: mockJobExecutor,
			}
		}

		mockJobService := &applicationMocks.JobServiceMock{
			UpdateJobStateFunc: func(ctx context.Context, job *domain.Job, newState domain.JobState) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()

		Convey("When a job is executed that errors during migration", func() {
			job := &domain.Job{
				ID:    fakeJobID,
				State: domain.JobStateMigrating,
				Config: &domain.JobConfig{
					Type: fakeJobType,
				},
			}
			mig.executeJob(ctx, job)
			mig.wg.Wait()

			Convey("Then the job is failed", func() {
				So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
				So(mockJobService.UpdateJobStateCalls()[0].NewState, ShouldEqual, domain.JobStateFailedMigration)
			})
		})
	})
}

func TestMigratorFailJob(t *testing.T) {
	Convey("Given a migrator with a mock job service", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			UpdateJobStateFunc: func(ctx context.Context, job *domain.Job, state domain.JobState) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()

		Convey("When failJob is called for a job with an active state", func() {
			job := &domain.Job{
				ID:    fakeJobID,
				State: domain.JobStateMigrating,
			}

			err := mig.failJob(ctx, job)

			Convey("Then the job service is called to update the job state to failed", func() {
				So(err, ShouldBeNil)
				So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
				So(mockJobService.UpdateJobStateCalls()[0].Job.ID, ShouldEqual, fakeJobID)
				So(mockJobService.UpdateJobStateCalls()[0].NewState, ShouldEqual, domain.JobStateFailedMigration)
			})
		})

		Convey("When failJob is called for a job with a pending state", func() {
			job := &domain.Job{
				ID:    fakeJobID,
				State: domain.JobStateSubmitted,
			}

			err := mig.failJob(ctx, job)

			Convey("Then the job service is not called to update the job", func() {
				So(err, ShouldNotBeNil)
				So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Given a migrator with a mock job service that errors when updating job state", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			UpdateJobStateFunc: func(ctx context.Context, job *domain.Job, state domain.JobState) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()

		Convey("When failJob is called for a job", func() {
			job := &domain.Job{
				ID: "test-job-id",
			}

			err := mig.failJob(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestMigratorFailJobByID(t *testing.T) {
	Convey("Given a migrator with a mock job service that returns a job in an active state", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return &domain.Job{
					ID:    jobID,
					State: domain.JobStateMigrating,
				}, nil
			},
			UpdateJobStateFunc: func(ctx context.Context, job *domain.Job, state domain.JobState) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()
		Convey("When failJobByID is called", func() {
			err := mig.failJobByID(ctx, "test-job-id")

			Convey("Then the job service is called to update the job state to failed", func() {
				So(err, ShouldBeNil)
				So(len(mockJobService.GetJobCalls()), ShouldEqual, 1)
				So(mockJobService.GetJobCalls()[0].JobID, ShouldEqual, "test-job-id")
				So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
				So(mockJobService.UpdateJobStateCalls()[0].Job.ID, ShouldEqual, "test-job-id")
				So(mockJobService.UpdateJobStateCalls()[0].NewState, ShouldEqual, domain.JobStateFailedMigration)
			})
		})
	})
	Convey("Given a migrator with a mock job service that returns a job in a failed state", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return &domain.Job{
					ID:    jobID,
					State: domain.JobStateFailedMigration,
				}, nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()
		Convey("When failJobByID is called", func() {
			err := mig.failJobByID(ctx, "test-job-id")

			Convey("Then the job service is not called to update the job", func() {
				So(err, ShouldBeNil)
				So(len(mockJobService.GetJobCalls()), ShouldEqual, 1)
				So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 0)
			})
		})
	})
	Convey("Given a migrator with a mock job service that errors when getting a job", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return nil, errors.New("test error")
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()
		Convey("When failJobByID is called", func() {
			err := mig.failJobByID(ctx, "test-job-id")

			Convey("Then an error is returned", func() {
				So(err.Error(), ShouldEqual, "test error")
				So(len(mockJobService.GetJobCalls()), ShouldEqual, 1)
				So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 0)
			})
		})
	})
}

func TestGetJobExecutor(t *testing.T) {
	Convey("Given a migrator with test executors", t, func() {
		var fakeJobType domain.JobType = "fake-job-type"

		mockJobExecutor := &executorMocks.JobExecutorMock{}

		getJobExecutors = func(_ application.JobService, _ *clients.ClientList) map[domain.JobType]executor.JobExecutor {
			return map[domain.JobType]executor.JobExecutor{
				fakeJobType: mockJobExecutor,
			}
		}

		mockJobService := &applicationMocks.JobServiceMock{}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()

		Convey("When getJobExecutor is called for a job with a known type", func() {
			job := &domain.Job{
				Config: &domain.JobConfig{
					Type: fakeJobType,
				},
			}

			jobExecutor, err := mig.getJobExecutor(ctx, job)

			Convey("Then the correct executor is returned", func() {
				So(err, ShouldBeNil)
				So(jobExecutor, ShouldEqual, mockJobExecutor)
			})
		})

		Convey("When getJobExecutor is called for a job with an unknown type", func() {
			job := &domain.Job{
				Config: &domain.JobConfig{
					Type: "unknown-job-type",
				},
			}

			jobExecutor, err := mig.getJobExecutor(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "no executor found for task type: unknown-job-type")
				So(jobExecutor, ShouldBeNil)
			})
		})
	})
}

func TestMonitorJobs(t *testing.T) {
	Convey("Given a migrator with a mock job service that returns no jobs", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			ClaimJobFunc: func(ctx context.Context) (*domain.Job, error) {
				return nil, nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorPollInterval: 10 * time.Millisecond,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx, cancel := context.WithCancel(context.Background())

		Convey("When monitorJobs is started and runs one iteration", func() {
			go func() {
				mig.monitorJobs(ctx)
			}()

			// Allow some time for the monitor to run
			time.Sleep(25 * time.Millisecond)
			cancel()

			Convey("Then the job service is called to claim jobs every poll interval", func() {
				So(len(mockJobService.ClaimJobCalls()), ShouldEqual, 3)
			})
		})
	})

	Convey("Given a migrator with a mock job service that returns a job", t, func() {
		requests := 0

		mockJobService := &applicationMocks.JobServiceMock{
			ClaimJobFunc: func(ctx context.Context) (*domain.Job, error) {
				if requests == 0 {
					requests += 1
					return &domain.Job{
						ID:    fakeJobID,
						State: domain.JobStateMigrating,
						Config: &domain.JobConfig{
							Type: fakeJobType,
						},
					}, nil
				} else {
					return nil, nil
				}
			},
		}

		mockJobExecutor := &executorMocks.JobExecutorMock{
			MigrateFunc: func(ctx context.Context, job *domain.Job) error {
				return nil
			},
		}

		getJobExecutors = func(_ application.JobService, _ *clients.ClientList) map[domain.JobType]executor.JobExecutor {
			return map[domain.JobType]executor.JobExecutor{
				fakeJobType: mockJobExecutor,
			}
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorPollInterval:            10 * time.Millisecond,
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx, cancel := context.WithCancel(context.Background())

		Convey("When monitorJobs is started and runs one iteration", func() {
			go func() {
				mig.monitorJobs(ctx)
			}()

			// Allow some time for the monitor to run
			time.Sleep(25 * time.Millisecond)
			cancel()

			Convey("Then the job service is called to claim jobs", func() {
				So(len(mockJobService.ClaimJobCalls()), ShouldBeGreaterThan, 3)

				Convey("And the job executor is called to migrate the claimed job", func() {
					So(len(mockJobExecutor.MigrateCalls()), ShouldEqual, 1)
				})
			})
		})
	})
}
