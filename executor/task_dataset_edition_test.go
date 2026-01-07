package executor

import (
	"context"
	"testing"

	applicationMocks "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/clients"
	clientMocks "github.com/ONSdigital/dis-migration-service/clients/mock"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDatasetEditionTaskExecutor(t *testing.T) {
	Convey("Given a dataset edition task executor with a zebedee client mock that returns a dataset with no versions", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobID string, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.TaskState) error { return nil },
			UpdateTaskFunc:      func(ctx context.Context, task *domain.Task) error { return nil },
		}

		mockClientList := &clients.ClientList{
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: "dataset",
						URI:  "/source-dataset-id/source-edition-id",
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetEditionTaskExecutor(mockJobService, mockClientList)

		Convey("When migrate is called for a task", func() {
			task := &domain.Task{
				ID:    "task-1",
				JobID: "job-1",
				Source: &domain.TaskMetadata{
					ID: "/source-dataset-id",
				},
				Target: &domain.TaskMetadata{
					DatasetID: "target-dataset-id",
				},
			}

			err := executor.Migrate(ctx, task)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And a version task is created", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 1)
					So(mockJobService.CreateTaskCalls()[0].Task.Type, ShouldEqual, domain.TaskTypeDatasetVersion)
					So(mockJobService.CreateTaskCalls()[0].Task.Source.ID, ShouldEqual, "/source-dataset-id/source-edition-id")
					So(mockJobService.CreateTaskCalls()[0].Task.Target.DatasetID, ShouldEqual, "target-dataset-id")
					So(mockJobService.CreateTaskCalls()[0].Task.Target.EditionID, ShouldEqual, "source-edition-id")

					Convey("And the task is updated", func() {
						So(len(mockJobService.UpdateTaskCalls()), ShouldEqual, 1)
						So(mockJobService.UpdateTaskCalls()[0].Task.Target.ID, ShouldEqual, "source-edition-id")

						So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
						So(mockJobService.UpdateTaskStateCalls()[0].TaskID, ShouldEqual, task.ID)
						So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.TaskStateInReview)
					})
				})
			})
		})
	})

	Convey("Given a dataset edition task executor with a zebedee client mock that returns a dataset with multiple versions", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobID string, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.TaskState) error { return nil },
			UpdateTaskFunc:      func(ctx context.Context, task *domain.Task) error { return nil },
		}

		mockClientList := &clients.ClientList{
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: "dataset",
						URI:  "/source-dataset-id/source-edition-id",
						Versions: []zebedee.Version{
							{
								URI: "/source-dataset-id/source-edition-id/previous/v1",
							},
							{
								URI: "/source-dataset-id/source-edition-id/previous/v2",
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetEditionTaskExecutor(mockJobService, mockClientList)

		Convey("When migrate is called for a task", func() {
			task := &domain.Task{
				ID:    "task-1",
				JobID: "job-1",
				Source: &domain.TaskMetadata{
					ID: "/source-dataset-id",
				},
				Target: &domain.TaskMetadata{
					DatasetID: "target-dataset-id",
				},
			}

			err := executor.Migrate(ctx, task)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And version tasks are created", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 3)
					So(mockJobService.CreateTaskCalls()[0].Task.Type, ShouldEqual, domain.TaskTypeDatasetVersion)
					So(mockJobService.CreateTaskCalls()[0].Task.Source.ID, ShouldEqual, "/source-dataset-id/source-edition-id")
					So(mockJobService.CreateTaskCalls()[0].Task.Target.DatasetID, ShouldEqual, "target-dataset-id")
					So(mockJobService.CreateTaskCalls()[0].Task.Target.EditionID, ShouldEqual, "source-edition-id")

					So(mockJobService.CreateTaskCalls()[1].Task.Type, ShouldEqual, domain.TaskTypeDatasetVersion)
					So(mockJobService.CreateTaskCalls()[1].Task.Source.ID, ShouldEqual, "/source-dataset-id/source-edition-id/previous/v1")
					So(mockJobService.CreateTaskCalls()[1].Task.Target.DatasetID, ShouldEqual, "target-dataset-id")
					So(mockJobService.CreateTaskCalls()[1].Task.Target.EditionID, ShouldEqual, "source-edition-id")

					So(mockJobService.CreateTaskCalls()[2].Task.Type, ShouldEqual, domain.TaskTypeDatasetVersion)
					So(mockJobService.CreateTaskCalls()[2].Task.Source.ID, ShouldEqual, "/source-dataset-id/source-edition-id/previous/v2")
					So(mockJobService.CreateTaskCalls()[2].Task.Target.DatasetID, ShouldEqual, "target-dataset-id")
					So(mockJobService.CreateTaskCalls()[2].Task.Target.EditionID, ShouldEqual, "source-edition-id")

					Convey("And the task is updated", func() {
						So(len(mockJobService.UpdateTaskCalls()), ShouldEqual, 1)
						So(mockJobService.UpdateTaskCalls()[0].Task.Target.ID, ShouldEqual, "source-edition-id")

						So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
						So(mockJobService.UpdateTaskStateCalls()[0].TaskID, ShouldEqual, task.ID)
						So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.TaskStateInReview)
					})
				})
			})
		})
	})

	// Convey("Given a dataset series task executor with a zebedee client mock errors", t, func() {
	// 	mockJobService := &applicationMocks.JobServiceMock{}
	// 	mockClientList := &clients.ClientList{
	// 		DatasetAPI: &datasetSDKMock.ClienterMock{},
	// 		Zebedee: &clientMocks.ZebedeeClientMock{
	// 			GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
	// 				return zebedee.DatasetLandingPage{}, errors.New("unknown error")
	// 			},
	// 		},
	// 	}

	// 	ctx := context.Background()

	// 	executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

	// 	Convey("When migrate is called for a task", func() {
	// 		task := &domain.Task{
	// 			ID:    "task-1",
	// 			JobID: "job-1",
	// 			Source: &domain.TaskMetadata{
	// 				ID: "source-dataset-id",
	// 			},
	// 			Target: &domain.TaskMetadata{
	// 				ID: "target-dataset-id",
	// 			},
	// 		}

	// 		err := executor.Migrate(ctx, task)

	// 		Convey("Then an error is returned", func() {
	// 			So(err, ShouldNotBeNil)

	// 			Convey("And the datasetAPI should not be called to create a dataset", func() {
	// 				// TODO: implement dataset creation check

	// 				Convey("And no edition tasks are created", func() {
	// 					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
	// 				})
	// 			})
	// 		})
	// 	})
	// })

	// Convey("Given a dataset series task executor with a zebedee client that returns a non dataset landing page", t, func() {
	// 	mockJobService := &applicationMocks.JobServiceMock{}
	// 	mockClientList := &clients.ClientList{
	// 		DatasetAPI: &datasetSDKMock.ClienterMock{},
	// 		Zebedee: &clientMocks.ZebedeeClientMock{
	// 			GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
	// 				return zebedee.DatasetLandingPage{
	// 					Type: "not_a_dataset_landing_page",
	// 				}, nil
	// 			},
	// 		},
	// 	}

	// 	ctx := context.Background()

	// 	executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

	// 	Convey("When migrate is called for a task", func() {
	// 		task := &domain.Task{
	// 			ID:    "task-1",
	// 			JobID: "job-1",
	// 			Source: &domain.TaskMetadata{
	// 				ID: "source-dataset-id",
	// 			},
	// 			Target: &domain.TaskMetadata{
	// 				ID: "target-dataset-id",
	// 			},
	// 		}

	// 		err := executor.Migrate(ctx, task)

	// 		Convey("Then an error is returned", func() {
	// 			So(err, ShouldNotBeNil)

	// 			Convey("And the datasetAPI should not be called to create a dataset", func() {
	// 				// TODO: implement dataset creation check
	// 				Convey("And no edition tasks are created", func() {
	// 					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
	// 				})
	// 			})
	// 		})
	// 	})
	// })

	// Convey("Given a dataset series task executor and a dataset API client that fails to create a dataset", t, func() {
	// 	mockJobService := &applicationMocks.JobServiceMock{}
	// 	mockClientList := &clients.ClientList{
	// 		DatasetAPI: &datasetSDKMock.ClienterMock{
	// 			CreateDatasetFunc: func(ctx context.Context, headers sdk.Headers, dataset models.Dataset) (models.DatasetUpdate, error) {
	// 				return models.DatasetUpdate{}, errors.New("failed to create dataset")
	// 			},
	// 		},
	// 		Zebedee: &clientMocks.ZebedeeClientMock{
	// 			GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
	// 				return zebedee.DatasetLandingPage{
	// 					Type: "dataset_landing_page",
	// 					Datasets: []zebedee.Link{
	// 						{
	// 							URI: "/datasets/test-dataset/editions/2021/versions/1",
	// 						},
	// 					},
	// 				}, nil
	// 			},
	// 		},
	// 	}

	// 	ctx := context.Background()

	// 	executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

	// 	Convey("When migrate is called for a task", func() {
	// 		task := &domain.Task{
	// 			ID:    "task-1",
	// 			JobID: "job-1",
	// 			Source: &domain.TaskMetadata{
	// 				ID: "source-dataset-id",
	// 			},
	// 			Target: &domain.TaskMetadata{
	// 				ID: "target-dataset-id",
	// 			},
	// 		}

	// 		err := executor.Migrate(ctx, task)

	// 		Convey("Then an error is returned", func() {
	// 			So(err, ShouldNotBeNil)

	// 			Convey("And no edition tasks are created for the dataset", func() {
	// 				So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
	// 			})
	// 		})
	// 	})
	// })

	// Convey("Given a dataset series task executor and a jobService that fails to update a task", t, func() {
	// 	mockJobService := &applicationMocks.JobServiceMock{
	// 		CreateTaskFunc: func(ctx context.Context, jobID string, task *domain.Task) (*domain.Task, error) {
	// 			return &domain.Task{}, nil
	// 		},
	// 		UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.TaskState) error {
	// 			return errors.New("failed to update task")
	// 		},
	// 	}
	// 	mockClientList := &clients.ClientList{
	// 		DatasetAPI: &datasetSDKMock.ClienterMock{
	// 			CreateDatasetFunc: func(ctx context.Context, headers sdk.Headers, dataset models.Dataset) (models.DatasetUpdate, error) {
	// 				return models.DatasetUpdate{}, nil
	// 			},
	// 		},
	// 		Zebedee: &clientMocks.ZebedeeClientMock{
	// 			GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
	// 				return zebedee.DatasetLandingPage{
	// 					Type: "dataset_landing_page",
	// 					Datasets: []zebedee.Link{
	// 						{
	// 							URI: "/datasets/test-dataset/editions/2021/versions/1",
	// 						},
	// 					},
	// 				}, nil
	// 			},
	// 		},
	// 	}

	// 	ctx := context.Background()

	// 	executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

	// 	Convey("When migrate is called for a task", func() {
	// 		task := &domain.Task{
	// 			ID:    "task-1",
	// 			JobID: "job-1",
	// 			Source: &domain.TaskMetadata{
	// 				ID: "source-dataset-id",
	// 			},
	// 			Target: &domain.TaskMetadata{
	// 				ID: "target-dataset-id",
	// 			},
	// 		}

	// 		err := executor.Migrate(ctx, task)

	// 		Convey("Then an error is returned", func() {
	// 			So(err, ShouldNotBeNil)
	// 		})
	// 	})
	// })

	// Convey("Given a dataset series task executor and a jobService that fails to create an edition task", t, func() {
	// 	mockJobService := &applicationMocks.JobServiceMock{
	// 		CreateTaskFunc: func(ctx context.Context, jobID string, task *domain.Task) (*domain.Task, error) {
	// 			return nil, errors.New("failed to create task")
	// 		},
	// 		UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.TaskState) error { return nil },
	// 	}
	// 	mockClientList := &clients.ClientList{
	// 		DatasetAPI: &datasetSDKMock.ClienterMock{
	// 			CreateDatasetFunc: func(ctx context.Context, headers sdk.Headers, dataset models.Dataset) (models.DatasetUpdate, error) {
	// 				return models.DatasetUpdate{}, nil
	// 			},
	// 		},
	// 		Zebedee: &clientMocks.ZebedeeClientMock{
	// 			GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
	// 				return zebedee.DatasetLandingPage{
	// 					Type: "dataset_landing_page",
	// 					Datasets: []zebedee.Link{
	// 						{
	// 							URI: "/datasets/test-dataset/editions/2021/versions/1",
	// 						},
	// 						{
	// 							URI: "/datasets/test-dataset/editions/2022/versions/1",
	// 						},
	// 					},
	// 				}, nil
	// 			},
	// 		},
	// 	}

	// 	ctx := context.Background()

	// 	executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

	// 	Convey("When migrate is called for a task", func() {
	// 		task := &domain.Task{
	// 			ID:    "task-1",
	// 			JobID: "job-1",
	// 			Source: &domain.TaskMetadata{
	// 				ID: "source-dataset-id",
	// 			},
	// 			Target: &domain.TaskMetadata{
	// 				ID: "target-dataset-id",
	// 			},
	// 		}

	// 		err := executor.Migrate(ctx, task)

	// 		Convey("Then an error is returned", func() {
	// 			So(err, ShouldNotBeNil)

	// 			Convey("And no further edition tasks are created for the dataset", func() {
	// 				So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 1)
	// 			})
	// 		})
	// 	})
	// })
}
