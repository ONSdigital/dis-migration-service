package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	applicationMock "github.com/ONSdigital/dis-migration-service/application/mock"
	"github.com/ONSdigital/dis-migration-service/config"
	migratorMock "github.com/ONSdigital/dis-migration-service/migrator/mock"
	authorisationMock "github.com/ONSdigital/dp-authorisation/v2/authorisation/mock"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	mockService := applicationMock.JobServiceMock{}
	mockMigrator := migratorMock.MigratorMock{}

	mockAuthMiddleware := &authorisationMock.MiddlewareMock{
		RequireFunc: func(permission string, handlerFunc http.HandlerFunc) http.HandlerFunc {
			return handlerFunc
		},
		CloseFunc: func(ctx context.Context) error {
			return nil
		},
	}

	Convey("Given an API instance", t, func() {
		r := mux.NewRouter()
		ctx := context.Background()
		cfg := &config.Config{}

		api := Setup(ctx, cfg, r, &mockService, &mockMigrator, mockAuthMiddleware)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(api.Router, "/v1/migration-jobs", "POST"), ShouldBeTrue)
			So(hasRoute(api.Router, "/v1/migration-jobs/myJob", "GET"), ShouldBeTrue)
			So(hasRoute(api.Router, "/v1/migration-jobs/myJob/tasks", "GET"), ShouldBeTrue)
			So(hasRoute(api.Router, "/v1/migration-jobs/myJob/events", "GET"), ShouldBeTrue)
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, http.NoBody)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
