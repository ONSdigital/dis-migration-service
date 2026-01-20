package mapper

import (
	"path/filepath"

	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	uploadAPI "github.com/ONSdigital/dp-upload-service/api"
	"github.com/google/uuid"
)

// MapResourceToUploadServiceMetadata maps a Zebedee resource URI
// and file size to an upload service Metadata struct. It derives
// the MIME type from the fileextension. If the MIME type is
// unsupported, it returns an error.
func MapResourceToUploadServiceMetadata(uri string, fileSize zebedee.FileSize) (uploadAPI.Metadata, error) {
	fileMimeType := DeriveMimeTypeFromFilename(uri)

	// Checking for supported distribution format now so we don't map unsupported files.
	_, err := MapMimeTypeToDistributionFormat(fileMimeType)
	if err != nil {
		return uploadAPI.Metadata{}, appErrors.ErrUnsupportedDistributionFormat
	}

	// isPublishable is currently always true for migrated files.
	isPublishable := true
	uploadMetadata := uploadAPI.Metadata{
		Path:          uuid.New().String(),
		IsPublishable: &isPublishable,
		SizeInBytes:   fileSize.Size,
		Title:         filepath.Base(uri),
		Type:          fileMimeType,
		Licence:       domain.OpenGovernmentLicence,
		LicenceUrl:    domain.OpenGovernmentLicenceURL,
	}
	return uploadMetadata, nil
}

// MapUploadServiceMetadataToDistribution maps upload service Metadata to a
// dp-dataset-api Distribution. It maps the MIME type to a DistributionFormat.
// If the MIME type is unsupported, it returns an error.
func MapUploadServiceMetadataToDistribution(uploadMetadata uploadAPI.Metadata) (datasetModels.Distribution, error) {
	distributionFormat, err := MapMimeTypeToDistributionFormat(uploadMetadata.Type)
	if err != nil {
		return datasetModels.Distribution{}, err
	}

	distribution := datasetModels.Distribution{
		ByteSize:    int64(uploadMetadata.SizeInBytes),
		DownloadURL: uploadMetadata.Path,
		Format:      distributionFormat,
		MediaType:   datasetModels.DistributionMediaType(uploadMetadata.Type),
		Title:       uploadMetadata.Title,
	}

	return distribution, nil
}
