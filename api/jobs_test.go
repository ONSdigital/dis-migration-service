package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	applicationMock "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/config"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	migratorMock "github.com/ONSdigital/dis-migration-service/migrator/mock"
	authorisationMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"

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

		mockAuthMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		r := mux.NewRouter()
		ctx := context.Background()
		cfg := &config.Config{}
		api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

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

	Convey("Given a test API instance and a mocked jobservice that returns not found", t, func() {
		mockService := applicationMock.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return nil, appErrors.ErrJobNotFound
			},
		}
		mockMigrator := migratorMock.MigratorMock{}

		mockAuthMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		r := mux.NewRouter()
		ctx := context.Background()
		cfg := &config.Config{}
		api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

		Convey("When a request is made for a missing job", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:30100/v1/migration-jobs/%s", testID), http.NoBody)
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then a 404 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusNotFound)
				So(resp.Body.String(), ShouldContainSubstring, appErrors.ErrJobNotFound.Error())
			})
		})
	})

	Convey("Given a test API instance and a mocked jobservice that returns an internal error", t, func() {
		mockService := applicationMock.JobServiceMock{
			GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
				return nil, errors.New("database failure")
			},
		}
		mockMigrator := migratorMock.MigratorMock{}

		mockAuthMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		r := mux.NewRouter()
		ctx := context.Background()
		cfg := &config.Config{}
		api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

		Convey("When a request is made and the service errors", func() {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:30100/v1/migration-jobs/%s", testID), http.NoBody)
			resp := httptest.NewRecorder()

			api.Router.ServeHTTP(resp, req)

			Convey("Then a 500 is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusInternalServerError)
				// you can assert on the body if your handleError writes a specific message
			})
		})
	})
}

func TestGetJobs(t *testing.T) {
	Convey("Given a test API instance and a mocked jobservice that returns multiple jobs", t, func() {
		mockService := applicationMock.JobServiceMock{
			GetJobsFunc: func(ctx context.Context, limit, offset int) ([]*domain.Job, int, error) {
				jobs := []*domain.Job{
					{ID: "job1"},
					{ID: "job2"},
				}
				return jobs, len(jobs), nil
			},
		}
		mockMigrator := migratorMock.MigratorMock{}

		mockAuthMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		r := mux.NewRouter()
		ctx := context.Background()
		cfg := &config.Config{}
		api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

		Convey("When a valid request is made", func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs", http.NoBody)
			resp := httptest.NewRecorder()
			api.Router.ServeHTTP(resp, req)

			Convey("Then multiple jobs are returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				bodyBytes, err := io.ReadAll(resp.Body)
				So(err, ShouldBeNil)

				var paginatedResponse struct {
					Items []domain.Job `json:"items"`
				}

				err = json.Unmarshal(bodyBytes, &paginatedResponse)
				So(err, ShouldBeNil)
				So(len(paginatedResponse.Items), ShouldEqual, 2)
			})
		})

		Convey("When an invalid request is made with a non-integer limit", func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs?limit=invalid", http.NoBody)
			resp := httptest.NewRecorder()
			api.Router.ServeHTTP(resp, req)

			Convey("Then a bad request is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				So(resp.Body.String(), ShouldContainSubstring, appErrors.ErrLimitInvalid.Error())
			})
		})
		Convey("When an invalid request is made with a non-integer offset", func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs?offset=invalid", http.NoBody)
			resp := httptest.NewRecorder()
			api.Router.ServeHTTP(resp, req)

			Convey("Then a bad request is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				So(resp.Body.String(), ShouldContainSubstring, appErrors.ErrOffsetInvalid.Error())
			})
		})

		Convey("When an invalid request is made with a limit exceeding max", func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs?limit=1000", http.NoBody)
			resp := httptest.NewRecorder()
			api.Router.ServeHTTP(resp, req)

			Convey("Then a bad request is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				So(resp.Body.String(), ShouldContainSubstring, appErrors.ErrLimitExceeded.Error())
			})
		})

		Convey("When an invalid request is made with invalid limit and offset", func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs?limit=invalid&offset=invalid", http.NoBody)
			resp := httptest.NewRecorder()
			api.Router.ServeHTTP(resp, req)

			Convey("Then a bad request is returned with multiple errors", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				So(resp.Body.String(), ShouldContainSubstring, appErrors.ErrLimitInvalid.Error())
				So(resp.Body.String(), ShouldContainSubstring, appErrors.ErrOffsetInvalid.Error())
			})
		})
	})

	Convey("Given a test API instance and a mocked jobservice that returns no jobs", t, func() {
		mockService := applicationMock.JobServiceMock{
			GetJobsFunc: func(ctx context.Context, limit, offset int) ([]*domain.Job, int, error) {
				return []*domain.Job{}, 0, nil
			},
		}
		mockMigrator := migratorMock.MigratorMock{}

		mockAuthMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		r := mux.NewRouter()
		ctx := context.Background()
		cfg := &config.Config{}
		api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

		Convey("When a valid request is made", func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs", http.NoBody)
			resp := httptest.NewRecorder()
			api.Router.ServeHTTP(resp, req)

			Convey("Then no jobs are returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				bodyBytes, err := io.ReadAll(resp.Body)
				So(err, ShouldBeNil)

				var paginatedResponse struct {
					Items []domain.Job `json:"items"`
				}

				err = json.Unmarshal(bodyBytes, &paginatedResponse)
				So(err, ShouldBeNil)
				So(len(paginatedResponse.Items), ShouldEqual, 0)
			})
		})
	})
}

func TestCreateJob(t *testing.T) {
	Convey("Given an test API instance and a mocked jobservice that creates a job", t, func() {
		testConfig := domain.JobConfig{
			SourceID: testSourceID,
			TargetID: testTargetID,
			Type:     testType,
		}

		createdJob := &domain.Job{
			Config:      &testConfig,
			ID:          testID,
			LastUpdated: time.Now().UTC(),
		}

		mockService := applicationMock.JobServiceMock{
			CreateJobFunc: func(ctx context.Context, jobConfig *domain.JobConfig) (*domain.Job, error) {
				return createdJob, nil
			},
		}

		mockMigrator := migratorMock.MigratorMock{
			MigrateFunc: func(ctx context.Context, job *domain.Job) {},
		}

		mockAuthMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
				return handlerFunc
			},
			CloseFunc: func(ctx context.Context) error {
				return nil
			},
		}

		r := mux.NewRouter()
		ctx := context.Background()
		cfg := &config.Config{}
		api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

		Convey("When a valid request is made", func() {
			bodyBytes, err := json.Marshal(testConfig)
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

				So(&job, ShouldEqual, createdJob)

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
