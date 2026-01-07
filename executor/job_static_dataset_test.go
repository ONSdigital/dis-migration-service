package executor

import (
	"context"
	"errors"
	"testing"

	applicationMocks "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testJobID = "job-1"
)

func TestJobStaticDataset(t *testing.T) {
	Convey("Given a static dataset job executor and a job service that does not error", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return &domain.Task{}, nil
			},
		}
		mockClientList := &clients.ClientList{}

		ctx := context.Background()

		executor := NewStaticDatasetJobExecutor(mockJobService, mockClientList)

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

				Convey("And a dataset series migration task is created for the dataset", func() {
					So(len(mockJobService.CreateTaskCalls()), ShouldEqual, 1)
					So(mockJobService.CreateTaskCalls()[0].JobNumber, ShouldEqual, 1)
					So(mockJobService.CreateTaskCalls()[0].Task.Type, ShouldEqual, domain.TaskTypeDatasetSeries)
					So(mockJobService.CreateTaskCalls()[0].Task.Source.ID, ShouldEqual, "source-dataset-id")
					So(mockJobService.CreateTaskCalls()[0].Task.Target.ID, ShouldEqual, "target-dataset-id")
				})
			})
		})
	})
	Convey("Given a static dataset job executor and a job service that errors when creating a task", t, func() {
		mockJobService := &applicationMocks.JobServiceMock{
			CreateTaskFunc: func(ctx context.Context, jobNumber int, task *domain.Task) (*domain.Task, error) {
				return nil, errors.New("create task error")
			},
			UpdateJobStateFunc: func(ctx context.Context, jobNumber int, state domain.State) error {
				return nil
			},
		}
		mockClientList := &clients.ClientList{}

		ctx := context.Background()

		executor := NewStaticDatasetJobExecutor(mockJobService, mockClientList)

		Convey("When migrate is called for a job", func() {
			job := &domain.Job{
				ID: "job-1",
				Config: &domain.JobConfig{
					SourceID: "source-dataset-id",
					TargetID: "target-dataset-id",
				},
				State: domain.StateMigrating,
			}

			err := executor.Migrate(ctx, job)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
