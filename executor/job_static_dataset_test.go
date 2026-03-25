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
	testJobNumber    = 1
	testCollectionID = "migration-collection-1"
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
}
