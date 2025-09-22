package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	migratorMock "github.com/ONSdigital/dis-migration-service/migrator/mock"
	storeMock "github.com/ONSdigital/dis-migration-service/store/mock"

	"github.com/ONSdigital/dis-migration-service/store"

	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	mockDatastore := store.Datastore{Backend: &storeMock.MongoDBMock{}}
	mockMigrator := migratorMock.MigratorMock{}

	Convey("Given an API instance", t, func() {
		r := mux.NewRouter()
		ctx := context.Background()
		api := Setup(ctx, r, &mockDatastore, &mockMigrator)

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
