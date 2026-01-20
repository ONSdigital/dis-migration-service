package executor

import (
	"context"
	"errors"
	"strconv"
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

var (
	testVersionTaskID = "task-1"
	testVersionID     = "1"
	testVersionTask   = &domain.Task{
		ID:        testVersionTaskID,
		JobNumber: testJobNumber,
		Source: &domain.TaskMetadata{
			ID: testEditionURI,
		},
		Target: &domain.TaskMetadata{
			DatasetID: testDatasetSeriesID,
			EditionID: testEditionID,
		},
	}
)

func generateFileName(n int) string {
	return "file" + strconv.Itoa(n) + ".csv"
}

func TestDatasetVersionTaskExecutor(t *testing.T) {
	Convey("Given a job service and datasetAPI clients that don't error", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error { return nil },
			UpdateTaskFunc:      func(ctx context.Context, task *domain.Task) error { return nil },
		}

		mockDatasetClient := &datasetSDKMock.ClienterMock{
			PostVersionFunc: func(ctx context.Context, headers sdk.Headers, datasetID, editionID, versionID string, version models.Version) (*models.Version, error) {
				return &models.Version{}, nil
			},
		}

		Convey("And a zebedee client that returns a dataset version with no files and no previous versions", func() {
			mockClientList := &clients.ClientList{
				DatasetAPI: mockDatasetClient,
				Zebedee: &clientMocks.ZebedeeClientMock{
					GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
						return zebedee.Dataset{
							Type:     zebedee.PageTypeDataset,
							URI:      testEditionURI,
							Versions: []zebedee.Version{},
						}, nil
					},
					GetDatasetLandingPageFunc: func(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.DatasetLandingPage, error) {
						return zebedee.DatasetLandingPage{
							Type: zebedee.PageTypeDatasetLandingPage,
						}, nil
					},
				},
			}

			ctx := context.Background()

			executor := NewDatasetVersionTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

			Convey("When migrate is called for a task", func() {
				err := executor.Migrate(ctx, testVersionTask)

				Convey("Then no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And the datasetAPI is called to create a version", func() {
						So(len(mockDatasetClient.PostVersionCalls()), ShouldEqual, 1)
						So(mockDatasetClient.PostVersionCalls()[0].DatasetID, ShouldEqual, testDatasetSeriesID)
						So(mockDatasetClient.PostVersionCalls()[0].EditionID, ShouldEqual, testEditionID)
						So(mockDatasetClient.PostVersionCalls()[0].VersionID, ShouldEqual, "1")
						So(mockDatasetClient.PostVersionCalls()[0].Headers.AccessToken, ShouldEqual, testServiceAuthToken)

						Convey("And no download tasks are created", func() {
							So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)

							Convey("And the task state is updated to InReview", func() {
								So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
								So(mockJobService.UpdateTaskStateCalls()[0].TaskID, ShouldEqual, testVersionTask.ID)
								So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.StateInReview)
							})
						})
					})
				})
			})
		})
		Convey("And a zebedee client that returns a dataset with multiple download files", func() {
			mockClientList := &clients.ClientList{
				DatasetAPI: mockDatasetClient,
				Zebedee: &clientMocks.ZebedeeClientMock{
					GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
						return zebedee.Dataset{
							Type: zebedee.PageTypeDataset,
							URI:  testEditionURI,
							Downloads: []zebedee.Download{
								{
									File: generateFileName(1),
								},
								{
									File: generateFileName(2),
								},
							},
						}, nil
					},
					GetDatasetLandingPageFunc: func(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.DatasetLandingPage, error) {
						return zebedee.DatasetLandingPage{
							Type: zebedee.PageTypeDatasetLandingPage,
						}, nil
					},
				},
			}

			ctx := context.Background()

			executor := NewDatasetVersionTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

			Convey("When migrate is called for a task", func() {
				err := executor.Migrate(ctx, testVersionTask)

				Convey("Then no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And version tasks are created", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 2)
						So(mockJobService.CreateTaskCalls()[0].Task.Type, ShouldEqual, domain.TaskTypeDatasetDownload)
						So(mockJobService.CreateTaskCalls()[0].Task.Source.ID, ShouldEqual, testEditionURI+"/"+generateFileName(1))
						So(mockJobService.CreateTaskCalls()[0].Task.Target.DatasetID, ShouldEqual, testDatasetSeriesID)
						So(mockJobService.CreateTaskCalls()[0].Task.Target.EditionID, ShouldEqual, testEditionID)
						So(mockJobService.CreateTaskCalls()[0].Task.Target.VersionID, ShouldEqual, "1")

						So(mockJobService.CreateTaskCalls()[1].Task.Type, ShouldEqual, domain.TaskTypeDatasetDownload)
						So(mockJobService.CreateTaskCalls()[1].Task.Source.ID, ShouldEqual, testEditionURI+"/"+generateFileName(2))
						So(mockJobService.CreateTaskCalls()[1].Task.Target.DatasetID, ShouldEqual, testDatasetSeriesID)
						So(mockJobService.CreateTaskCalls()[1].Task.Target.EditionID, ShouldEqual, testEditionID)
						So(mockJobService.CreateTaskCalls()[1].Task.Target.VersionID, ShouldEqual, "1")

						Convey("And the task is updated", func() {
							So(len(mockJobService.UpdateTaskCalls()), ShouldEqual, 1)
							So(mockJobService.UpdateTaskCalls()[0].Task.Target.ID, ShouldEqual, "1")

							So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
							So(mockJobService.UpdateTaskStateCalls()[0].TaskID, ShouldEqual, testVersionTask.ID)
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
				GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
					return zebedee.Dataset{}, errors.New("unknown error")
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetVersionTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			task := testVersionTask

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the datasetAPI should not be called to create a dataset", func() {
					So(len(mockDatasetClient.PostVersionCalls()), ShouldEqual, 0)

					Convey("And no edition tasks are created", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a dataset version task executor with a zebedee client that returns a non dataset page", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{}
		mockDatasetClient := &datasetSDKMock.ClienterMock{}
		mockClientList := &clients.ClientList{
			DatasetAPI: mockDatasetClient,
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: "not_a_dataset_page",
					}, nil
				},
				GetDatasetLandingPageFunc: func(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: zebedee.PageTypeDatasetLandingPage,
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetVersionTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			task := testVersionTask

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And the datasetAPI should not be called to create a dataset", func() {
					So(len(mockDatasetClient.PostVersionCalls()), ShouldEqual, 0)
					Convey("And no edition tasks are created", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a dataset version task executor and a dataset API client that fails to create a version", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			UpdateTaskFunc: func(ctx context.Context, task *domain.Task) error {
				return nil
			},
		}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{
				PostVersionFunc: func(ctx context.Context, headers sdk.Headers, datasetID, editionID, versionID string, version models.Version) (*models.Version, error) {
					return &models.Version{}, errors.New("failed to create version")
				},
			},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: zebedee.PageTypeDatasetLandingPage,
						Datasets: []zebedee.Link{
							{
								URI: "/datasets/test-dataset/editions/2021/versions/1",
							},
						},
					}, nil
				},
				GetDatasetFunc: func(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: zebedee.PageTypeDataset,
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetVersionTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			task := testVersionTask

			err := executor.Migrate(ctx, task)

			Convey("Then the task is updated with the target ID", func() {
				So(len(mockJobService.UpdateTaskCalls()), ShouldEqual, 1)
				So(mockJobService.UpdateTaskCalls()[0].Task.Target.ID, ShouldEqual, "1")

				Convey("And an error is returned", func() {
					So(err, ShouldNotBeNil)

					Convey("And no download tasks are created for the dataset", func() {
						So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a dataset version task executor and a jobService that fails to update a task", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskFunc: func(ctx context.Context, task *domain.Task) error { return nil },
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error {
				return errors.New("failed to update task")
			},
		}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{
				PostVersionFunc: func(ctx context.Context, headers sdk.Headers, datasetID, editionID, versionID string, version models.Version) (*models.Version, error) {
					return &models.Version{}, nil
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
				GetDatasetFunc: func(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: zebedee.PageTypeDataset,
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetVersionTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			task := testVersionTask

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a dataset series task executor and a jobService that fails to create a download task", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return nil, errors.New("failed to create task")
			},
			UpdateTaskFunc:      func(ctx context.Context, task *domain.Task) error { return nil },
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error { return nil },
		}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{
				PostVersionFunc: func(ctx context.Context, headers sdk.Headers, datasetID, editionID, versionID string, version models.Version) (*models.Version, error) {
					return &models.Version{}, nil
				},
			},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetLandingPageFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.DatasetLandingPage, error) {
					return zebedee.DatasetLandingPage{
						Type: zebedee.PageTypeDatasetLandingPage,
					}, nil
				},
				GetDatasetFunc: func(ctx context.Context, userAccessToken, collectionID, lang, path string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: zebedee.PageTypeDataset,
						Downloads: []zebedee.Download{
							{
								File: generateFileName(1),
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetVersionTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			task := testVersionTask

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And no further tasks are created for the version", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 1)
				})
			})
		})
	})
}
