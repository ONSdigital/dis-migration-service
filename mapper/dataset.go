package mapper

import (
	"context"
	"errors"
	"strings"

	"github.com/ONSdigital/dis-migration-service/cache"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
)

// MapDatasetLandingPageToDatasetAPI maps a Zebedee dataset landing page
// to a Dataset API dataset model. It extracts topic IDs from the URI and
// merges them with any existing topics from the page data.
func MapDatasetLandingPageToDatasetAPI(ctx context.Context, datasetID string, pageData zebedee.DatasetLandingPage, topicCache *cache.TopicCache) (*datasetModels.Dataset, error) {
	if pageData.Type != zebedee.PageTypeDatasetLandingPage {
		return nil, errors.New("invalid page type for dataset landing page")
	}

	if topicCache == nil {
		return nil, errors.New("topicCache is required for dataset mapping")
	}

	// Extract topic IDs from the URI and merge with any existing topics from Zebedee data
	// Zebedee data contains topicIDs in Description.Topics (secondaryTopics) and CanonicalTopic
	existingTopicIDs := pageData.Description.Topics
	if pageData.Description.CanonicalTopic != "" {
		existingTopicIDs = append(existingTopicIDs, pageData.Description.CanonicalTopic)
	}

	topicIDs := cache.ExtractTopicIDsFromURI(ctx, pageData.URI, existingTopicIDs, topicCache)

	if len(topicIDs) == 0 {
		return nil, errors.New("no topics found for dataset - datasets must have at least one topic")
	}

	// TODO: confirm handling of NextRelease field if unset.
	nextRelease := "To be confirmed"
	if pageData.Description.NextRelease != "" {
		nextRelease = pageData.Description.NextRelease
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
		License:  domain.OpenGovernmentLicence,
		// Warning: NextRelease is a string in both Zebedee and Dataset API.
		NextRelease: nextRelease,
		QMI:         getQMILink(pageData.RelatedMethodology),
		Title:       pageData.Description.Title,
		Topics:      topicIDs,
		Type:        datasetModels.Static.String(),
	}

	return dataset, nil
}

func getQMILink(methodologyLinks []zebedee.Link) *datasetModels.GeneralDetails {
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
