package migrator

import (
	"context"
	"errors"
	"testing"

	applicationMocks "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/cache"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/slack"
	slackMocks "github.com/ONSdigital/dis-migration-service/slack/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCheckAndUpdateJobStateBasedOnTasks(t *testing.T) {
	Convey("Given a migrator and job service where all tasks are in target state", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					Label:     "Test Job",
					JobNumber: fakeJobNumber,
					State:     domain.StateMigrating,
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				// All 3 tasks in target state
				return []*domain.Task{
					{ID: "task-1", JobNumber: jobNumber, State: domain.StateInReview},
					{ID: "task-2", JobNumber: jobNumber, State: domain.StateInReview},
					{ID: "task-3", JobNumber: jobNumber, State: domain.StateInReview},
				}, 3, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 3, nil
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State, userID string) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()
		rule := StateTransitionRule{
			TaskTargetState: domain.StateInReview,
			JobTargetState:  domain.StateInReview,
			Description:     "All tasks migrated, job moves to in_review",
		}

		Convey("When checking and updating job state based on tasks", func() {
			err := mig.CheckAndUpdateJobStateBasedOnTasks(ctx, fakeJobNumber, rule)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And the job state should be updated to in_review", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
					So(mockJobService.UpdateJobStateCalls()[0].JobNumber, ShouldEqual, fakeJobNumber)
					So(mockJobService.UpdateJobStateCalls()[0].NewState, ShouldEqual, domain.StateInReview)
				})

				Convey("And a Slack notification should be sent", func() {
					So(len(mockSlackClient.SendInfoCalls()), ShouldEqual, 1)
					So(mockSlackClient.SendInfoCalls()[0].Summary, ShouldEqual, "Job migration completed successfully")
					So(mockSlackClient.SendInfoCalls()[0].Details["Job Label"], ShouldEqual, "Test Job")
					So(mockSlackClient.SendInfoCalls()[0].Details["Job Number"], ShouldEqual, fakeJobNumber)
				})
			})
		})
	})

	Convey("Given a migrator where all tasks complete publishing", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: fakeJobNumber,
					State:     domain.StatePublishing,
					Label:     "Publishing Job",
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{
					{ID: "task-1", JobNumber: jobNumber, State: domain.StatePublished},
					{ID: "task-2", JobNumber: jobNumber, State: domain.StatePublished},
				}, 2, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 2, nil
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State, userID string) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()
		rule := StateTransitionRule{
			TaskTargetState: domain.StatePublished,
			JobTargetState:  domain.StatePublished,
			Description:     "All tasks published, job moves to published",
		}

		Convey("When checking and updating job state based on tasks", func() {
			err := mig.CheckAndUpdateJobStateBasedOnTasks(ctx, fakeJobNumber, rule)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And a Slack notification should be sent with publishing summary", func() {
					So(len(mockSlackClient.SendInfoCalls()), ShouldEqual, 1)
					So(mockSlackClient.SendInfoCalls()[0].Summary, ShouldEqual, "Job publishing completed successfully")
					So(mockSlackClient.SendInfoCalls()[0].Details["Job Label"], ShouldEqual, "Publishing Job")
					So(mockSlackClient.SendInfoCalls()[0].Details["Job Number"], ShouldEqual, fakeJobNumber)
				})
			})
		})
	})

	Convey("Given a migrator where Slack notification fails", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return errors.New("slack API error")
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: fakeJobNumber,
					State:     domain.StateMigrating,
					Label:     "Test Job",
					//Number: 42,
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{
					{ID: "task-1", State: domain.StateInReview},
				}, 1, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 1, nil
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State, userID string) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()
		rule := StateTransitionRule{
			TaskTargetState: domain.StateInReview,
			JobTargetState:  domain.StateInReview,
			Description:     "All tasks migrated, job moves to in_review",
		}

		Convey("When checking and updating job state based on tasks", func() {
			err := mig.CheckAndUpdateJobStateBasedOnTasks(ctx, fakeJobNumber, rule)

			Convey("Then no error should be returned (Slack failure doesn't fail the operation)", func() {
				So(err, ShouldBeNil)

				Convey("And the job state should still be updated", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
				})

				Convey("And Slack notification was attempted", func() {
					So(len(mockSlackClient.SendInfoCalls()), ShouldEqual, 1)
				})
			})
		})
	})

	Convey("Given a migrator where not all tasks are in target state", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					State:     domain.StateMigrating,
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				// Only 2 out of 3 tasks in target state
				return []*domain.Task{
					{ID: "task-1", State: domain.StateInReview},
					{ID: "task-2", State: domain.StateInReview},
				}, 2, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 3, nil // 3 total tasks
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()
		rule := StateTransitionRule{
			TaskTargetState: domain.StateInReview,
			JobTargetState:  domain.StateInReview,
			Description:     "All tasks migrated, job moves to in_review",
		}

		Convey("When checking and updating job state based on tasks", func() {
			err := mig.CheckAndUpdateJobStateBasedOnTasks(ctx, fakeJobNumber, rule)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And the job state should not be updated", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 0)
				})

				Convey("And no Slack notification should be sent", func() {
					So(len(mockSlackClient.SendInfoCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a migrator that fails to get the job", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return nil, errors.New("database error")
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()
		rule := StateTransitionRule{
			TaskTargetState: domain.StateInReview,
			JobTargetState:  domain.StateInReview,
			Description:     "All tasks migrated, job moves to in_review",
		}

		Convey("When checking and updating job state based on tasks", func() {
			err := mig.CheckAndUpdateJobStateBasedOnTasks(ctx, fakeJobNumber, rule)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "database error")

				Convey("And the job state should not be updated", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 0)
				})

				Convey("And no Slack notification should be sent", func() {
					So(len(mockSlackClient.SendInfoCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a migrator that fails to count tasks in target state", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					State:     domain.StateMigrating,
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return nil, 0, errors.New("database error")
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()
		rule := StateTransitionRule{
			TaskTargetState: domain.StateInReview,
			JobTargetState:  domain.StateInReview,
			Description:     "All tasks migrated, job moves to in_review",
		}

		Convey("When checking and updating job state based on tasks", func() {
			err := mig.CheckAndUpdateJobStateBasedOnTasks(ctx, fakeJobNumber, rule)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "database error")

				Convey("And the job state should not be updated", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a migrator that fails to count total tasks", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
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
				return 0, errors.New("database error")
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()
		rule := StateTransitionRule{
			TaskTargetState: domain.StateInReview,
			JobTargetState:  domain.StateInReview,
			Description:     "All tasks migrated, job moves to in_review",
		}

		Convey("When checking and updating job state based on tasks", func() {
			err := mig.CheckAndUpdateJobStateBasedOnTasks(ctx, fakeJobNumber, rule)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "database error")

				Convey("And the job state should not be updated", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a migrator that fails to update job state", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					State:     domain.StateMigrating,
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{
					{ID: "task-1", State: domain.StateInReview},
				}, 1, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 1, nil
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State, userID string) error {
				return errors.New("database error")
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()
		rule := StateTransitionRule{
			TaskTargetState: domain.StateInReview,
			JobTargetState:  domain.StateInReview,
			Description:     "All tasks migrated, job moves to in_review",
		}

		Convey("When checking and updating job state based on tasks", func() {
			err := mig.CheckAndUpdateJobStateBasedOnTasks(ctx, fakeJobNumber, rule)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "database error")

				Convey("And the update job state should have been called", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
				})

				Convey("And no Slack notification should be sent (update failed)", func() {
					So(len(mockSlackClient.SendInfoCalls()), ShouldEqual, 0)
				})
			})
		})
	})
}

func TestTriggerJobStateTransitionIfComplete(t *testing.T) {
	Convey("Given a migrator with multiple transition rules where no conditions are met", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					State:     domain.StateMigrating,
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				// No tasks in target state
				return []*domain.Task{}, 0, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 3, nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()

		Convey("When triggering job state transition", func() {
			err := mig.TriggerJobStateTransitions(ctx, fakeJobNumber)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And no job state updates should occur", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 0)
				})

				Convey("And no Slack notifications should be sent", func() {
					So(len(mockSlackClient.SendInfoCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a migrator where the first rule condition is met", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					State:     domain.StateMigrating,
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				// For in_review state filter (first rule): return 2 - first rule SHOULD trigger
				if len(states) > 0 && states[0] == domain.StateInReview {
					return []*domain.Task{
						{ID: "task-1", State: domain.StateInReview},
						{ID: "task-2", State: domain.StateInReview},
					}, 2, nil
				}
				// For published state filter (second rule): return 0 - second rule should NOT trigger
				if len(states) > 0 && states[0] == domain.StatePublished {
					return []*domain.Task{}, 0, nil
				}
				// Default
				return []*domain.Task{}, 0, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 2, nil
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State, userID string) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()

		Convey("When triggering job state transition", func() {
			err := mig.TriggerJobStateTransitions(ctx, fakeJobNumber)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And job state should be updated to in_review exactly once", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
					So(mockJobService.UpdateJobStateCalls()[0].NewState, ShouldEqual, domain.StateInReview)
				})

				Convey("And a Slack notification should be sent exactly once", func() {
					So(len(mockSlackClient.SendInfoCalls()), ShouldEqual, 1)
					So(mockSlackClient.SendInfoCalls()[0].Summary, ShouldEqual, "Job migration completed successfully")
				})
			})
		})
	})

	Convey("Given a migrator where a transition rule check fails", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					State:     domain.StateMigrating,
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return nil, 0, errors.New("database error")
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()

		Convey("When triggering job state transition", func() {
			err := mig.TriggerJobStateTransitions(ctx, fakeJobNumber)

			Convey("Then an error should be returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "database error")
			})
		})
	})

	Convey("Given a migrator where the second rule condition is met", t, func() {
		mockSlackClient := &slackMocks.ClienterMock{
			SendInfoFunc: func(ctx context.Context, summary string, details slack.SlackDetails) error {
				return nil
			},
		}

		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					State:     domain.StatePublishing,
					Label:     "Test Publishing Job",
				}, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				// Simulate different behavior based on the filter
				// For in_review state filter (first rule): return 0 - first rule should NOT trigger
				if len(states) > 0 && states[0] == domain.StateInReview {
					return []*domain.Task{}, 0, nil
				}
				// For published state filter (second rule): return all tasks - second rule should trigger
				if len(states) > 0 && states[0] == domain.StatePublished {
					return []*domain.Task{
						{ID: "task-1", State: domain.StatePublished},
						{ID: "task-2", State: domain.StatePublished},
					}, 2, nil
				}
				// Default - shouldn't reach here
				return []*domain.Task{}, 0, nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 2, nil
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State, userID string) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		topicCache, _ := cache.NewPopulatedTopicCacheForTest(context.Background())
		mig, _ := NewDefaultMigrator(cfg, mockJobService, mockClients, mockSlackClient, topicCache)
		ctx := context.Background()

		Convey("When triggering job state transition", func() {
			err := mig.TriggerJobStateTransitions(ctx, fakeJobNumber)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And job state should be updated to published exactly once", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
					So(mockJobService.UpdateJobStateCalls()[0].NewState, ShouldEqual, domain.StatePublished)
				})

				Convey("And a Slack notification should be sent for publishing completion", func() {
					So(len(mockSlackClient.SendInfoCalls()), ShouldEqual, 1)
					So(mockSlackClient.SendInfoCalls()[0].Summary, ShouldEqual, "Job publishing completed successfully")
					So(mockSlackClient.SendInfoCalls()[0].Details["Job Number"], ShouldEqual, fakeJobNumber)
				})
			})
		})
	})
}

func Test_isActiveStateCompletion(t *testing.T) {
	Convey("Given different state transitions", t, func() {
		Convey("When transitioning from migrating to in_review", func() {
			result := isActiveStateCompletion(domain.StateMigrating, domain.StateInReview)
			Convey("Then it should be recognized as active state completion", func() {
				So(result, ShouldBeTrue)
			})
		})

		Convey("When transitioning from publishing to published", func() {
			result := isActiveStateCompletion(domain.StatePublishing, domain.StatePublished)
			Convey("Then it should be recognized as active state completion", func() {
				So(result, ShouldBeTrue)
			})
		})

		Convey("When transitioning from post_publishing to completed", func() {
			result := isActiveStateCompletion(domain.StatePostPublishing, domain.StateCompleted)
			Convey("Then it should be recognized as active state completion", func() {
				So(result, ShouldBeTrue)
			})
		})

		Convey("When transitioning from migrating to failed_migration", func() {
			result := isActiveStateCompletion(domain.StateMigrating, domain.StateFailedMigration)
			Convey("Then it should not be recognized as active state completion", func() {
				So(result, ShouldBeFalse)
			})
		})

		Convey("When transitioning from submitted to migrating", func() {
			result := isActiveStateCompletion(domain.StateSubmitted, domain.StateMigrating)
			Convey("Then it should not be recognized as active state completion", func() {
				So(result, ShouldBeFalse)
			})
		})

		Convey("When transitioning from in_review to approved", func() {
			result := isActiveStateCompletion(domain.StateInReview, domain.StateApproved)
			Convey("Then it should not be recognized as active state completion", func() {
				So(result, ShouldBeFalse)
			})
		})
	})
}

func Test_getJobCompletionSummary(t *testing.T) {
	Convey("Given different state transitions", t, func() {
		mig := &migrator{}

		Convey("When getting summary for migrating completion", func() {
			result := mig.getJobCompletionSummary(domain.StateMigrating, domain.StateInReview)
			Convey("Then it should return migration completion message", func() {
				So(result, ShouldEqual, "Job migration completed successfully")
			})
		})

		Convey("When getting summary for publishing completion", func() {
			result := mig.getJobCompletionSummary(domain.StatePublishing, domain.StatePublished)
			Convey("Then it should return publishing completion message", func() {
				So(result, ShouldEqual, "Job publishing completed successfully")
			})
		})

		Convey("When getting summary for post-publishing completion", func() {
			result := mig.getJobCompletionSummary(domain.StatePostPublishing, domain.StateCompleted)
			Convey("Then it should return post-publishing completion message", func() {
				So(result, ShouldEqual, "Job post-publishing completed successfully")
			})
		})

		Convey("When getting summary for other state transition", func() {
			result := mig.getJobCompletionSummary(domain.StateSubmitted, domain.StateMigrating)
			Convey("Then it should return generic state update message", func() {
				So(result, ShouldEqual, "Job state updated")
			})
		})
	})
}
