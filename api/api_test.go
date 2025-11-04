package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	applicationMock "github.com/ONSdigital/dis-migration-service/application/mock"
	migratorMock "github.com/ONSdigital/dis-migration-service/migrator/mock"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	mockService := applicationMock.JobServiceMock{}
	mockMigrator := migratorMock.MigratorMock{}

	Convey("Given an API instance", t, func() {
		r := mux.NewRouter()
		ctx := context.Background()

		api := Setup(ctx, r, &mockService, &mockMigrator)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(api.Router, "/v1/migration-jobs", "POST"), ShouldBeTrue)
			So(hasRoute(api.Router, "/v1/migration-jobs/myJob", "GET"), ShouldBeTrue)
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, http.NoBody)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}
