package mapper

import (
	"errors"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
)

// MapDatasetVersionToDatasetAPI maps a Zebedee dataset version to a
// Dataset API dataset model.
func MapDatasetVersionToDatasetAPI(editionID string, pageData zebedee.Dataset, seriesData zebedee.DatasetLandingPage, editionData zebedee.Dataset) (*datasetModels.Version, error) {
	if pageData.Type != zebedee.PageTypeDataset {
		return nil, errors.New("invalid page type for dataset version page")
	}

	version := &datasetModels.Version{
		Edition:      editionID,
		EditionTitle: pageData.Description.Edition,
		Version:      getVersion(pageData.Versions),
		ReleaseDate:  pageData.Description.ReleaseDate,
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
