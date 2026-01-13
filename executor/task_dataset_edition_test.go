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

const (
	testEditionURI    = "/source-dataset-id/source-edition-id"
	testEditionID     = "source-edition-id"
	testEditionTaskID = "task-1"
)

var (
	testEditionTask = &domain.Task{
		ID:        testEditionTaskID,
		JobNumber: testJobNumber,
		Source: &domain.TaskMetadata{
			ID: testEditionURI,
		},
		Target: &domain.TaskMetadata{
			DatasetID: testDatasetSeriesID,
		},
	}
)

func generatePreviousVersionURI(baseURI string, versionNumber int) string {
	return baseURI + "/previous/v" + strconv.Itoa(versionNumber)
}

func TestDatasetEditionTaskExecutor(t *testing.T) {
	Convey("Given a dataset edition task executor with a zebedee client mock that returns a dataset with no versions", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error { return nil },
			UpdateTaskFunc:      func(ctx context.Context, task *domain.Task) error { return nil },
		}

		mockClientList := &clients.ClientList{
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: zebedee.PageTypeDataset,
						URI:  testEditionURI,
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetEditionTaskExecutor(mockJobService, mockClientList, "")

		Convey("When migrate is called for a task", func() {
			err := executor.Migrate(ctx, testEditionTask)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And a version task is created", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 1)
					So(mockJobService.CreateTaskCalls()[0].Task.Type, ShouldEqual, domain.TaskTypeDatasetVersion)
					So(mockJobService.CreateTaskCalls()[0].Task.Source.ID, ShouldEqual, testEditionURI)
					So(mockJobService.CreateTaskCalls()[0].Task.Target.DatasetID, ShouldEqual, testDatasetSeriesID)
					So(mockJobService.CreateTaskCalls()[0].Task.Target.EditionID, ShouldEqual, testEditionID)

					Convey("And the task is updated", func() {
						So(len(mockJobService.UpdateTaskCalls()), ShouldEqual, 1)
						So(mockJobService.UpdateTaskCalls()[0].Task.Target.ID, ShouldEqual, testEditionID)
						So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
						So(mockJobService.UpdateTaskStateCalls()[0].TaskID, ShouldEqual, testEditionTask.ID)
						So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.StateInReview)
					})
				})
			})
		})
	})

	Convey("Given a dataset edition task executor with a zebedee client mock that returns a dataset with multiple versions", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error { return nil },
			UpdateTaskFunc:      func(ctx context.Context, task *domain.Task) error { return nil },
		}

		mockClientList := &clients.ClientList{
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: zebedee.PageTypeDataset,
						URI:  testEditionURI,
						Versions: []zebedee.Version{
							{
								URI: generatePreviousVersionURI(testEditionURI, 1),
							},
							{
								URI: generatePreviousVersionURI(testEditionURI, 2),
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetEditionTaskExecutor(mockJobService, mockClientList, "")

		Convey("When migrate is called for a task", func() {
			err := executor.Migrate(ctx, testEditionTask)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And version tasks are created", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 3)
					So(mockJobService.CreateTaskCalls()[0].Task.Type, ShouldEqual, domain.TaskTypeDatasetVersion)
					So(mockJobService.CreateTaskCalls()[0].Task.Source.ID, ShouldEqual, testEditionURI)
					So(mockJobService.CreateTaskCalls()[0].Task.Target.DatasetID, ShouldEqual, testDatasetSeriesID)
					So(mockJobService.CreateTaskCalls()[0].Task.Target.EditionID, ShouldEqual, testEditionID)

					So(mockJobService.CreateTaskCalls()[1].Task.Type, ShouldEqual, domain.TaskTypeDatasetVersion)
					So(mockJobService.CreateTaskCalls()[1].Task.Source.ID, ShouldEqual, generatePreviousVersionURI(testEditionURI, 1))
					So(mockJobService.CreateTaskCalls()[1].Task.Target.DatasetID, ShouldEqual, testDatasetSeriesID)
					So(mockJobService.CreateTaskCalls()[1].Task.Target.EditionID, ShouldEqual, testEditionID)

					So(mockJobService.CreateTaskCalls()[2].Task.Type, ShouldEqual, domain.TaskTypeDatasetVersion)
					So(mockJobService.CreateTaskCalls()[2].Task.Source.ID, ShouldEqual, generatePreviousVersionURI(testEditionURI, 2))
					So(mockJobService.CreateTaskCalls()[2].Task.Target.DatasetID, ShouldEqual, testDatasetSeriesID)
					So(mockJobService.CreateTaskCalls()[2].Task.Target.EditionID, ShouldEqual, testEditionID)

					Convey("And the task is updated", func() {
						So(len(mockJobService.UpdateTaskCalls()), ShouldEqual, 1)
						So(mockJobService.UpdateTaskCalls()[0].Task.Target.ID, ShouldEqual, testEditionID)
						So(len(mockJobService.UpdateTaskStateCalls()), ShouldEqual, 1)
						So(mockJobService.UpdateTaskStateCalls()[0].TaskID, ShouldEqual, testEditionTask.ID)
						So(mockJobService.UpdateTaskStateCalls()[0].NewState, ShouldEqual, domain.StateInReview)
					})
				})
			})
		})
	})

	Convey("Given a dataset edition task executor with a zebedee client that errors", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
					return zebedee.Dataset{}, errors.New("unknown error")
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetEditionTaskExecutor(mockJobService, mockClientList, "")

		Convey("When migrate is called for a task", func() {
			task := testEditionTask

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And no version tasks are created", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a dataset edition task executor with a zebedee client that returns a non dataset page", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			UpdateTaskFunc: func(ctx context.Context, task *domain.Task) error { return nil },
		}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: "not_a_dataset_page",
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetEditionTaskExecutor(mockJobService, mockClientList, "")

		Convey("When migrate is called for a task", func() {
			task := testEditionTask

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And no version tasks are created", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a dataset edition task executor and a jobService that fails to update a task", t, func() {
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
				CreateDatasetFunc: func(ctx context.Context, headers sdk.Headers, dataset models.Dataset) (models.DatasetUpdate, error) {
					return models.DatasetUpdate{}, nil
				},
			},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: zebedee.PageTypeDataset,
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetEditionTaskExecutor(mockJobService, mockClientList, "")

		Convey("When migrate is called for a task", func() {
			task := testEditionTask

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a dataset edition task executor and a jobService that fails to create a version task", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return nil, errors.New("failed to create task")
			},
			UpdateTaskFunc:      func(ctx context.Context, task *domain.Task) error { return nil },
			UpdateTaskStateFunc: func(ctx context.Context, taskID string, state domain.State) error { return nil },
		}
		mockClientList := &clients.ClientList{
			DatasetAPI: &datasetSDKMock.ClienterMock{
				CreateDatasetFunc: func(ctx context.Context, headers sdk.Headers, dataset models.Dataset) (models.DatasetUpdate, error) {
					return models.DatasetUpdate{}, nil
				},
			},
			Zebedee: &clientMocks.ZebedeeClientMock{
				GetDatasetFunc: func(ctx context.Context, collectionID, edition, lang, datasetID string) (zebedee.Dataset, error) {
					return zebedee.Dataset{
						Type: zebedee.PageTypeDataset,
						Versions: []zebedee.Version{
							{
								URI: "/previous/v1",
							},
						},
					}, nil
				},
			},
		}

		ctx := context.Background()

		executor := NewDatasetEditionTaskExecutor(mockJobService, mockClientList, "")

		Convey("When migrate is called for a task", func() {
			task := testEditionTask

			err := executor.Migrate(ctx, task)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)

				Convey("And no further version tasks are created for the dataset", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 1)
				})
			})
		})
	})
}
