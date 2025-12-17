package mapper

import (
	"errors"
	"strings"

	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
)

// MapDatasetLandingPageToDatasetAPI maps a Zebedee dataset landing page to a
// Dataset API dataset model.
func MapDatasetLandingPageToDatasetAPI(datasetID string, pageData zebedee.DatasetLandingPage) (*datasetModels.Dataset, error) {
	if pageData.Type != zebedee.PageTypeDatasetLandingPage {
		return nil, errors.New("invalid page type for dataset landing page")
	}

	dataset := &datasetModels.Dataset{
		Description: pageData.Description.Summary,
		// Zebedee only allows one contact per dataset landing page.
		// Datase API supports multiple contacts.
		Contacts: []datasetModels.ContactDetails{
			{
				Name:      pageData.Description.Contact.Name,
				Email:     pageData.Description.Contact.Email,
				Telephone: pageData.Description.Contact.Telephone,
			},
		},
		ID:       datasetID,
		Keywords: pageData.Description.Keywords,
		License:  domain.OpenGovermentLicence,
		// Warning: NextRelease is a string in both Zebedee and Dataset API.
		NextRelease: pageData.Description.NextRelease,
		QMI:         getQMILink(pageData.RelatedMethodology),
		Title:       pageData.Description.Title,
		Type:        datasetModels.Static.String(),
	}

	return dataset, nil
}

func getQMILink(methodologyLinks []zebedee.Related) *datasetModels.GeneralDetails {
	for _, link := range methodologyLinks {
		if strings.Contains(link.URI, "/qmi/") {
			return &datasetModels.GeneralDetails{
				Description: link.Summary,
				Title:       link.Title,
				HRef:        link.URI,
			}
		}
	}
	return &datasetModels.GeneralDetails{}
}
