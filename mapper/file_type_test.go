package mapper

import (
	"fmt"
	"testing"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDeriveMimeType(t *testing.T) {
	Convey("Given a file with a valid name and mime type", t, func() {
		cases := []struct {
			input    string
			expected string
		}{
			{"file.csv", MimeTypeCSV},
			{"file.xlsx", MimeTypeXLSX},
			{"file.xls", MimeTypeXLS},
			{"file.sdmx", MimeTypeSDMX},
			{"file.csdb", MimeTypeCSDB},
			{"file.csvw", MimeTypeCSVW},
		}
		for _, c := range cases {
			Convey(fmt.Sprintf("When the file name is %s", c.input), func() {
				mimeType := DeriveMimeTypeFromFilename(c.input)

				Convey("Then the correct mime type is returned", func() {
					So(mimeType, ShouldEqual, c.expected)
				})
			})
		}
	})

	Convey("Given a file with an unknown extension", t, func() {
		Convey("When the file name is file.unknown", func() {
			mimeType := DeriveMimeTypeFromFilename("file.unknown")

			Convey("Then the default mime type is returned", func() {
				So(mimeType, ShouldEqual, MimeTypeOctetStream)
			})
		})
	})
}

func TestMapMimeTypeToDistributionFormat(t *testing.T) {
	Convey("Given a valid mime type", t, func() {
		cases := []struct {
			input    string
			expected datasetModels.DistributionFormat
		}{
			{MimeTypeCSV, datasetModels.DistributionFormatCSV},
			{MimeTypeXLSX, datasetModels.DistributionFormatXLSX},
			{MimeTypeXLS, datasetModels.DistributionFormatXLS},
			{MimeTypeSDMX, datasetModels.DistributionFormatSDMX},
			{MimeTypeCSDB, datasetModels.DistributionFormatCSDB},
			{MimeTypeCSVW, datasetModels.DistributionFormatCSVWMeta},
		}
		for _, c := range cases {
			Convey(fmt.Sprintf("When the mime type is %s", c.input), func() {
				format, err := MapMimeTypeToDistributionFormat(c.input)

				Convey("Then the correct distribution format is returned", func() {
					So(err, ShouldBeNil)
					So(format, ShouldEqual, c.expected)
				})
			})
		}
	})

	Convey("Given an invalid mime type", t, func() {
		Convey("When the mime type is application/unknown", func() {
			_, err := MapMimeTypeToDistributionFormat("application/unknown")

			Convey("Then an unsupported distribution format error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, appErrors.ErrUnsupportedDistributionFormat)
			})
		})
	})
}
