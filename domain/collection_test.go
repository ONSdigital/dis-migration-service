package domain

import (
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewMigrationCollection(t *testing.T) {
	Convey("Given a job number", t, func() {
		jobNumber := 123

		Convey("When NewMigrationCollection is called", func() {
			collection := NewMigrationCollection(jobNumber)

			Convey("Then it returns a collection with the expected name and type", func() {
				So(collection.Name, ShouldEqual, CollectionNamePrefix+" 123")
				So(collection.Type, ShouldEqual, zebedee.CollectionTypeAutomated)
			})
		})
	})
}
