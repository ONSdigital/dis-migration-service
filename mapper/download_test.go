package mapper

import (
	"testing"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	uploadAPI "github.com/ONSdigital/dp-upload-service/api"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMapResourceToUploadServiceMetadata(t *testing.T) {
	Convey("Given a resource URI and file size", t, func() {
		uri := "path/to/testfile.csv"
		fileSize := zebedee.FileSize{Size: 2048}

		Convey("When mapping to upload service metadata", func() {
			metadata, err := MapResourceToUploadServiceMetadata(uri, fileSize)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the metadata fields are mapped correctly", func() {
				So(metadata.Title, ShouldEqual, "testfile.csv")
				So(metadata.SizeInBytes, ShouldEqual, int64(2048))
				So(metadata.Type, ShouldEqual, "text/csv")
				So(metadata.IsPublishable, ShouldNotBeNil)
				So(*metadata.IsPublishable, ShouldBeTrue)
			})
		})
	})

	Convey("Given a resource URI with an unsupported file type", t, func() {
		uri := "path/to/unsupportedfile.xyz"
		fileSize := zebedee.FileSize{Size: 1024}

		Convey("When mapping to upload service metadata", func() {
			_, err := MapResourceToUploadServiceMetadata(uri, fileSize)

			Convey("Then an unsupported distribution format error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, appErrors.ErrUnsupportedDistributionFormat)
			})
		})
	})
}

func TestMapUploadServiceMetadataToDistribution(t *testing.T) {
	Convey("Given valid upload service metadata", t, func() {
		uploadMetadata := uploadAPI.Metadata{
			Path:        "upload/path/to/file.csv",
			SizeInBytes: 4096,
			Title:       "file.csv",
			Type:        "text/csv",
		}
		Convey("When mapping to dataset distribution", func() {
			distribution, err := MapUploadServiceMetadataToDistribution(uploadMetadata)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the distribution fields are mapped correctly", func() {
				So(distribution.DownloadURL, ShouldEqual, "upload/path/to/file.csv")
				So(distribution.ByteSize, ShouldEqual, int64(4096))
				So(distribution.Title, ShouldEqual, "file.csv")
				So(distribution.Format, ShouldEqual, datasetModels.DistributionFormatCSV)
				So(distribution.MediaType, ShouldEqual, datasetModels.DistributionMediaTypeCSV)
			})
		})
	})
}
