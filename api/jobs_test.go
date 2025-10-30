package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	applicationMock "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	migratorMock "github.com/ONSdigital/dis-migration-service/migrator/mock"

	"github.com/google/uuid"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testSourceID = "/test-source-id"
	testTargetID = "test-target-id"
	testType     = domain.JobTypeStaticDataset
)

var (
	testID = uuid.New().String()
)

func TestGetJob(t *testing.T) {
	Convey("Given a test API instance and a mocked jobservice that returns a job", t, func() {
		mockService := applicationMock.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return &domain.Job{
					ID: jobID,
				}, nil
			},
		}
		mockMigrator := migratorMock.MigratorMock{}

		r := mux.NewRouter()
		ctx := context.Background()
		api := Setup(ctx, r, &mockService, &mockMigrator)

		Convey("When a valid request is made", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:30100/v1/migration-jobs/%s", testID), http.NoBody)
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then a job is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)
				So(resp.Body.String(), ShouldContainSubstring, testID)
			})
		})
	})
}

func TestCreateJob(t *testing.T) {
	Convey("Given an test API instance and a mocked jobservice that creates a job", t, func() {
		mockService := applicationMock.JobServiceMock{
			CreateJobFunc: func(ctx context.Context, jobConfig *domain.JobConfig) (*domain.Job, error) {
				return &domain.Job{
					Config:      jobConfig,
					ID:          testID,
					LastUpdated: time.Now().Format(time.RFC3339),
				}, nil
			},
		}

		mockMigrator := migratorMock.MigratorMock{
			MigrateFunc: func(ctx context.Context, job *domain.Job) {},
		}

		r := mux.NewRouter()
		ctx := context.Background()

		api := Setup(ctx, r, &mockService, &mockMigrator)

		Convey("When a valid request is made", func() {
			body := domain.JobConfig{
				SourceID: testSourceID,
				TargetID: testTargetID,
				Type:     testType,
			}

			bodyBytes, err := json.Marshal(body)
			So(err, ShouldBeNil)

			req := httptest.NewRequest(http.MethodPost, "http://localhost:30100/v1/migration-jobs", bytes.NewBuffer(bodyBytes))
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then a created job is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusAccepted)

				bodyString := resp.Body.Bytes()

				var job domain.Job

				err := json.Unmarshal(bodyString, &job)
				So(err, ShouldBeNil)

				So(job.Config.SourceID, ShouldEqual, testSourceID)
				So(job.Config.TargetID, ShouldEqual, testTargetID)
				So(job.Config.Type, ShouldEqual, testType)

				// Check LastUpdated is a valid RFC3339 timestamp
				parsedTime, err := time.Parse(time.RFC3339, job.LastUpdated)
				So(err, ShouldBeNil)

				// Check LastUpdated is recent (within the last 1 minute)
				now := time.Now()
				So(parsedTime.After(now.Add(-1*time.Minute)), ShouldBeTrue)
				So(parsedTime.Before(now.Add(1*time.Minute)), ShouldBeTrue)

				// Assert job ID is valid uuid
				err = uuid.Validate(job.ID)
				So(err, ShouldBeNil)

				Convey("And the Migrator is called to start", func() {
					So(len(mockMigrator.MigrateCalls()), ShouldEqual, 1)
				})
			})
		})

		Convey("When an invalid request is made", func() {
			bodyBytes := []byte("invalidJson")

			req := httptest.NewRequest(http.MethodPost, "http://localhost:30100/v1/migration-jobs", bytes.NewBuffer(bodyBytes))
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then a bad request is returned", func() {
				bodyString := resp.Body.String()

				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				So(bodyString, ShouldContainSubstring, appErrors.ErrUnableToParseBody.Error())

				Convey("And the Migrator is not called to start", func() {
					So(len(mockMigrator.MigrateCalls()), ShouldEqual, 0)
				})
			})
		})

		Convey("When a valid request is made with an invalid parameter", func() {
			body := domain.JobConfig{
				SourceID: testSourceID,
				TargetID: testTargetID,
			}

			bodyBytes, err := json.Marshal(body)
			So(err, ShouldBeNil)

			req := httptest.NewRequest(http.MethodPost, "http://localhost:30100/v1/migration-jobs", bytes.NewBuffer(bodyBytes))
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then a bad request is returned", func() {
				bodyString := resp.Body.String()

				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				So(bodyString, ShouldContainSubstring, appErrors.ErrJobTypeNotProvided.Error())

				Convey("And the Migrator is not called to start", func() {
					So(len(mockMigrator.MigrateCalls()), ShouldEqual, 0)
				})
			})
		})

		Convey("When a valid request is made with multiple invalid parameters", func() {
			body := domain.JobConfig{
				SourceID: testSourceID,
			}

			bodyBytes, err := json.Marshal(body)
			So(err, ShouldBeNil)

			req := httptest.NewRequest(http.MethodPost, "http://localhost:30100/v1/migration-jobs", bytes.NewBuffer(bodyBytes))
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then a bad request is returned with multiple errors", func() {
				bodyString := resp.Body.String()

				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				So(bodyString, ShouldContainSubstring, appErrors.ErrJobTypeNotProvided.Error())
				So(bodyString, ShouldContainSubstring, appErrors.ErrTargetIDNotProvided.Error())

				Convey("And the Migrator is not called to start", func() {
					So(len(mockMigrator.MigrateCalls()), ShouldEqual, 0)
				})
			})
		})
	})
}
