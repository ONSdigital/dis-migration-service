package executor

import (
	"context"
	"errors"
	"testing"

	applicationMocks "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/clients"
	clientMocks "github.com/ONSdigital/dis-migration-service/clients/mock"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testJobNumber     = 1
	testCollectionID  = "migration-collection-1"
	testDatasetAPIURL = "http://localhost:22000"
)

var (
	errTest = errors.New("test error")
)

func TestJobStaticDataset(t *testing.T) {
	Convey("Given a static dataset job executor and a job service that does not error", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateJobCollectionIDFunc: func(ctx context.Context, jobNumber int, collectionID string) error {
				return nil
			},
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 2, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{
					{ID: "task-1", State: domain.StateInReview},
					{ID: "task-2", State: domain.StateInReview},
				}, 2, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error {
				return nil
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, state domain.State, userID string) error {
				return nil
			},
		}
		mockZebedeeClient := &clientMocks.ZebedeeClientMock{
			ApproveCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
				return nil
			},
			CreateCollectionFunc: func(ctx context.Context, userAuthToken string, collection zebedee.Collection) (zebedee.Collection, error) {
				collection.ID = testCollectionID
				return collection, nil
			},
			GetCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) (zebedee.Collection, error) {
				return zebedee.Collection{ApprovalStatus: zebedee.CollectionStatusComplete}, nil
			},
			PublishCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
				return nil
			},
		}
		mockClientList := &clients.ClientList{
			Zebedee: mockZebedeeClient,
		}

		ctx := context.Background()

		executor := NewStaticDatasetJobExecutor(mockJobService, mockClientList, "faketoken")

		Convey("When migrate is called for a job", func() {
			job := &domain.Job{
				JobNumber: 1,
				Config: &domain.JobConfig{
					SourceID: "source-dataset-id",
					TargetID: "target-dataset-id",
				},
				State: domain.StateMigrating,
			}

			err := executor.Migrate(ctx, job)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And a collection is created for the migration job", func() {
					So(len(mockZebedeeClient.CreateCollectionCalls()), ShouldEqual, 1)
					So(mockZebedeeClient.CreateCollectionCalls()[0].Collection.Name, ShouldEqual, "Migration Collection for Job 1")
					So(mockZebedeeClient.CreateCollectionCalls()[0].Collection.Type, ShouldEqual, zebedee.CollectionTypeAutomated)
					So(mockJobService.UpdateJobCollectionIDCalls()[0].CollectionID, ShouldEqual, testCollectionID)

					Convey("And a dataset series migration task is created for the dataset", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 1)
						So(mockJobService.CreateTaskCalls()[0].JobNumber, ShouldEqual, testJobNumber)
						So(mockJobService.CreateTaskCalls()[0].Task.Type, ShouldEqual, domain.TaskTypeDatasetSeries)
						So(mockJobService.CreateTaskCalls()[0].Task.Source.ID, ShouldEqual, "source-dataset-id")
						So(mockJobService.CreateTaskCalls()[0].Task.Target.ID, ShouldEqual, "target-dataset-id")
					})
				})
			})
		})

		Convey("When publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.Publish(ctx, job)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And the collection is approved", func() {
					So(mockZebedeeClient.ApproveCollectionCalls(), ShouldHaveLength, 1)
					So(mockZebedeeClient.GetCollectionCalls(), ShouldHaveLength, 1)

					Convey("And all tasks are updated to approved", func() {
						So(mockJobService.UpdateTaskStateCalls(), ShouldHaveLength, 2)
						So(mockJobService.UpdateTaskStateCalls()[0].TaskID, ShouldEqual, "task-1")
						So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.StateApproved)
						So(mockJobService.UpdateTaskStateCalls()[1].TaskID, ShouldEqual, "task-2")
						So(mockJobService.UpdateTaskStateCalls()[1].NewState, ShouldEqual, domain.StateApproved)
					})
				})
			})
		})

		Convey("When post-publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePostPublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.PostPublish(ctx, job)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And the zebedee collection is published", func() {
					So(mockZebedeeClient.PublishCollectionCalls(), ShouldHaveLength, 1)
					So(mockZebedeeClient.PublishCollectionCalls()[0].CollectionID, ShouldEqual, testCollectionID)

					Convey("And all tasks are updated to post-publishing then completed", func() {
						So(mockJobService.UpdateTaskStateCalls(), ShouldHaveLength, 2)
						So(mockJobService.UpdateTaskStateCalls()[0].TaskID, ShouldEqual, "task-1")
						So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.StatePendingPostPublish)
						So(mockJobService.UpdateTaskStateCalls()[1].TaskID, ShouldEqual, "task-2")
						So(mockJobService.UpdateTaskStateCalls()[1].NewState, ShouldEqual, domain.StatePendingPostPublish)
					})
				})
			})
		})
	})

	Convey("Given a static dataset job executor and a job service that errors when creating a task", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return nil, errTest
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, state domain.State, userID string) error {
				return nil
			},
			UpdateJobCollectionIDFunc: func(ctx context.Context, jobNumber int, collectionID string) error {
				return nil
			},
		}
		mockZebedeeClient := &clientMocks.ZebedeeClientMock{
			CreateCollectionFunc: func(ctx context.Context, userAuthToken string, collection zebedee.Collection) (zebedee.Collection, error) {
				return collection, nil
			},
		}
		mockClientList := &clients.ClientList{
			Zebedee: mockZebedeeClient,
		}

		ctx := context.Background()

		executor := NewStaticDatasetJobExecutor(mockJobService, mockClientList, "faketoken")

		Convey("When migrate is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				Config: &domain.JobConfig{
					SourceID: "source-dataset-id",
					TargetID: "target-dataset-id",
				},
				State: domain.StateMigrating,
			}

			err := executor.Migrate(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
			})
		})
	})

	Convey("Given a static dataset job executor and a zebedee client that errors when creating a collection", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, state domain.State, userID string) error {
				return nil
			},
		}
		mockZebedeeClient := &clientMocks.ZebedeeClientMock{
			CreateCollectionFunc: func(ctx context.Context, userAuthToken string, collection zebedee.Collection) (zebedee.Collection, error) {
				return zebedee.Collection{}, errTest
			},
		}
		mockClientList := &clients.ClientList{
			Zebedee: mockZebedeeClient,
		}

		ctx := context.Background()

		executor := NewStaticDatasetJobExecutor(mockJobService, mockClientList, "faketoken")

		Convey("When migrate is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				Config: &domain.JobConfig{
					SourceID: "source-dataset-id",
					TargetID: "target-dataset-id",
				},
				State: domain.StateMigrating,
			}

			err := executor.Migrate(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
				So(mockJobService.UpdateJobCollectionIDCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given a static dataset job executor and a zebedee client that errors when getting collection status", t, func() {
		mockZebedeeClient := &clientMocks.ZebedeeClientMock{
			ApproveCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
				return nil
			},
			GetCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) (zebedee.Collection, error) {
				return zebedee.Collection{}, errTest
			},
		}
		mockClientList := &clients.ClientList{
			Zebedee: mockZebedeeClient,
		}

		ctx := context.Background()
		executor := NewStaticDatasetJobExecutor(&applicationMocks.JobServiceMock{}, mockClientList, "faketoken")

		Convey("When publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.Publish(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
			})
		})
	})

	Convey("Given a static dataset job executor and a zebedee client that returns an error collection status", t, func() {
		mockZebedeeClient := &clientMocks.ZebedeeClientMock{
			ApproveCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
				return nil
			},
			GetCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) (zebedee.Collection, error) {
				return zebedee.Collection{ApprovalStatus: zebedee.CollectionStatusError}, nil
			},
		}
		mockClientList := &clients.ClientList{
			Zebedee: mockZebedeeClient,
		}

		ctx := context.Background()
		executor := NewStaticDatasetJobExecutor(&applicationMocks.JobServiceMock{}, mockClientList, "faketoken")

		Convey("When publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.Publish(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a static dataset job executor and a zebedee client that returns pending then approved collection status", t, func() {
		callCount := 0
		mockJobService := &applicationMocks.JobServiceMock{
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 1, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{
					{ID: "task-1", State: domain.StateInReview},
				}, 1, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error {
				return nil
			},
		}
		mockZebedeeClient := &clientMocks.ZebedeeClientMock{
			ApproveCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
				return nil
			},
			GetCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) (zebedee.Collection, error) {
				callCount++
				if callCount < 3 {
					return zebedee.Collection{ApprovalStatus: zebedee.CollectionStatusInProgress}, nil
				}
				return zebedee.Collection{ApprovalStatus: zebedee.CollectionStatusComplete}, nil
			},
		}
		mockClientList := &clients.ClientList{
			Zebedee: mockZebedeeClient,
		}

		ctx := context.Background()
		executor := NewStaticDatasetJobExecutor(mockJobService, mockClientList, "faketoken")

		Convey("When publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.Publish(ctx, job)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(mockZebedeeClient.GetCollectionCalls(), ShouldHaveLength, 3)
				So(mockJobService.UpdateTaskStateCalls(), ShouldHaveLength, 1)
			})
		})
	})

	Convey("Given a static dataset job executor and a job service that errors when getting tasks", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 2, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber, limit, offset int) ([]*domain.Task, int, error) {
				return nil, 0, errTest
			},
		}
		mockClientList := &clients.ClientList{
			Zebedee: &clientMocks.ZebedeeClientMock{
				ApproveCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
					return nil
				},
				GetCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) (zebedee.Collection, error) {
					return zebedee.Collection{ApprovalStatus: zebedee.CollectionStatusComplete}, nil
				},
				PublishCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
					return nil
				},
			},
		}

		ctx := context.Background()
		executor := NewStaticDatasetJobExecutor(mockJobService, mockClientList, "faketoken")

		Convey("When publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.Publish(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
				So(mockJobService.UpdateTaskStateCalls(), ShouldHaveLength, 0)
			})
		})

		Convey("When post-publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePostPublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.PostPublish(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
				So(mockJobService.UpdateTaskStateCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given a static dataset job executor and a job service that errors when updating task state", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 2, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{
					{ID: "task-1", State: domain.StateInReview},
					{ID: "task-2", State: domain.StateInReview},
				}, 2, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error {
				return errTest
			},
		}
		mockClientList := &clients.ClientList{
			Zebedee: &clientMocks.ZebedeeClientMock{
				ApproveCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
					return nil
				},
				GetCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) (zebedee.Collection, error) {
					return zebedee.Collection{ApprovalStatus: zebedee.CollectionStatusComplete}, nil
				},
				PublishCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
					return nil
				},
			},
		}

		ctx := context.Background()
		executor := NewStaticDatasetJobExecutor(mockJobService, mockClientList, "faketoken")

		Convey("When publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.Publish(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
				So(mockJobService.UpdateTaskStateCalls(), ShouldHaveLength, 1)
			})
		})

		Convey("When post-publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePendingPostPublish,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.PostPublish(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
				So(mockJobService.UpdateTaskStateCalls(), ShouldHaveLength, 1)
			})
		})
	})

	Convey("Given a static dataset job executor and a job service that errors when counting tasks", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 0, errTest
			},
		}
		mockClientList := &clients.ClientList{
			Zebedee: &clientMocks.ZebedeeClientMock{
				ApproveCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
					return nil
				},
				GetCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) (zebedee.Collection, error) {
					return zebedee.Collection{ApprovalStatus: zebedee.CollectionStatusComplete}, nil
				},
				PublishCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
					return nil
				},
			},
		}

		ctx := context.Background()
		executor := NewStaticDatasetJobExecutor(mockJobService, mockClientList, "faketoken")

		Convey("When publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.Publish(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
				So(mockJobService.GetJobTasksCalls(), ShouldHaveLength, 0)
			})
		})

		Convey("When post-publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePostPublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.PostPublish(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
				So(mockJobService.GetJobTasksCalls(), ShouldHaveLength, 0)
			})
		})
	})

	Convey("Given a static dataset job executor and a zebedee client that errors when publishing the collection", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 2, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{
					{ID: "task-1", State: domain.StatePublished},
					{ID: "task-2", State: domain.StatePublished},
				}, 2, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error {
				return nil
			},
		}
		mockZebedeeClient := &clientMocks.ZebedeeClientMock{
			PublishCollectionFunc: func(ctx context.Context, userAuthToken string, collectionID string) error {
				return errTest
			},
		}
		mockClientList := &clients.ClientList{
			Zebedee: mockZebedeeClient,
		}

		ctx := context.Background()
		executor := NewStaticDatasetJobExecutor(mockJobService, mockClientList, "faketoken")

		Convey("When post-publish is called for a job", func() {
			job := &domain.Job{
				JobNumber: testJobNumber,
				State:     domain.StatePostPublishing,
				Config: &domain.JobConfig{
					CollectionID: testCollectionID,
				},
			}

			err := executor.PostPublish(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
			})
		})
	})

	Convey("Given a static dataset job executor with download tasks for reversion", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 1, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{
					{
						ID:   "download-task-1",
						Type: domain.TaskTypeDatasetDownload,
						Source: &domain.TaskMetadata{
							ID: "/source/file.csv",
						},
						Target: &domain.TaskMetadata{
							DatasetID: "target-dataset-id",
							EditionID: "historical",
							VersionID: "1",
						},
					},
				}, 1, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error { return nil },
		}

		executor := NewStaticDatasetJobExecutor(
			mockJobService,
			&clients.ClientList{},
			"faketoken",
		)

		Convey("When revert is called for a job", func() {
			err := executor.Revert(context.Background(), &domain.Job{
				JobNumber: testJobNumber,
				Config: &domain.JobConfig{
					TargetID: "target-dataset-id",
				},
			})

			Convey("Then no error is returned and task states are updated", func() {
				So(err, ShouldBeNil)
				So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
				So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.StateRejected)
			})
		})
	})

	Convey("Given a static dataset job executor with zebedee collection cleanup during revert", t, func() {
		jobTaskCalls := 0
		mockJobService := &applicationMocks.JobServiceMock{
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 2, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				jobTaskCalls++
				if offset == 0 {
					return []*domain.Task{
						{
							ID:   "series-task",
							Type: domain.TaskTypeDatasetSeries,
							Source: &domain.TaskMetadata{
								ID: "/datasets/my-dataset",
							},
						},
					}, 2, nil
				}

				return []*domain.Task{
					{
						ID:   "version-task",
						Type: domain.TaskTypeDatasetVersion,
						Source: &domain.TaskMetadata{
							ID: "/datasets/my-dataset/editions/time-series/versions/1",
						},
					},
				}, 2, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error { return nil },
		}

		mockZebedeeClient := &clientMocks.ZebedeeClientMock{
			DeleteCollectionContentFunc: func(ctx context.Context, userAuthToken, collectionID, path string) error {
				return errors.New("404 not found")
			},
			DeleteCollectionFunc: func(ctx context.Context, userAuthToken, collectionID string) error {
				return nil
			},
		}

		executor := NewStaticDatasetJobExecutor(
			mockJobService,
			&clients.ClientList{
				Zebedee: mockZebedeeClient,
			},
			"faketoken",
		)

		Convey("When revert is called for a job with a collection ID", func() {
			err := executor.Revert(context.Background(), &domain.Job{
				JobNumber: testJobNumber,
				Config: &domain.JobConfig{
					TargetID:     "target-dataset-id",
					CollectionID: testCollectionID,
				},
			})

			Convey("Then no error is returned and zebedee cleanup methods are called", func() {
				So(err, ShouldBeNil)
				So(jobTaskCalls, ShouldEqual, 3)
				So(len(mockZebedeeClient.DeleteCollectionContentCalls()), ShouldEqual, 2)
				So(mockZebedeeClient.DeleteCollectionContentCalls()[0].CollectionID, ShouldEqual, testCollectionID)
				So(mockZebedeeClient.DeleteCollectionContentCalls()[0].Path, ShouldEqual, "/datasets/my-dataset")
				So(mockZebedeeClient.DeleteCollectionContentCalls()[1].CollectionID, ShouldEqual, testCollectionID)
				So(mockZebedeeClient.DeleteCollectionContentCalls()[1].Path, ShouldEqual, "/datasets/my-dataset/editions/time-series/versions/1")
				So(len(mockZebedeeClient.DeleteCollectionCalls()), ShouldEqual, 1)
				So(mockZebedeeClient.DeleteCollectionCalls()[0].CollectionID, ShouldEqual, testCollectionID)
			})
		})
	})
}
