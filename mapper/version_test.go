package mapper

import (
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	testEditionID = "test-edition-id"
)

func TestMapDatasetVersionToDatasetAPI(t *testing.T) {
	Convey("Given a Zebedee dataset version page with edition and series data", t, func() {
		pageData := getTestDatasetVersionPage()
		seriesData := getTestDatasetLandingPage()
		editionData := getTestDatasetEditionPage()

		Convey("When it is mapped to a Dataset API version", func() {
			version, err := MapDatasetVersionToDatasetAPI(testEditionID, pageData, seriesData, editionData)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And the dataset fields are mapped correctly", func() {
					So(version.Edition, ShouldEqual, testEditionID)
					So(version.EditionTitle, ShouldEqual, "Test Edition Title")
					So(version.Version, ShouldEqual, 1)
					So(version.ReleaseDate, ShouldEqual, "2024-01-01")
					So(version.UsageNotes, ShouldNotBeNil)
					So(len(*version.UsageNotes), ShouldEqual, 1)
					So((*version.UsageNotes)[0].Title, ShouldEqual, "Usage Notes")
					So((*version.UsageNotes)[0].Note, ShouldEqual, "These are the usage notes for the dataset.")
				})
			})
		})
	})

	Convey("Given a Zebedee dataset version page with a national statistics designation", t, func() {
		pageData := getTestDatasetVersionPage()
		seriesData := getTestDatasetLandingPage()
		editionData := getTestDatasetEditionPage()

		pageData.Description.NationalStatistic = true

		Convey("When it is mapped to a Dataset API version", func() {
			version, err := MapDatasetVersionToDatasetAPI(testEditionID, pageData, seriesData, editionData)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And the quality designation is applied correctly", func() {
					So(version.QualityDesignation, ShouldEqual, datasetModels.QualityDesignationAccreditedOfficial)
				})
			})
		})
	})

	Convey("Given a Zebedee dataset version page with multiple versions and a correction notice on the edition", t, func() {
		pageData := getTestDatasetVersionPage()
		seriesData := getTestDatasetLandingPage()
		editionData := getTestDatasetEditionPage()

		pageData.Versions = []zebedee.Version{
			{
				URI:    "/datasets/test-dataset/editions/test-edition-id/versions/1",
				Notice: "",
			},
		}

		Convey("When it is mapped to a Dataset API version", func() {
			version, err := MapDatasetVersionToDatasetAPI(testEditionID, pageData, seriesData, editionData)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)

				Convey("And the correction notice is mapped correctly", func() {
					So(version.Version, ShouldEqual, 2)
					So(version.Alerts, ShouldNotBeNil)
					So(len(*version.Alerts), ShouldEqual, 1)
					So((*version.Alerts)[0].Type, ShouldEqual, datasetModels.AlertTypeCorrection)
					So((*version.Alerts)[0].Description, ShouldEqual, "Correction notice for version 2.")
				})
			})
		})
	})

	Convey("Given a Zebedee page that is not a dataset", t, func() {
		pageData := zebedee.Dataset{
			Type: "unknown type",
		}

		seriesData := getTestDatasetLandingPage()
		editionData := getTestDatasetEditionPage()

		Convey("When it is mapped to a Dataset API dataset", func() {
			version, err := MapDatasetVersionToDatasetAPI(testEditionID, pageData, seriesData, editionData)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "invalid page type for dataset version page")
			})

			Convey("Then the dataset version is nil", func() {
				So(version, ShouldBeNil)
			})
		})
	})
}

func getTestDatasetVersionPage() zebedee.Dataset {
	return zebedee.Dataset{
		Type: zebedee.PageTypeDataset,
		Description: zebedee.Description{
			Title:   "Test Version Title",
			Edition: "Test Edition Title",
			Summary: "This is a summary of the test dataset.",
			Contact: zebedee.Contact{
				Name:      "John Doe",
				Email:     "john.doe@example.com",
				Telephone: "123-456-7890",
			},
			Keywords:    []string{"test", "dataset", "sample"},
			NextRelease: "2024-12-31",
			ReleaseDate: "2024-01-01",
		},
		Versions: []zebedee.Version{},
	}
}

func getTestDatasetEditionPage() zebedee.Dataset {
	return zebedee.Dataset{
		Type: zebedee.PageTypeDataset,
		Description: zebedee.Description{
			Title: "Test Dataset Title",
		},
		Versions: []zebedee.Version{
			{
				URI:    "/datasets/test-dataset/editions/2021/versions/1",
				Notice: "",
			},
			{
				URI:    "/datasets/test-dataset/editions/2022/versions/2",
				Notice: "Correction notice for version 2.",
			},
		},
	}
}
