package migrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ONSdigital/dis-migration-service/application"
	applicationMocks "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/cache"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/executor"
	executorMocks "github.com/ONSdigital/dis-migration-service/executor/mock"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	fakeTaskType domain.TaskType = "fake-task-type"
	fakeTaskID   string          = "fake-task-id"
)

func TestMigratorExecuteTask(t *testing.T) {
	Convey("Given a migrator with test executors", t, func() {
		mockTestExecutor := &executorMocks.TaskExecutorMock{
			MigrateFunc: func(ctx context.Context, task *domain.Task) error {
				return nil
			},
		}

		getTaskExecutors = func(_ application.JobService, _ *clients.ClientList, _ *config.Config, _ *cache.TopicCache) map[domain.TaskType]executor.TaskExecutor {
			return map[domain.TaskType]executor.TaskExecutor{
				fakeTaskType: mockTestExecutor,
			}
		}

		mockJobService := &applicationMocks.JobServiceMock{
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, newState domain.State) error {
				return nil
			},
			ClaimTaskFunc: func(ctx context.Context) (*domain.Task, error) {
				return &domain.Task{
					Type: fakeTaskType,
				}, nil
			},
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) { return &domain.Job{}, nil },
			GetNextJobNumberFunc: func(ctx context.Context) (*domain.Counter, error) {
				fakeCounter := domain.Counter{}
				return &fakeCounter, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{}, 0, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 0, nil
			},
		}

		mockClients := &clients.ClientList{}
		mockSlackClient := createMockSlackClient()
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient)
		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		mig.SetTopicCache(nil)
		ctx := context.Background()

		Convey("When a task in state migrating is executed", func() {
			task := &domain.Task{
				ID:        fakeTaskID,
				JobNumber: 101,
				Type:      fakeTaskType,
				State:     domain.StateMigrating,
			}

			mig.executeTask(ctx, task)
			mig.wg.Wait()

			Convey("Then the executor is called to migrate", func() {
				So(len(mockTestExecutor.MigrateCalls()), ShouldEqual, 1)
				So(mockTestExecutor.MigrateCalls()[0].Task.Type, ShouldEqual, fakeTaskType)
			})
		})

		Convey("When a task in an unknown state is executed", func() {
			task := &domain.Task{
				Type:  fakeTaskType,
				State: "unknown-state",
			}

			mig.executeTask(ctx, task)
			mig.wg.Wait()

			Convey("Then the executor is not called to migrate", func() {
				So(len(mockTestExecutor.MigrateCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Given a migrator with no executors", t, func() {
		getTaskExecutors = func(_ application.JobService, _ *clients.ClientList, _ *config.Config, _ *cache.TopicCache) map[domain.TaskType]executor.TaskExecutor {
			return map[domain.TaskType]executor.TaskExecutor{}
		}

		mockJobService := &applicationMocks.JobServiceMock{
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, newState domain.State) error {
				return nil
			},
			GetNextJobNumberFunc: func(ctx context.Context) (*domain.Counter, error) {
				fakeCounter := domain.Counter{}
				return &fakeCounter, nil
			},
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					State:     domain.StateMigrating,
				}, nil
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State, userID string) error {
				return nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{}, 0, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 1, nil
			},
		}

		mockClients := &clients.ClientList{}
		mockSlackClient := createMockSlackClient()
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient)
		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		mig.SetTopicCache(nil)
		ctx := context.Background()

		Convey("When a task is executed", func() {
			task := &domain.Task{
				Type:  "unknown-task-type",
				State: domain.StateMigrating,
			}

			mig.executeTask(ctx, task)
			mig.wg.Wait()

			Convey("Then the task is failed", func() {
				So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
				So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.StateFailedMigration)
			})
		})
	})

	Convey("Given a migrator with topic cache and test executor", t, func() {
		ctx := context.Background()
		var capturedTopicCache *cache.TopicCache

		mockTestExecutor := &executorMocks.TaskExecutorMock{
			MigrateFunc: func(ctx context.Context, task *domain.Task) error {
				return nil
			},
		}

		// Create and populate topic cache
		topicCache, _ := cache.NewTopicCache(ctx, nil)
		subtopicsMap := cache.NewSubTopicsMap()
		subtopicsMap.AppendSubtopicID("business", cache.Subtopic{
			ID:         "business-456",
			Slug:       "business",
			ParentSlug: "",
		})
		testTopic := &cache.Topic{
			ID:   cache.TopicCacheKey,
			List: subtopicsMap,
		}
		topicCache.Set(cache.TopicCacheKey, testTopic)

		getTaskExecutors = func(_ application.JobService, _ *clients.ClientList, _ *config.Config, tc *cache.TopicCache) map[domain.TaskType]executor.TaskExecutor {
			capturedTopicCache = tc
			return map[domain.TaskType]executor.TaskExecutor{
				fakeTaskType: mockTestExecutor,
			}
		}

		mockJobService := &applicationMocks.JobServiceMock{
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, newState domain.State) error {
				return nil
			},
			ClaimTaskFunc: func(ctx context.Context) (*domain.Task, error) {
				return &domain.Task{
					Type: fakeTaskType,
				}, nil
			},
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) { return &domain.Job{}, nil },
			GetNextJobNumberFunc: func(ctx context.Context) (*domain.Counter, error) {
				fakeCounter := domain.Counter{}
				return &fakeCounter, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{}, 0, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 0, nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		mig.SetTopicCache(topicCache)

		Convey("When SetTopicCache is called with a valid cache", func() {
			Convey("Then the task executors are initialized with the topic cache", func() {
				So(capturedTopicCache, ShouldNotBeNil)
				So(capturedTopicCache, ShouldEqual, topicCache)

				Convey("And when a task is executed", func() {
					task := &domain.Task{
						ID:        fakeTaskID,
						JobNumber: 101,
						Type:      fakeTaskType,
						State:     domain.StateMigrating,
					}

					mig.executeTask(ctx, task)
					mig.wg.Wait()

					Convey("Then the executor is called to migrate", func() {
						So(len(mockTestExecutor.MigrateCalls()), ShouldEqual, 1)
						So(mockTestExecutor.MigrateCalls()[0].Task.Type, ShouldEqual, fakeTaskType)
					})
				})
			})
		})
	})

	Convey("Given a migrator with an executor that fails to execute the task", t, func() {
		mockTaskExecutor := &executorMocks.TaskExecutorMock{
			MigrateFunc: func(ctx context.Context, task *domain.Task) error {
				return errors.New("migration error")
			},
		}

		getTaskExecutors = func(_ application.JobService, _ *clients.ClientList, _ *config.Config, _ *cache.TopicCache) map[domain.TaskType]executor.TaskExecutor {
			return map[domain.TaskType]executor.TaskExecutor{
				fakeTaskType: mockTaskExecutor,
			}
		}

		mockJobService := &applicationMocks.JobServiceMock{
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, newState domain.State) error {
				return nil
			},
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: fakeJobNumber,
					State:     domain.StateMigrating,
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{}, 0, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 1, nil
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State, userID string) error {
				return nil
			},
			GetNextJobNumberFunc: func(ctx context.Context) (*domain.Counter, error) {
				fakeCounter := domain.Counter{}
				return &fakeCounter, nil
			},
		}

		mockClients := &clients.ClientList{}
		mockSlackClient := createMockSlackClient()
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient)
		mig.SetTopicCache(nil)
		ctx := context.Background()

		Convey("When a task is executed that errors during migration", func() {
			task := &domain.Task{
				JobNumber: fakeJobNumber,
				Type:      fakeTaskType,
				State:     domain.StateMigrating,
			}
			mig.executeTask(ctx, task)
			mig.wg.Wait()

			Convey("Then the task is failed", func() {
				So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
				So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.StateFailedMigration)

				Convey("And the job is failed", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
					So(mockJobService.UpdateJobStateCalls()[0].NewState, ShouldEqual, domain.StateFailedMigration)
				})
			})
		})
	})
}

func TestMigratorFailTask(t *testing.T) {
	Convey("Given a migrator with a mock job service", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, newState domain.State) error {
				return nil
			},
			GetNextJobNumberFunc: func(ctx context.Context) (*domain.Counter, error) {
				fakeCounter := domain.Counter{}
				return &fakeCounter, nil
			},
		}

		mockClients := &clients.ClientList{}
		mockSlackClient := createMockSlackClient()
		cfg := &config.Config{}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient)
		mig.SetTopicCache(nil)
		ctx := context.Background()

		Convey("When failTask is called for a task with an active state", func() {
			task := &domain.Task{
				ID:    fakeTaskID,
				State: domain.StateMigrating,
			}

			err := mig.failTask(ctx, task, errors.New("test error"), failureReasonExecutionFailed)

			Convey("Then the job service is called to update the task state to failed", func() {
				So(err, ShouldBeNil)
				So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
				So(mockJobService.UpdateTaskStateCalls()[0].TaskID, ShouldEqual, fakeTaskID)
				So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.StateFailedMigration)
			})
		})

		Convey("When failTask is called for a task with a pending state", func() {
			task := &domain.Task{
				Type:  fakeTaskType,
				State: domain.StateSubmitted,
			}

			err := mig.failTask(ctx, task, errors.New("test error"), failureReasonExecutionFailed)

			Convey("Then the job service is not called to update the task", func() {
				So(err, ShouldNotBeNil)
				So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 0)
			})
		})
	})

	Convey("Given a migrator with a mock job service that errors when updating task state", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, newState domain.State) error {
				return errors.New("update error")
			},
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					Label:     "Test Job",
				}, nil
			},
			GetNextJobNumberFunc: func(ctx context.Context) (*domain.Counter, error) {
				fakeCounter := domain.Counter{}
				return &fakeCounter, nil
			},
		}

		mockClients := &clients.ClientList{}
		mockSlackClient := createMockSlackClient()
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient)
		ctx := context.Background()

		Convey("When failTask is called for a task", func() {
			task := &domain.Task{
				Type:  "test-task-type",
				State: domain.StateMigrating,
			}

			err := mig.failTask(ctx, task, errors.New("test error"), failureReasonExecutionFailed)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetTaskExecutor(t *testing.T) {
	Convey("Given a migrator with test executors", t, func() {
		mockTaskExecutor := &executorMocks.TaskExecutorMock{}

		getTaskExecutors = func(_ application.JobService, _ *clients.ClientList, _ *config.Config, _ *cache.TopicCache) map[domain.TaskType]executor.TaskExecutor {
			return map[domain.TaskType]executor.TaskExecutor{
				fakeTaskType: mockTaskExecutor,
			}
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetNextJobNumberFunc: func(ctx context.Context) (*domain.Counter, error) {
				fakeCounter := domain.Counter{}
				return &fakeCounter, nil
			},
		}

		mockClients := &clients.ClientList{}
		mockSlackClient := createMockSlackClient()
		cfg := &config.Config{}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient)
		mig.SetTopicCache(nil)
		ctx := context.Background()

		Convey("When getTaskExecutor is called for a task with a known type", func() {
			task := &domain.Task{
				Type: fakeTaskType,
			}

			taskExecutor, err := mig.getTaskExecutor(ctx, task)

			Convey("Then the correct executor is returned", func() {
				So(err, ShouldBeNil)
				So(taskExecutor, ShouldEqual, mockTaskExecutor)
			})
		})

		Convey("When getTaskExecutor is called for a task with an unknown type", func() {
			task := &domain.Task{
				Type: "unknown-task-type",
			}

			taskExecutor, err := mig.getTaskExecutor(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "no executor found for task type: unknown-task-type")
				So(taskExecutor, ShouldBeNil)
			})
		})
	})
}

func TestGetTaskExecutors(t *testing.T) {
	Convey("Given getTaskExecutors is called with nil topic cache", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{}

		mockClients := &clients.ClientList{}

		cfg := &config.Config{}

		Convey("When getTaskExecutors is called with nil topic cache", func() {
			executors := getTaskExecutors(mockJobService, mockClients, cfg, nil)

			Convey("Then a map of task executors is returned", func() {
				So(executors, ShouldNotBeNil)
				So(len(executors), ShouldBeGreaterThan, 0)
			})
		})
	})

	Convey("Given getTaskExecutors is called with a valid topic cache", t, func() {
		ctx := context.Background()
		mockJobService := &applicationMocks.JobServiceMock{}

		mockClients := &clients.ClientList{}

		cfg := &config.Config{}

		// Create a real topic cache
		topicCache, err := cache.NewTopicCache(ctx, nil)
		So(err, ShouldBeNil)

		// Populate cache with test data
		subtopicsMap := cache.NewSubTopicsMap()
		subtopicsMap.AppendSubtopicID("economy", cache.Subtopic{
			ID:         "economy-123",
			Slug:       "economy",
			ParentSlug: "",
		})
		testTopic := &cache.Topic{
			ID:   cache.TopicCacheKey,
			List: subtopicsMap,
		}
		topicCache.Set(cache.TopicCacheKey, testTopic)

		Convey("When getTaskExecutors is called with topic cache", func() {
			executors := getTaskExecutors(mockJobService, mockClients, cfg, topicCache)

			Convey("Then a map of task executors is returned", func() {
				So(executors, ShouldNotBeNil)
				So(len(executors), ShouldBeGreaterThan, 0)
			})
		})
	})
}

func TestMonitorTasks(t *testing.T) {
	Convey("Given a migrator with a mock job service that returns no tasks", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			ClaimTaskFunc: func(ctx context.Context) (*domain.Task, error) {
				return nil, nil
			},
			GetNextJobNumberFunc: func(ctx context.Context) (*domain.Counter, error) {
				fakeCounter := domain.Counter{}
				return &fakeCounter, nil
			},
		}

		mockClients := &clients.ClientList{}
		mockSlackClient := createMockSlackClient()
		cfg := &config.Config{
			MigratorPollInterval: 10 * time.Millisecond,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient)
		mig.SetTopicCache(nil)
		ctx, cancel := context.WithCancel(context.Background())

		Convey("When monitorTasks is started and runs one iteration", func() {
			go func() {
				mig.monitorTasks(ctx)
			}()

			time.Sleep(25 * time.Millisecond)
			cancel()

			Convey("Then the job service is called to claim tasks every poll interval", func() {
				So(len(mockJobService.ClaimTaskCalls()), ShouldEqual, 3)
			})
		})
	})

	Convey("Given a migrator with a mock job service that returns a task", t, func() {
		requests := 0

		mockJobService := &applicationMocks.JobServiceMock{
			ClaimTaskFunc: func(ctx context.Context) (*domain.Task, error) {
				if requests == 0 {
					requests += 1
					return &domain.Task{
						ID:    fakeTaskID,
						State: domain.StateMigrating,
						Type:  fakeTaskType,
					}, nil
				}
				return nil, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, newState domain.State) error {
				return nil
			},
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					State:     domain.StateMigrating,
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{}, 0, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 1, nil
			},
			GetNextJobNumberFunc: func(ctx context.Context) (*domain.Counter, error) {
				fakeCounter := domain.Counter{}
				return &fakeCounter, nil
			},
		}

		mockTaskExecutor := &executorMocks.TaskExecutorMock{
			MigrateFunc: func(ctx context.Context, task *domain.Task) error {
				return nil
			},
		}

		getTaskExecutors = func(_ application.JobService, _ *clients.ClientList, _ *config.Config, _ *cache.TopicCache) map[domain.TaskType]executor.TaskExecutor {
			return map[domain.TaskType]executor.TaskExecutor{
				fakeTaskType: mockTaskExecutor,
			}
		}

		mockClients := &clients.ClientList{}
		mockSlackClient := createMockSlackClient()
		cfg := &config.Config{
			MigratorPollInterval:            10 * time.Millisecond,
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient)
		mig.SetTopicCache(nil)
		ctx, cancel := context.WithCancel(context.Background())

		Convey("When monitorTasks is started and runs one iteration", func() {
			go func() {
				mig.monitorTasks(ctx)
			}()

			time.Sleep(25 * time.Millisecond)
			cancel()

			Convey("Then the job service is called to claim tasks", func() {
				So(len(mockJobService.ClaimTaskCalls()), ShouldBeGreaterThan, 3)

				Convey("And the task executor is called to migrate the claimed task", func() {
					So(len(mockTaskExecutor.MigrateCalls()), ShouldEqual, 1)
				})
			})
		})
	})
}
