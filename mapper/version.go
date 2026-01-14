package mapper

import (
	"errors"

	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
)

// MapDatasetVersionToDatasetAPI maps a Zebedee dataset version to a
// Dataset API dataset model.
func MapDatasetVersionToDatasetAPI(editionID string, pageData zebedee.Dataset, seriesData zebedee.DatasetLandingPage, editionData zebedee.Dataset) (*datasetModels.Version, error) {
	if pageData.Type != zebedee.PageTypeDataset {
		return nil, errors.New("invalid page type for dataset version page")
	}

	distributions, err := mapDownloadsToDistributions(pageData.Downloads)
	if err != nil {
		return nil, err
	}

	version := &datasetModels.Version{
		Distributions: &distributions,
		Edition:       editionID,
		EditionTitle:  pageData.Description.Edition,
		Version:       getVersion(pageData.Versions),
		ReleaseDate:   pageData.Description.ReleaseDate,
		Type:          clients.DatasetVersionTypeStatic,
	}

	if pageData.Description.NationalStatistic {
		version.QualityDesignation = datasetModels.QualityDesignationAccreditedOfficial
	}

	if version.Version > 1 {
		correctionNotice := getCorrectionNotice(editionData, version.Version)
		if correctionNotice != "" {
			version.Alerts = &[]datasetModels.Alert{
				{
					Type:        datasetModels.AlertTypeCorrection,
					Description: correctionNotice,
				},
			}
		}
	}

	if seriesData.Section.Markdown != "" {
		version.UsageNotes = &[]datasetModels.UsageNote{
			{
				Title: seriesData.Section.Title,
				Note:  seriesData.Section.Markdown,
			},
		}
	}

	return version, nil
}

func mapDownloadsToDistributions(downloads []zebedee.Download) ([]datasetModels.Distribution, error) {
	distributions := make([]datasetModels.Distribution, 0, len(downloads))

	for _, download := range downloads {
		distributionFormat, err := mapFileNameToDistributionFormat(download.File)
		if err != nil {
			return nil, err
		}

		distribution := datasetModels.Distribution{
			Title:  download.File,
			Format: distributionFormat,
		}
		distributions = append(distributions, distribution)
	}

	return distributions, nil
}

func mapFileNameToDistributionFormat(fileName string) (datasetModels.DistributionFormat, error) {
	mimeType := DeriveMimeTypeFromFilename(fileName)
	return MapMimeTypeToDistributionFormat(mimeType)
}

func getVersion(versions []zebedee.Version) int {
	if len(versions) > 0 {
		return len(versions) + 1
	} else {
		return 1
	}
}

func getCorrectionNotice(editionData zebedee.Dataset, version int) string {
	if editionData.Versions != nil && len(editionData.Versions) >= version {
		return editionData.Versions[version-1].Notice
	}
	return ""
}
