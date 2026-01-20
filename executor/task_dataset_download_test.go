package executor

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	applicationMocks "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/clients"
	clientMocks "github.com/ONSdigital/dis-migration-service/clients/mock"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	"github.com/ONSdigital/dp-dataset-api/sdk"
	datasetSDKMock "github.com/ONSdigital/dp-dataset-api/sdk/mocks"
	"github.com/ONSdigital/dp-upload-service/api"
	uploadSDK "github.com/ONSdigital/dp-upload-service/sdk"
	uploadSDKMock "github.com/ONSdigital/dp-upload-service/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	testDownloadTaskID = "task-1"
	testDownloadURI    = "/source-dataset-id/source-edition-id/file1.csv"

	testFileName = "file1.csv"
	testFileData = `
		col1,col2,col3
		val1,val2,val3
	`

	testDownloadTask = &domain.Task{
		ID:        testDownloadTaskID,
		JobNumber: testJobNumber,
		Source: &domain.TaskMetadata{
			ID: testDownloadURI,
		},
		Target: &domain.TaskMetadata{
			DatasetID: testDatasetSeriesID,
			EditionID: testEditionID,
			VersionID: testVersionID,
		},
	}
)

func TestDatasetDownloadTaskExecutor(t *testing.T) {
	Convey("Given a job service and clients that don't error", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error { return nil },
			UpdateTaskFunc:      func(ctx context.Context, task *domain.Task) error { return nil },
		}

		mockDatasetClient := &datasetSDKMock.ClienterMock{
			GetVersionWithHeadersFunc: func(ctx context.Context, headers sdk.Headers, datasetID, edition, version string) (datasetModels.Version, sdk.ResponseHeaders, error) {
				return datasetModels.Version{
					Distributions: &[]datasetModels.Distribution{},
				}, sdk.ResponseHeaders{}, nil
			},
			PutVersionFunc: func(ctx context.Context, headers sdk.Headers, datasetID, editionID, versionID string, version datasetModels.Version) (datasetModels.Version, error) {
				return datasetModels.Version{}, nil
			},
		}

		mockUploadClient := &uploadSDKMock.ClienterMock{
			UploadFunc: func(ctx context.Context, fileContent io.ReadCloser, metadata api.Metadata, headers uploadSDK.Headers) error {
				return nil
			},
		}

		Convey("And a zebedee client that returns a file stream and size", func() {
			mockClientList := &clients.ClientList{
				DatasetAPI: mockDatasetClient,
				Zebedee: &clientMocks.ZebedeeClientMock{
					GetResourceStreamFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (io.ReadCloser, error) {
						return io.NopCloser(bytes.NewReader([]byte(testFileData))), nil
					},
					GetFileSizeFunc: func(ctx context.Context, userAccessToken, collectionID, lang, uri string) (zebedee.FileSize, error) {
						return zebedee.FileSize{Size: len(testFileData)}, nil
					},
				},
				UploadService: mockUploadClient,
			}

			ctx := context.Background()

			executor := NewDatasetDownloadTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

			Convey("When migrate is called for a download task", func() {
				err := executor.Migrate(ctx, testDownloadTask)

				Convey("Then no error is returned", func() {
					So(err, ShouldBeNil)

					Convey("And the UploadService is called to upload a file", func() {
						So(len(mockUploadClient.UploadCalls()), ShouldEqual, 1)
						So(mockUploadClient.UploadCalls()[0].Metadata.SizeInBytes, ShouldEqual, int64(len(testFileData)))
						So(mockUploadClient.UploadCalls()[0].Metadata.Title, ShouldEqual, testFileName)
						So(mockUploadClient.UploadCalls()[0].Metadata.Type, ShouldEqual, "text/csv")
						So(mockUploadClient.UploadCalls()[0].Metadata.Licence, ShouldEqual, domain.OpenGovernmentLicence)
						So(mockUploadClient.UploadCalls()[0].Metadata.LicenceUrl, ShouldEqual, domain.OpenGovernmentLicenceURL)
						So(mockUploadClient.UploadCalls()[0].Metadata.IsPublishable, ShouldNotBeNil)
						So(*mockUploadClient.UploadCalls()[0].Metadata.IsPublishable, ShouldBeTrue)

						Convey("And the dataset API is called to update the version with the new distribution", func() {
							So(len(mockDatasetClient.PutVersionCalls()), ShouldEqual, 1)
							So(mockDatasetClient.PutVersionCalls()[0].DatasetID, ShouldEqual, testDownloadTask.Target.DatasetID)
							So(mockDatasetClient.PutVersionCalls()[0].EditionID, ShouldEqual, testDownloadTask.Target.EditionID)
							So(mockDatasetClient.PutVersionCalls()[0].VersionID, ShouldEqual, testDownloadTask.Target.VersionID)

							distributions := *mockDatasetClient.PutVersionCalls()[0].Version.Distributions
							distribution := distributions[0]
							So(distribution.ByteSize, ShouldEqual, int64(len(testFileData)))
							So(distribution.Format, ShouldEqual, datasetModels.DistributionFormatCSV)
							So(distribution.MediaType, ShouldEqual, datasetModels.DistributionMediaTypeCSV)
							So(distribution.Title, ShouldEqual, testFileName)
							So(distribution.DownloadURL, ShouldNotBeEmpty)
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

		Convey("Given a dataset download task executor with a zebedee client mock that errors", func() {
			mockJobService := &applicationMocks.JobServiceMock{}
			mockDatasetClient := &datasetSDKMock.ClienterMock{}
			mockClientList := &clients.ClientList{
				DatasetAPI: mockDatasetClient,
				Zebedee: &clientMocks.ZebedeeClientMock{
					GetResourceStreamFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (io.ReadCloser, error) {
						return nil, errors.New("failed to get resource stream")
					},
					GetFileSizeFunc: func(ctx context.Context, userAccessToken, collectionID, lang, uri string) (zebedee.FileSize, error) {
						return zebedee.FileSize{}, nil
					},
				},
				UploadService: mockUploadClient,
			}

			ctx := context.Background()

			executor := NewDatasetDownloadTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

			Convey("When migrate is called for a task", func() {
				task := testDownloadTask

				err := executor.Migrate(ctx, task)

				Convey("Then an error is returned", func() {
					So(err, ShouldNotBeNil)

					Convey("And the upstream services should not be called to upload a file or it's metadata", func() {
						So(len(mockUploadClient.UploadCalls()), ShouldEqual, 0)
						So(len(mockDatasetClient.PutVersionCalls()), ShouldEqual, 0)
					})
				})
			})
		})

		Convey("Given a dataset download task executor and an upload service client that fails to upload a file", func() {
			mockJobService := &applicationMocks.JobServiceMock{
				UpdateTaskFunc: func(ctx context.Context, task *domain.Task) error {
					return nil
				},
			}
			mockUploadClient := &uploadSDKMock.ClienterMock{
				UploadFunc: func(ctx context.Context, fileContent io.ReadCloser, metadata api.Metadata, headers uploadSDK.Headers) error {
					return errors.New("failed to upload file")
				},
			}

			mockDatasetClient := &datasetSDKMock.ClienterMock{}

			mockClientList := &clients.ClientList{
				DatasetAPI: mockDatasetClient,
				Zebedee: &clientMocks.ZebedeeClientMock{
					GetResourceStreamFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (io.ReadCloser, error) {
						return io.NopCloser(bytes.NewReader([]byte(testFileData))), nil
					},
					GetFileSizeFunc: func(ctx context.Context, userAccessToken, collectionID, lang, uri string) (zebedee.FileSize, error) {
						return zebedee.FileSize{Size: len(testFileData)}, nil
					},
				},
				UploadService: mockUploadClient,
			}

			ctx := context.Background()

			executor := NewDatasetDownloadTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

			Convey("When migrate is called for a task", func() {
				task := testDownloadTask

				err := executor.Migrate(ctx, task)

				Convey("Then an error is returned", func() {
					So(err, ShouldNotBeNil)

					Convey("And no version updates are called", func() {
						So(len(mockDatasetClient.PutVersionCalls()), ShouldEqual, 0)
					})
				})
			})
		})
	})

	Convey("Given a dataset download task executor and a jobService that fails to update a task", t, func() {
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
				GetVersionWithHeadersFunc: func(ctx context.Context, headers sdk.Headers, datasetID, edition, version string) (datasetModels.Version, sdk.ResponseHeaders, error) {
					return datasetModels.Version{
						Distributions: &[]datasetModels.Distribution{},
					}, sdk.ResponseHeaders{}, nil
				},
				PutVersionFunc: func(ctx context.Context, headers sdk.Headers, datasetID, editionID, versionID string, version datasetModels.Version) (datasetModels.Version, error) {
					return datasetModels.Version{}, nil
				},
			},
			UploadService: &uploadSDKMock.ClienterMock{
				UploadFunc: func(ctx context.Context, fileContent io.ReadCloser, metadata api.Metadata, headers uploadSDK.Headers) error {
					return nil
				},
			},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetResourceStreamFunc: func(ctx context.Context, userAuthToken, collectionID, lang, path string) (io.ReadCloser, error) {
					return io.NopCloser(bytes.NewReader([]byte(testFileData))), nil
				},
				GetFileSizeFunc: func(ctx context.Context, userAccessToken, collectionID, lang, uri string) (zebedee.FileSize, error) {
					return zebedee.FileSize{Size: len(testFileData)}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetDownloadTaskExecutor(mockJobService, mockClientList, testServiceAuthToken)

		Convey("When migrate is called for a task", func() {
			task := testDownloadTask

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
