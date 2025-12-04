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
	datasetSDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDatasetSeriesTaskExecutor(t *testing.T) {
	Convey("Given a dataset series task executor with a zebedee client mock that returns a dataset series", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobID string, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskFunc: func(ctx context.Context, task *domain.Task) error { return nil },
		}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: "dataset_landing_page",
						Datasets: []zebedee.Related{
							{
								URI: "/datasets/test-dataset/editions/2021/versions/1",
							},
							{
								URI: "/datasets/test-dataset/editions/2022/versions/1",
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(ctx, mockJobService, mockClientList)

		Convey("When migrate is called for a task", func() {
			task := &domain.Task{
				ID:    "task-1",
				JobID: "job-1",
				Source: &domain.TaskMetadata{
					ID: "source-dataset-id",
				},
				Target: &domain.TaskMetadata{
					ID: "target-dataset-id",
				},
			}

			err := executor.Migrate(ctx, task)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And the datasetAPI is called to create a dataset", func() {
					// TODO: implement dataset creation check

					Convey("And an edition task is created for each dataset", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 2)
						So(mockJobService.CreateTaskCalls()[0].Task.Type, ShouldEqual, domain.TaskTypeDatasetEdition)
						So(mockJobService.CreateTaskCalls()[0].Task.Source.ID, ShouldEqual, "/datasets/test-dataset/editions/2021/versions/1")
						So(mockJobService.CreateTaskCalls()[0].Task.Target.DatasetID, ShouldEqual, "target-dataset-id")
						So(mockJobService.CreateTaskCalls()[1].Task.Type, ShouldEqual, domain.TaskTypeDatasetEdition)
						So(mockJobService.CreateTaskCalls()[1].Task.Source.ID, ShouldEqual, "/datasets/test-dataset/editions/2022/versions/1")
						So(mockJobService.CreateTaskCalls()[1].Task.Target.DatasetID, ShouldEqual, "target-dataset-id")

						Convey("And the task state is updated to InReview", func() {
							So(task.State, ShouldEqual, domain.TaskStateInReview)
							So(len(mockJobService.UpdateTaskCalls()), ShouldEqual, 1)
							So(mockJobService.UpdateTaskCalls()[0].Task.State, ShouldEqual, domain.TaskStateInReview)
						})
					})
				})
			})
		})
	})

	Convey("Given a dataset series task executor with a zebedee client mock errors", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{}, errors.New("unknown error")
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(ctx, mockJobService, mockClientList)

		Convey("When migrate is called for a task", func() {
			task := &domain.Task{
				ID:    "task-1",
				JobID: "job-1",
				Source: &domain.TaskMetadata{
					ID: "source-dataset-id",
				},
				Target: &domain.TaskMetadata{
					ID: "target-dataset-id",
				},
			}

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the datasetAPI should not be called to create a dataset", func() {
					// TODO: implement dataset creation check

					Convey("And no edition tasks are created", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a dataset series task executor with a zebedee client that returns a non dataset landing page", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: "not_a_dataset_landing_page",
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(ctx, mockJobService, mockClientList)

		Convey("When migrate is called for a task", func() {
			task := &domain.Task{
				ID:    "task-1",
				JobID: "job-1",
				Source: &domain.TaskMetadata{
					ID: "source-dataset-id",
				},
				Target: &domain.TaskMetadata{
					ID: "target-dataset-id",
				},
			}

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the datasetAPI should not be called to create a dataset", func() {
					// TODO: implement dataset creation check
					Convey("And no edition tasks are created", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a dataset series task executor and a jobService that fails to update a task", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobID string, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskFunc: func(ctx context.Context, task *domain.Task) error {
				return errors.New("failed to update task")
			},
		}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: "dataset_landing_page",
						Datasets: []zebedee.Related{
							{
								URI: "/datasets/test-dataset/editions/2021/versions/1",
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(ctx, mockJobService, mockClientList)

		Convey("When migrate is called for a task", func() {
			task := &domain.Task{
				ID:    "task-1",
				JobID: "job-1",
				Source: &domain.TaskMetadata{
					ID: "source-dataset-id",
				},
				Target: &domain.TaskMetadata{
					ID: "target-dataset-id",
				},
			}

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a dataset series task executor and a jobService that fails to create an edition task", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobID string, task *domain.Task) (*domain.Task, error) {
				return nil, errors.New("failed to create task")
			},
			UpdateTaskFunc: func(ctx context.Context, task *domain.Task) error { return nil },
		}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: "dataset_landing_page",
						Datasets: []zebedee.Related{
							{
								URI: "/datasets/test-dataset/editions/2021/versions/1",
							},
							{
								URI: "/datasets/test-dataset/editions/2022/versions/1",
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(ctx, mockJobService, mockClientList)

		Convey("When migrate is called for a task", func() {
			task := &domain.Task{
				ID:    "task-1",
				JobID: "job-1",
				Source: &domain.TaskMetadata{
					ID: "source-dataset-id",
				},
				Target: &domain.TaskMetadata{
					ID: "target-dataset-id",
				},
			}

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And no further edition tasks are created for the dataset", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 1)
				})
			})
		})
	})
}
