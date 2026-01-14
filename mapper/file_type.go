package mapper

import (
	"path/filepath"
	"strings"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
)

const (
	// ExtensionCSDB is the file extension for CSDB files.
	ExtensionCSDB = ".csdb"
	// ExtensionCSV is the file extension for CSV files.
	ExtensionCSV = ".csv"
	// ExtensionCSVW is the file extension for CSVW files.
	ExtensionCSVW = ".csvw"
	// ExtensionSDMX is the file extension for SDMX files.
	ExtensionSDMX = ".sdmx"
	// ExtensionXLS is the file extension for Microsoft Excel XLS files.
	ExtensionXLS = ".xls"
	// ExtensionXLSX is the file extension for Microsoft Excel XLSX files.
	ExtensionXLSX = ".xlsx"
	// MimeTypeCSDB is the MIME type for CSDB files.
	MimeTypeCSDB = "application/csdb"
	// MimeTypeCSV is the MIME type for CSV files.
	MimeTypeCSV = "text/csv"
	// MimeTypeCSVW is the MIME type for CSVW files.
	MimeTypeCSVW = "application/csvw+json"
	// MimeTypeSDMX is the MIME type for SDMX files.
	MimeTypeSDMX = "application/sdmx+xml"
	// MimeTypeXLS is the MIME type for Microsoft Excel XLS files.
	MimeTypeXLS = "application/vnd.ms-excel"
	// MimeTypeXLSX is the MIME type for Microsoft Excel XLSX files.
	MimeTypeXLSX = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	// MimeTypeOctetStream is the generic MIME type for binary data.
	MimeTypeOctetStream = "application/octet-stream"
)

// MapMimeTypeToDistributionFormat maps a MIME type string to a
// DistributionFormat as used by the dp-dataset-api. If the MIME type is
// unsupported, it returns an error.
func MapMimeTypeToDistributionFormat(mimeType string) (datasetModels.DistributionFormat, error) {
	if format, ok := mimeTypeToDistributionFormat[mimeType]; ok {
		return format, nil
	}
	return "", appErrors.ErrUnsupportedDistributionFormat
}

var extensionToMimeType = map[string]string{
	ExtensionCSV:  MimeTypeCSV,
	ExtensionSDMX: MimeTypeSDMX,
	ExtensionXLS:  MimeTypeXLS,
	ExtensionXLSX: MimeTypeXLSX,
	ExtensionCSDB: MimeTypeCSDB,
	ExtensionCSVW: MimeTypeCSVW,
}

var mimeTypeToDistributionFormat map[string]datasetModels.DistributionFormat = map[string]datasetModels.DistributionFormat{
	MimeTypeCSV:  datasetModels.DistributionFormatCSV,
	MimeTypeSDMX: datasetModels.DistributionFormatSDMX,
	MimeTypeXLS:  datasetModels.DistributionFormatXLS,
	MimeTypeXLSX: datasetModels.DistributionFormatXLSX,
	MimeTypeCSDB: datasetModels.DistributionFormatCSDB,
	MimeTypeCSVW: datasetModels.DistributionFormatCSVWMeta,
}

// DeriveMimeTypeFromFilename returns the MIME type for a given file name.
// If the type cannot be determined, it returns "application/octet-stream".
func DeriveMimeTypeFromFilename(filename string) string {
	extension := strings.ToLower(filepath.Ext(filename))
	if mimeType, ok := extensionToMimeType[extension]; ok {
		return mimeType
	}
	return MimeTypeOctetStream
}
