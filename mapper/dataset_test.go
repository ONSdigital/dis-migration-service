package mapper

import (
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMapDatasetLandingPageToDatasetAPI(t *testing.T) {
	Convey("Given a Zebedee dataset landing page", t, func() {
		pageData := getTestDatasetLandingPage()

		Convey("When it is mapped to a Dataset API dataset", func() {
			dataset, err := MapDatasetLandingPageToDatasetAPI("test-dataset-id", pageData)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the dataset fields are mapped correctly", func() {
				So(dataset.ID, ShouldEqual, "test-dataset-id")
				So(dataset.Title, ShouldEqual, "Test Dataset Title")
				So(dataset.Description, ShouldEqual, "This is a summary of the test dataset.")
				So(len(dataset.Contacts), ShouldEqual, 1)
				So(dataset.Contacts[0].Name, ShouldEqual, "John Doe")
				So(dataset.Contacts[0].Email, ShouldEqual, "john.doe@example.com")
				So(dataset.Contacts[0].Telephone, ShouldEqual, "123-456-7890")
				So(dataset.Keywords, ShouldResemble, []string{"test", "dataset", "sample"})
				So(dataset.NextRelease, ShouldEqual, "2024-12-31")
				So(dataset.QMI, ShouldNotBeNil)
				So(dataset.QMI.HRef, ShouldEqual, "/methodology/qmi/test-qmi")
				So(dataset.QMI.Title, ShouldEqual, "Test QMI Title")
				So(dataset.QMI.Description, ShouldEqual, "This is a summary of the test QMI.")
				So(dataset.License, ShouldEqual, "Open Government Licence v3.0")
			})
		})
	})

	Convey("Given a Zebedee page that is not a dataset landing page", t, func() {
		pageData := zebedee.DatasetLandingPage{
			Type: "unknown type",
		}

		Convey("When it is mapped to a Dataset API dataset", func() {
			dataset, err := MapDatasetLandingPageToDatasetAPI("test-dataset-id", pageData)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "invalid page type for dataset landing page")
			})

			Convey("Then the dataset is nil", func() {
				So(dataset, ShouldBeNil)
			})
		})
	})
}

func getTestDatasetLandingPage() zebedee.DatasetLandingPage {
	return zebedee.DatasetLandingPage{
		Type: zebedee.PageTypeDatasetLandingPage,
		Description: zebedee.Description{
			Title:   "Test Dataset Title",
			Summary: "This is a summary of the test dataset.",
			Contact: zebedee.Contact{
				Name:      "John Doe",
				Email:     "john.doe@example.com",
				Telephone: "123-456-7890",
			},
			Keywords:    []string{"test", "dataset", "sample"},
			NextRelease: "2024-12-31",
		},
		RelatedMethodology: []zebedee.Related{
			{
				URI:     "/methodology/qmi/test-qmi",
				Title:   "Test QMI Title",
				Summary: "This is a summary of the test QMI.",
			},
		},
	}
}
