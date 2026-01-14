package mapper

import (
	"mime"
	"path/filepath"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
)

// MapMimeTypeToDistributionFormat maps a MIME type string to a
// DistributionFormat as used by the dp-dataset-api. If the MIME type is
// unsupported, it returns an error.
func MapMimeTypeToDistributionFormat(mimeType string) (datasetModels.DistributionFormat, error) {
	switch mimeType {
	case "text/csv":
		return datasetModels.DistributionFormatCSV, nil
	case "application/sdmx+xml":
		return datasetModels.DistributionFormatSDMX, nil
	case "application/vnd.ms-excel":
		return datasetModels.DistributionFormatXLS, nil
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return datasetModels.DistributionFormatXLSX, nil
	case "application/csdb":
		return datasetModels.DistributionFormatCSDB, nil
	case "application/csvw+json":
		return datasetModels.DistributionFormatCSVWMeta, nil
	default:
		return "", appErrors.ErrUnsupportedDistributionFormat
	}
}

// DeriveMimeType returns the MIME type for a given file name.
// If the type cannot be determined, it returns "application/octet-stream".
func DeriveMimeType(filename string) string {
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}
