package migrator

import (
	"context"
	"errors"
	"testing"

	applicationMocks "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCheckAndUpdateJobStateBasedOnTasks(t *testing.T) {
	Convey("Given a migrator and job service where all tasks are in target state", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
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
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
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
			})
		})
	})

	Convey("Given a migrator where not all tasks are in target state", t, func() {
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

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
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
			})
		})
	})

	Convey("Given a migrator that fails to get the job", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return nil, errors.New("database error")
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
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

	Convey("Given a migrator that fails to count tasks in target state", t, func() {
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

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
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

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
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
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State) error {
				return errors.New("database error")
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
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
			})
		})
	})
}

func TestTriggerJobStateTransitionIfComplete(t *testing.T) {
	Convey("Given a migrator with multiple transition rules where no conditions are met", t, func() {
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

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()

		Convey("When triggering job state transition", func() {
			err := mig.TriggerJobStateTransitions(ctx, fakeJobNumber)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And no job state updates should occur", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a migrator where the first rule condition is met", t, func() {
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
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()

		Convey("When triggering job state transition", func() {
			err := mig.TriggerJobStateTransitions(ctx, fakeJobNumber)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And job state should be updated to in_review exactly once", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
					So(mockJobService.UpdateJobStateCalls()[0].NewState, ShouldEqual, domain.StateInReview)
				})
			})
		})
	})

	Convey("Given a migrator where a transition rule check fails", t, func() {
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

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
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
		mockJobService := &applicationMocks.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobNumber int) (*domain.Job, error) {
				return &domain.Job{
					JobNumber: jobNumber,
					State:     domain.StateInReview,
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
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, newState domain.State) error {
				return nil
			},
		}

		mockClients := &clients.ClientList{}
		cfg := &config.Config{
			MigratorMaxConcurrentExecutions: 1,
		}

		mig := NewDefaultMigrator(cfg, mockJobService, mockClients)
		ctx := context.Background()

		Convey("When triggering job state transition", func() {
			err := mig.TriggerJobStateTransitions(ctx, fakeJobNumber)

			Convey("Then no error should be returned", func() {
				So(err, ShouldBeNil)

				Convey("And job state should be updated to published exactly once", func() {
					So(len(mockJobService.UpdateJobStateCalls()), ShouldEqual, 1)
					So(mockJobService.UpdateJobStateCalls()[0].NewState, ShouldEqual, domain.StatePublished)
				})
			})
		})
	})
}
