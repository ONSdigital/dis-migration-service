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
	"github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/sdk"
	datasetSDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testServiceAuthToken = "test-service-auth-token"
	testDatasetSeriesID  = "target-dataset-id"
	testDatasetSeriesURI = "/source-dataset-uri"
	testSeriesTaskID     = "task-1"
)

var (
	testSeriesTask = &domain.Task{
		ID:        testSeriesTaskID,
		JobNumber: testJobNumber,
		Source: &domain.TaskMetadata{
			ID: testDatasetSeriesURI,
		},
		Target: &domain.TaskMetadata{
			ID: testDatasetSeriesID,
		},
	}
)

func getEditionURI(base, edition string) string {
	return base + "/" + edition
}

func TestDatasetSeriesTaskExecutor(t *testing.T) {
	Convey("Given a dataset series task executor with a zebedee client mock that returns a dataset series and a dataset API client mock that creates datasets", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error { return nil },
		}
		mockDatasetClient := &datasetSDKMock.ClienterMock{
			CreateDatasetFunc: func(ctx context.Context, headers sdk.Headers, dataset models.Dataset) (models.DatasetUpdate, error) {
				return models.DatasetUpdate{}, nil
			},
		}

		mockClientList := &clients.ClientList{
			DatasetAPI: mockDatasetClient,
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: zebedee.PageTypeDatasetLandingPage,
						Datasets: []zebedee.Link{
							{
								URI: getEditionURI(testDatasetSeriesURI, "2021"),
							},
							{
								URI: getEditionURI(testDatasetSeriesURI, "2022"),
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			err := executor.Migrate(ctx, testSeriesTask)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And the datasetAPI is called to create a dataset", func() {
					So(len(mockDatasetClient.CreateDatasetCalls()), ShouldEqual, 1)
					So(mockDatasetClient.CreateDatasetCalls()[0].Dataset.ID, ShouldEqual, testDatasetSeriesID)
					So(mockDatasetClient.CreateDatasetCalls()[0].Headers.AccessToken, ShouldEqual, testServiceAuthToken)

					Convey("And an edition task is created for each dataset", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 2)
						So(mockJobService.CreateTaskCalls()[0].Task.Type, ShouldEqual, domain.TaskTypeDatasetEdition)
						So(mockJobService.CreateTaskCalls()[0].Task.Source.ID, ShouldEqual, getEditionURI(testDatasetSeriesURI, "2021"))
						So(mockJobService.CreateTaskCalls()[0].Task.Target.DatasetID, ShouldEqual, testDatasetSeriesID)
						So(mockJobService.CreateTaskCalls()[1].Task.Type, ShouldEqual, domain.TaskTypeDatasetEdition)
						So(mockJobService.CreateTaskCalls()[1].Task.Source.ID, ShouldEqual, getEditionURI(testDatasetSeriesURI, "2022"))
						So(mockJobService.CreateTaskCalls()[1].Task.Target.DatasetID, ShouldEqual, testDatasetSeriesID)

						Convey("And the task state is updated to InReview", func() {
							So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
							So(mockJobService.UpdateTaskStateCalls()[0].TaskID, ShouldEqual, testSeriesTask.ID)
							So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.StateInReview)
						})
					})
				})
			})
		})
	})

	Convey("Given a dataset series task executor with a zebedee client mock errors", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{}
		mockDatasetClient := &datasetSDKMock.ClienterMock{}
		mockClientList := &clients.ClientList{
			DatasetAPI: mockDatasetClient,
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{}, errors.New("unknown error")
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			err := executor.Migrate(ctx, testSeriesTask)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the datasetAPI should not be called to create a dataset", func() {
					So(len(mockDatasetClient.CreateDatasetCalls()), ShouldEqual, 0)

					Convey("And no edition tasks are created", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a dataset series task executor with a zebedee client that returns a non dataset landing page", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{}
		mockDatasetClient := &datasetSDKMock.ClienterMock{}
		mockClientList := &clients.ClientList{
			DatasetAPI: mockDatasetClient,
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: "not_a_dataset_landing_page",
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			err := executor.Migrate(ctx, testSeriesTask)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the datasetAPI should not be called to create a dataset", func() {
					So(len(mockDatasetClient.CreateDatasetCalls()), ShouldEqual, 0)

					Convey("And no edition tasks are created", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a dataset series task executor and a dataset API client that fails to create a dataset", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{
				CreateDatasetFunc: func(ctx context.Context, headers sdk.Headers, dataset models.Dataset) (models.DatasetUpdate, error) {
					return models.DatasetUpdate{}, errors.New("failed to create dataset")
				},
			},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: "dataset_landing_page",
						Datasets: []zebedee.Link{
							{
								URI: "/datasets/test-dataset/editions/2021/versions/1",
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			err := executor.Migrate(ctx, testSeriesTask)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And no edition tasks are created for the dataset", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a dataset series task executor and a jobService that fails to update a task", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error {
				return errors.New("failed to update task")
			},
		}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{
				CreateDatasetFunc: func(ctx context.Context, headers sdk.Headers, dataset models.Dataset) (models.DatasetUpdate, error) {
					return models.DatasetUpdate{}, nil
				},
			},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: zebedee.PageTypeDatasetLandingPage,
						Datasets: []zebedee.Link{
							{
								URI: getEditionURI(testDatasetSeriesURI, "2021"),
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			err := executor.Migrate(ctx, testSeriesTask)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a dataset series task executor and a jobService that fails to create an edition task", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return nil, errors.New("failed to create task")
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error { return nil },
		}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{
				CreateDatasetFunc: func(ctx context.Context, headers sdk.Headers, dataset models.Dataset) (models.DatasetUpdate, error) {
					return models.DatasetUpdate{}, nil
				},
			},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: zebedee.PageTypeDatasetLandingPage,
						Datasets: []zebedee.Link{
							{
								URI: getEditionURI(testDatasetSeriesURI, "2021"),
							},
							{
								URI: getEditionURI(testDatasetSeriesURI, "2022"),
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetSeriesTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			err := executor.Migrate(ctx, testSeriesTask)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And no further edition tasks are created for the dataset", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 1)
				})
			})
		})
	})
}
