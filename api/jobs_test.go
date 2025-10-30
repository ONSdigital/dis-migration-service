package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	migratorMock "github.com/ONSdigital/dis-migration-service/migrator/mock"

	"github.com/ONSdigital/dis-migration-service/store/mock"

	"github.com/ONSdigital/dis-migration-service/store"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testID       = "testID"
	testSourceID = "test-source-id"
	testTargetID = "test-target-id"
	testType     = "test-type"
)

func TestGetJob(t *testing.T) {
	Convey("Given a test API instance and a mocked datastore that returns a job", t, func() {
		mockDatastore := store.Datastore{
			Backend: &mock.MongoDBMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return &domain.Job{
						ID: jobID,
					}, nil
				},
				CloseFunc: func(ctx context.Context) error { return nil },
			},
		}
		mockMigrator := migratorMock.MigratorMock{}

		r := mux.NewRouter()
		ctx := context.Background()
		api := Setup(ctx, r, &mockDatastore, &mockMigrator)

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
	Convey("Given an test API instance and a mocked datastore that creates a job", t, func() {
		mockDatastore := store.Datastore{
			Backend: &mock.MongoDBMock{
				CreateJobFunc: func(ctx context.Context, job *domain.Job) (*domain.Job, error) {
					return job, nil
				},
				CloseFunc: func(ctx context.Context) error { return nil },
			},
		}

		mockMigrator := migratorMock.MigratorMock{
			MigrateFunc: func(ctx context.Context, job *domain.Job) {},
		}

		r := mux.NewRouter()
		ctx := context.Background()
		api := Setup(ctx, r, &mockDatastore, &mockMigrator)

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

				bodyString := resp.Body.String()

				So(bodyString, ShouldContainSubstring, testSourceID)
				So(bodyString, ShouldContainSubstring, testTargetID)
				So(bodyString, ShouldContainSubstring, testType)

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
