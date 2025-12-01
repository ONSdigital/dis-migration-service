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
	testJobID    = "job-123"
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
			GetJobsFunc: func(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error) {
				jobs := []*domain.Job{
					{ID: "job1", State: domain.JobStateSubmitted},
					{ID: "job2", State: domain.JobStateApproved},
					{ID: "job3", State: domain.JobStateCompleted},
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

		Convey("When a valid request is made with no state filter", func() {
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
				So(len(paginatedResponse.Items), ShouldEqual, 3)

				Convey("And the service is called with all states", func() {
					So(len(mockService.GetJobsCalls()), ShouldEqual, 1)
					So(mockService.GetJobsCalls()[0].States, ShouldResemble, []domain.JobState{})
				})
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

		Convey("When a request is made with a single valid state", func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs?state=submitted", http.NoBody)
			resp := httptest.NewRecorder()
			api.Router.ServeHTTP(resp, req)

			Convey("Then only jobs matching that state are returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				Convey("And the service is called with the submitted state", func() {
					So(len(mockService.GetJobsCalls()), ShouldEqual, 1)
					So(mockService.GetJobsCalls()[0].States, ShouldResemble, []domain.JobState{domain.JobStateSubmitted})
				})
			})
		})

		Convey("When a request is made with multiple valid states", func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs?state=submitted&state=approved", http.NoBody)
			resp := httptest.NewRecorder()
			api.Router.ServeHTTP(resp, req)

			Convey("Then jobs matching any of the states are returned", func() {
				So(resp.Code, ShouldEqual, http.StatusOK)

				Convey("And the service is called with both states", func() {
					So(len(mockService.GetJobsCalls()), ShouldEqual, 1)
					So(mockService.GetJobsCalls()[0].States, ShouldResemble, []domain.JobState{domain.JobStateSubmitted, domain.JobStateApproved})
				})
			})
		})

		Convey("When a request is made with an invalid state", func() {
			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs?state=unknown", http.NoBody)
			resp := httptest.NewRecorder()
			api.Router.ServeHTTP(resp, req)

			Convey("Then a 400 Bad Request is returned", func() {
				So(resp.Code, ShouldEqual, http.StatusBadRequest)
				So(resp.Body.String(), ShouldContainSubstring, "job state parameter is invalid")

				Convey("And the service is not called", func() {
					So(len(mockService.GetJobsCalls()), ShouldEqual, 0)
				})
			})
		})
	})

	Convey("Given a test API instance and a mocked jobservice that returns no jobs", t, func() {
		mockService := applicationMock.JobServiceMock{
			GetJobsFunc: func(ctx context.Context, states []domain.JobState, limit, offset int) ([]*domain.Job, int, error) {
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
	Convey("Given a test API instance and a mocked jobservice that creates a job", t, func() {
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

func TestGetJobTasks(t *testing.T) {
	Convey("Given a test API instance and a mocked jobservice", t, func() {
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

		Convey("missing job id should return ErrJobIDNotProvided", func() {
			mockService := applicationMock.JobServiceMock{}
			api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

			// Build request and set empty mux var to simulate missing job id
			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs//tasks", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: ""})
			rr := httptest.NewRecorder()

			items, total, err := api.getJobTasks(rr, req, 10, 0)
			So(items, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, appErrors.ErrJobIDNotProvided)
		})

		Convey("job not found should return ErrJobNotFound", func() {
			mockService := applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return nil, appErrors.ErrJobNotFound
				},
			}
			api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs/job-123/tasks", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: "job-123"})
			rr := httptest.NewRecorder()

			items, total, err := api.getJobTasks(rr, req, 10, 0)
			So(items, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, appErrors.ErrJobNotFound)
		})

		Convey("GetJob returns unexpected error is propagated", func() {
			testErr := fmt.Errorf("db failure")
			mockService := applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return nil, testErr
				},
			}
			api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs/job-123/tasks", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: "job-123"})
			rr := httptest.NewRecorder()

			items, total, err := api.getJobTasks(rr, req, 10, 0)
			So(items, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(err, ShouldBeError, testErr)
		})

		Convey("success returns tasks and total", func() {
			testJobID := testJobID
			mockTasks := []*domain.Task{
				{ID: "t1", JobID: testJobID},
				{ID: "t2", JobID: testJobID},
			}

			mockService := applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return &domain.Job{ID: jobID}, nil
				},
				GetJobTasksFunc: func(ctx context.Context, jobID string, limit int, offset int) ([]*domain.Task, int, error) {
					// return tasks as interface{} to match service signature used in handler
					return mockTasks, len(mockTasks), nil
				},
			}
			api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs/job-123/tasks", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: testJobID})
			rr := httptest.NewRecorder()

			items, total, err := api.getJobTasks(rr, req, 10, 0)
			So(err, ShouldBeNil)
			So(total, ShouldEqual, 2)

			// items is returned as interface{}, assert underlying value
			gotTasks, ok := items.([]*domain.Task)
			So(ok, ShouldBeTrue)
			So(len(gotTasks), ShouldEqual, 2)
			So(gotTasks[0].ID, ShouldEqual, "t1")
			So(gotTasks[1].ID, ShouldEqual, "t2")
		})

		Convey("GetJobTasks returns an error is propagated", func() {
			testErr := fmt.Errorf("find failure")
			mockService := applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return &domain.Job{ID: jobID}, nil
				},
				GetJobTasksFunc: func(ctx context.Context, jobID string, limit int, offset int) ([]*domain.Task, int, error) {
					return nil, 0, testErr
				},
			}
			api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet, "http://localhost:30100/v1/migration-jobs/job-123/tasks", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: "job-123"})
			rr := httptest.NewRecorder()

			items, total, err := api.getJobTasks(rr, req, 10, 0)
			So(items, ShouldBeNil)
			So(total, ShouldEqual, 0)
			So(err, ShouldBeError, testErr)
		})
	})
}

func TestGetJobEvents(t *testing.T) {
	Convey("Given the getJobEvents endpoint", t, func() {
		mockMigrator := migratorMock.MigratorMock{}
		mockAuthMiddleware := &authorisationMock.MiddlewareMock{
			RequireFunc: func(permission string, handler http.HandlerFunc) http.HandlerFunc {
				return handler
			},
			CloseFunc: func(ctx context.Context) error { return nil },
		}

		r := mux.NewRouter()
		ctx := context.Background()
		cfg := &config.Config{}

		Convey("And jobID is missing from the request", func() {
			mockService := &applicationMock.JobServiceMock{}
			api := Setup(ctx, cfg, r, mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet,
				"http://localhost:30100/v1/migration-jobs//events", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: ""})
			rr := httptest.NewRecorder()

			Convey("When getJobEvents is called", func() {
				items, total, err := api.getJobEvents(rr, req, 10, 0)

				Convey("Then it returns ErrJobIDNotProvided", func() {
					So(items, ShouldBeNil)
					So(total, ShouldEqual, 0)
					So(err, ShouldEqual, appErrors.ErrJobIDNotProvided)
				})
			})
		})

		Convey("And GetJob returns ErrJobNotFound", func() {
			mockService := &applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return nil, appErrors.ErrJobNotFound
				},
			}
			api := Setup(ctx, cfg, r, mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet,
				"http://localhost:30100/v1/migration-jobs/job-123/events", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: testJobID})
			rr := httptest.NewRecorder()

			Convey("When getJobEvents is called", func() {
				items, total, err := api.getJobEvents(rr, req, 10, 0)

				Convey("Then it returns ErrJobNotFound", func() {
					So(items, ShouldBeNil)
					So(total, ShouldEqual, 0)
					So(err, ShouldEqual, appErrors.ErrJobNotFound)
				})
			})
		})

		Convey("And GetJob returns an unexpected error", func() {
			testErr := fmt.Errorf("db failure")
			mockService := &applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return nil, testErr
				},
			}
			api := Setup(ctx, cfg, r, mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet,
				"http://localhost:30100/v1/migration-jobs/job-123/events", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: testJobID})
			rr := httptest.NewRecorder()

			Convey("When getJobEvents is called", func() {
				items, total, err := api.getJobEvents(rr, req, 10, 0)

				Convey("Then it returns the error", func() {
					So(items, ShouldBeNil)
					So(total, ShouldEqual, 0)
					So(err, ShouldEqual, testErr)
				})
			})
		})

		Convey("And GetJob and GetJobEvents return successfully", func() {
			mockEvents := []*domain.Event{
				{
					ID:        "event-1",
					JobID:     testJobID,
					CreatedAt: "2025-11-19T13:30:00Z",
					Action:    "submitted",
					RequestedBy: &domain.User{
						ID:    "user-1",
						Email: "user1@ons.gov.uk",
					},
				},
				{
					ID:        "event-2",
					JobID:     testJobID,
					CreatedAt: "2025-11-19T13:35:00Z",
					Action:    "approved",
					RequestedBy: &domain.User{
						ID:    "user-2",
						Email: "user2@ons.gov.uk",
					},
				},
			}

			mockService := &applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return &domain.Job{ID: jobID}, nil
				},
				GetJobEventsFunc: func(ctx context.Context, jobID string, limit int, offset int) ([]*domain.Event, int, error) {
					return mockEvents, len(mockEvents), nil
				},
			}
			api := Setup(ctx, cfg, r, mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet,
				"http://localhost:30100/v1/migration-jobs/job-123/events", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: testJobID})
			rr := httptest.NewRecorder()

			Convey("When getJobEvents is called", func() {
				items, total, err := api.getJobEvents(rr, req, 10, 0)

				Convey("Then it returns the events with correct details", func() {
					So(err, ShouldBeNil)
					So(total, ShouldEqual, 2)

					gotEvents, ok := items.([]*domain.Event)
					So(ok, ShouldBeTrue)
					So(len(gotEvents), ShouldEqual, 2)
					So(gotEvents[0].ID, ShouldEqual, "event-1")
					So(gotEvents[0].Action, ShouldEqual, "submitted")
					So(gotEvents[1].ID, ShouldEqual, "event-2")
					So(gotEvents[1].Action, ShouldEqual, "approved")
				})
			})
		})

		Convey("And GetJob succeeds but GetJobEvents returns an error", func() {
			testErr := fmt.Errorf("find failure")
			mockService := &applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return &domain.Job{ID: jobID}, nil
				},
				GetJobEventsFunc: func(ctx context.Context, jobID string, limit int, offset int) ([]*domain.Event, int, error) {
					return nil, 0, testErr
				},
			}
			api := Setup(ctx, cfg, r, mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet,
				"http://localhost:30100/v1/migration-jobs/job-123/events", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: testJobID})
			rr := httptest.NewRecorder()

			Convey("When getJobEvents is called", func() {
				items, total, err := api.getJobEvents(rr, req, 10, 0)

				Convey("Then it returns the error", func() {
					So(items, ShouldBeNil)
					So(total, ShouldEqual, 0)
					So(err, ShouldEqual, testErr)
				})
			})
		})

		Convey("And custom limit and offset are provided", func() {
			mockEvents := []*domain.Event{
				{
					ID:          "event-2",
					JobID:       testJobID,
					Action:      "approved",
					RequestedBy: &domain.User{ID: "user-2"},
				},
			}

			mockService := &applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return &domain.Job{ID: jobID}, nil
				},
				GetJobEventsFunc: func(ctx context.Context, jobID string, limit int, offset int) ([]*domain.Event, int, error) {
					So(limit, ShouldEqual, 1)
					So(offset, ShouldEqual, 1)
					return mockEvents, 5, nil
				},
			}
			api := Setup(ctx, cfg, r, mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet,
				"http://localhost:30100/v1/migration-jobs/job-123/events?limit=1&offset=1",
				http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: testJobID})
			rr := httptest.NewRecorder()

			Convey("When getJobEvents is called with limit 1 and offset 1", func() {
				items, total, err := api.getJobEvents(rr, req, 1, 1)

				Convey("Then it returns the paginated results", func() {
					So(err, ShouldBeNil)
					So(total, ShouldEqual, 5)

					gotEvents, ok := items.([]*domain.Event)
					So(ok, ShouldBeTrue)
					So(len(gotEvents), ShouldEqual, 1)
					So(gotEvents[0].ID, ShouldEqual, "event-2")
				})
			})
		})

		Convey("And GetJobEvents returns an empty list", func() {
			mockService := &applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return &domain.Job{ID: jobID}, nil
				},
				GetJobEventsFunc: func(ctx context.Context, jobID string, limit int, offset int) ([]*domain.Event, int, error) {
					return []*domain.Event{}, 0, nil
				},
			}
			api := Setup(ctx, cfg, r, mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet,
				"http://localhost:30100/v1/migration-jobs/job-123/events", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: testJobID})
			rr := httptest.NewRecorder()

			Convey("When getJobEvents is called", func() {
				items, total, err := api.getJobEvents(rr, req, 10, 0)

				Convey("Then it returns an empty list with zero total", func() {
					So(err, ShouldBeNil)
					So(total, ShouldEqual, 0)

					gotEvents, ok := items.([]*domain.Event)
					So(ok, ShouldBeTrue)
					So(len(gotEvents), ShouldEqual, 0)
				})
			})
		})

		Convey("And GetJobEvents returns events with different user types", func() {
			mockEvents := []*domain.Event{
				{
					ID:     "event-human",
					JobID:  testJobID,
					Action: "approved",
					RequestedBy: &domain.User{
						ID:    "human-user",
						Email: "human@ons.gov.uk",
					},
				},
				{
					ID:     "event-service",
					JobID:  testJobID,
					Action: "migrating",
					RequestedBy: &domain.User{
						ID: "service-account",
					},
				},
			}

			mockService := &applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return &domain.Job{ID: jobID}, nil
				},
				GetJobEventsFunc: func(ctx context.Context, jobID string, limit int, offset int) ([]*domain.Event, int, error) {
					return mockEvents, len(mockEvents), nil
				},
			}
			api := Setup(ctx, cfg, r, mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet,
				"http://localhost:30100/v1/migration-jobs/job-123/events", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: testJobID})
			rr := httptest.NewRecorder()

			Convey("When getJobEvents is called", func() {
				items, total, err := api.getJobEvents(rr, req, 10, 0)

				Convey("Then it preserves user details correctly", func() {
					So(err, ShouldBeNil)
					So(total, ShouldEqual, 2)

					gotEvents, ok := items.([]*domain.Event)
					So(ok, ShouldBeTrue)
					So(len(gotEvents), ShouldEqual, 2)
					So(gotEvents[0].RequestedBy.Email, ShouldEqual, "human@ons.gov.uk")
					So(gotEvents[1].RequestedBy.Email, ShouldEqual, "")
				})
			})
		})

		Convey("And GetJobEvents returns events with various action types", func() {
			mockEvents := []*domain.Event{
				{
					ID:          "event-1",
					JobID:       testJobID,
					Action:      "submitted",
					RequestedBy: &domain.User{ID: "user-1"},
				},
				{
					ID:          "event-2",
					JobID:       testJobID,
					Action:      "migrating",
					RequestedBy: &domain.User{ID: "user-2"},
				},
				{
					ID:          "event-3",
					JobID:       testJobID,
					Action:      "completed",
					RequestedBy: &domain.User{ID: "user-3"},
				},
			}

			mockService := &applicationMock.JobServiceMock{
				GetJobFunc: func(ctx context.Context, jobID string) (*domain.Job, error) {
					return &domain.Job{ID: jobID}, nil
				},
				GetJobEventsFunc: func(ctx context.Context, jobID string, limit int, offset int) ([]*domain.Event, int, error) {
					return mockEvents, len(mockEvents), nil
				},
			}
			api := Setup(ctx, cfg, r, mockService, &mockMigrator, mockAuthMiddleware)

			req := httptest.NewRequest(http.MethodGet,
				"http://localhost:30100/v1/migration-jobs/job-123/events", http.NoBody)
			req = mux.SetURLVars(req, map[string]string{PathParameterJobID: testJobID})
			rr := httptest.NewRecorder()

			Convey("When getJobEvents is called", func() {
				items, total, err := api.getJobEvents(rr, req, 10, 0)

				Convey("Then it returns all events with their action types", func() {
					So(err, ShouldBeNil)
					So(total, ShouldEqual, 3)

					gotEvents, ok := items.([]*domain.Event)
					So(ok, ShouldBeTrue)
					So(len(gotEvents), ShouldEqual, 3)
					So(gotEvents[0].Action, ShouldEqual, "submitted")
					So(gotEvents[1].Action, ShouldEqual, "migrating")
					So(gotEvents[2].Action, ShouldEqual, "completed")
				})
			})
		})
	})
}
