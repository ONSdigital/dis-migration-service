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
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetSDK "github.com/ONSdigital/dp-dataset-api/sdk"
	datasetSDKMocks "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	filesSDK "github.com/ONSdigital/dp-files-api/sdk"
	filesSDKMocks "github.com/ONSdigital/dp-files-api/sdk/mocks"
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
		}
		mockZebedeeClient := &clientMocks.ZebedeeClientMock{
			CreateCollectionFunc: func(ctx context.Context, userAuthToken string, collection zebedee.Collection) (zebedee.Collection, error) {
				collection.ID = testCollectionID
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
					So(mockZebedeeClient.CreateCollectionCalls()[0].Collection.Type, ShouldEqual, "manual")
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

	Convey("Given a static dataset job executor with download tasks for reversion", t, func() {
		originalDeleteDatasetFromAPI := deleteDatasetFromAPI
		defer func() { deleteDatasetFromAPI = originalDeleteDatasetFromAPI }()

		deletedDatasetID := ""
		deleteDatasetFromAPI = func(ctx context.Context, datasetAPIURL, datasetID, serviceAuthToken string) error {
			deletedDatasetID = datasetID
			return nil
		}

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
		}

		mockDatasetClient := &datasetSDKMocks.ClienterMock{
			URLFunc: func() string {
				return testDatasetAPIURL
			},
			GetVersionWithHeadersFunc: func(ctx context.Context, headers datasetSDK.Headers, datasetID, edition, version string) (datasetModels.Version, datasetSDK.ResponseHeaders, error) {
				return datasetModels.Version{
					Distributions: &[]datasetModels.Distribution{
						{Title: "file.csv", DownloadURL: "/uploads/file.csv"},
					},
				}, datasetSDK.ResponseHeaders{ETag: "test-etag"}, nil
			},
			PutVersionFunc: func(ctx context.Context, headers datasetSDK.Headers, datasetID, editionID, versionID string, version datasetModels.Version) (datasetModels.Version, error) {
				return datasetModels.Version{}, nil
			},
		}

		mockFilesClient := &filesSDKMocks.ClienterMock{
			DeleteFileFunc: func(ctx context.Context, filePath string, headers filesSDK.Headers) error {
				return nil
			},
		}

		executor := NewStaticDatasetJobExecutor(
			mockJobService,
			&clients.ClientList{DatasetAPI: mockDatasetClient, FilesAPI: mockFilesClient},
			"faketoken",
		)

		Convey("When revert is called for a job", func() {
			err := executor.Revert(context.Background(), &domain.Job{
				JobNumber: testJobNumber,
				Config: &domain.JobConfig{
					TargetID: "target-dataset-id",
				},
			})

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
				So(len(mockDatasetClient.GetVersionWithHeadersCalls()), ShouldEqual, 1)
				So(len(mockDatasetClient.PutVersionCalls()), ShouldEqual, 1)
				So(len(mockFilesClient.DeleteFileCalls()), ShouldEqual, 1)
				So(deletedDatasetID, ShouldEqual, "target-dataset-id")
			})
		})
	})

	Convey("Given a static dataset job executor and dataset deletion failure during revert", t, func() {
		originalDeleteDatasetFromAPI := deleteDatasetFromAPI
		defer func() { deleteDatasetFromAPI = originalDeleteDatasetFromAPI }()

		deleteDatasetFromAPI = func(ctx context.Context, datasetAPIURL, datasetID, serviceAuthToken string) error {
			return errTest
		}

		mockJobService := &applicationMocks.JobServiceMock{
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 0, nil
			},
		}

		executor := NewStaticDatasetJobExecutor(
			mockJobService,
			&clients.ClientList{DatasetAPI: &datasetSDKMocks.ClienterMock{URLFunc: func() string {
				return testDatasetAPIURL
			}}},
			"faketoken",
		)

		Convey("When revert is called for a job", func() {
			err := executor.Revert(context.Background(), &domain.Job{
				JobNumber: testJobNumber,
				Config: &domain.JobConfig{
					TargetID: "target-dataset-id",
				},
			})

			Convey("Then an error is returned", func() {
				So(err, ShouldEqual, errTest)
			})
		})
	})

	Convey("Given a static dataset job executor with zebedee collection cleanup during revert", t, func() {
		originalDeleteDatasetFromAPI := deleteDatasetFromAPI
		defer func() { deleteDatasetFromAPI = originalDeleteDatasetFromAPI }()

		deleteDatasetFromAPI = func(ctx context.Context, datasetAPIURL, datasetID, serviceAuthToken string) error {
			return nil
		}

		mockJobService := &applicationMocks.JobServiceMock{
			CountTasksByJobNumberFunc: func(ctx context.Context, jobNumber int) (int, error) {
				return 2, nil
			},
			GetJobTasksFunc: func(ctx context.Context, states []domain.State, jobNumber int, limit, offset int) ([]*domain.Task, int, error) {
				return []*domain.Task{
					{
						ID:   "series-task",
						Type: domain.TaskTypeDatasetSeries,
						Source: &domain.TaskMetadata{
							ID: "/datasets/my-dataset",
						},
					},
					{
						ID:   "version-task",
						Type: domain.TaskTypeDatasetVersion,
						Source: &domain.TaskMetadata{
							ID: "/datasets/my-dataset/editions/time-series/versions/1",
						},
					},
				}, 2, nil
			},
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
				DatasetAPI: &datasetSDKMocks.ClienterMock{URLFunc: func() string {
					return testDatasetAPIURL
				}},
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
