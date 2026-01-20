package mapper

import (
	"context"
	"testing"

	"github.com/ONSdigital/dis-migration-service/cache"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMapDatasetLandingPageToDatasetAPI(t *testing.T) {
	Convey("Given a Zebedee dataset landing page", t, func() {
		ctx := context.Background()
		pageData := getTestDatasetLandingPage()
		topicCache, _ := cache.NewPopulatedTopicCacheForTest(ctx)

		Convey("When it is mapped to a Dataset API dataset", func() {
			dataset, err := MapDatasetLandingPageToDatasetAPI(ctx, "test-dataset-id", pageData, topicCache)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the dataset fields are mapped correctly", func() {
				So(dataset.ID, ShouldEqual, "test-dataset-id")
				So(dataset.Title, ShouldEqual, "Test Dataset Title")
				So(dataset.Description, ShouldEqual, "This is a summary of the test dataset.")
				So(len(dataset.Contacts), ShouldEqual, 1)
				So(dataset.Contacts[0].Name, ShouldEqual, "John Doe")
				So(dataset.Contacts[0].Email, ShouldEqual, "john.doe@example.com")
				So(dataset.Contacts[0].Telephone, ShouldEqual, "123-456-7890")
				So(dataset.Keywords, ShouldResemble, []string{"test", "dataset", "sample"})
				So(dataset.NextRelease, ShouldEqual, "2024-12-31")
				So(dataset.QMI, ShouldNotBeNil)
				So(dataset.QMI.HRef, ShouldEqual, "/methodology/qmi/test-qmi")
				So(dataset.QMI.Title, ShouldEqual, "Test QMI Title")
				So(dataset.QMI.Description, ShouldEqual, "This is a summary of the test QMI.")
				So(dataset.License, ShouldEqual, "Open Government Licence v3.0")
			})
		})
	})

	Convey("Given a Zebedee page that is not a dataset landing page", t, func() {
		ctx := context.Background()
		pageData := zebedee.DatasetLandingPage{
			Type: "unknown type",
		}
		topicCache, _ := cache.NewTopicCache(ctx, nil)

		Convey("When it is mapped to a Dataset API dataset", func() {
			dataset, err := MapDatasetLandingPageToDatasetAPI(ctx, "test-dataset-id", pageData, topicCache)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "invalid page type for dataset landing page")
			})

			Convey("Then the dataset is nil", func() {
				So(dataset, ShouldBeNil)
			})
		})
	})

	Convey("Given a nil topicCache", t, func() {
		ctx := context.Background()
		pageData := getTestDatasetLandingPage()

		Convey("When mapping is attempted", func() {
			dataset, err := MapDatasetLandingPageToDatasetAPI(ctx, "test-dataset-id", pageData, nil)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "topicCache is required for dataset mapping")
			})

			Convey("Then the dataset is nil", func() {
				So(dataset, ShouldBeNil)
			})
		})
	})

	Convey("Given a Zebedee dataset landing page with a URI containing topic segments", t, func() {
		ctx := context.Background()
		pageData := zebedee.DatasetLandingPage{
			Type: zebedee.PageTypeDatasetLandingPage,
			URI:  "/economy/inflationandpriceindices/datasets/consumerpriceinflation",
			Description: zebedee.Description{
				Title:   "Consumer Price Inflation",
				Summary: "Price indices, percentage changes and weights for different product groupings.",
				Contact: zebedee.Contact{
					Name:  "Test Contact",
					Email: "test@example.com",
				},
			},
		}

		Convey("When a populated topic cache is provided", func() {
			topicCache, _ := cache.NewTopicCache(ctx, nil)

			// Populate the cache with test topics
			subtopicsMap := cache.NewSubTopicsMap()
			subtopicsMap.AppendSubtopicID("economy", cache.Subtopic{
				ID:         "1234",
				Slug:       "economy",
				ParentSlug: "",
			})
			subtopicsMap.AppendSubtopicID("inflationandpriceindices", cache.Subtopic{
				ID:         "5678",
				Slug:       "inflationandpriceindices",
				ParentSlug: "economy",
			})
			testTopic := &cache.Topic{
				ID:   cache.TopicCacheKey,
				List: subtopicsMap,
			}
			topicCache.Set(cache.TopicCacheKey, testTopic)

			dataset, err := MapDatasetLandingPageToDatasetAPI(ctx, "test-dataset-id", pageData, topicCache)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then topics are extracted from the URI and mapped correctly", func() {
				So(dataset, ShouldNotBeNil)
				So(dataset.Topics, ShouldNotBeEmpty)
				So(len(dataset.Topics), ShouldBeGreaterThan, 0)

				// Verify that topic IDs from the URI are included
				topicMap := make(map[string]bool)
				for _, topicID := range dataset.Topics {
					topicMap[topicID] = true
				}

				// Should contain the economy topic
				So(topicMap["1234"], ShouldBeTrue)
				// Should contain the inflationandpriceindices topic
				So(topicMap["5678"], ShouldBeTrue)
			})
		})

		Convey("When topic cache has no matching topics for URI segments", func() {
			topicCache, _ := cache.NewTopicCache(ctx, nil)

			// Populate cache with topics that don't match the URI segments
			subtopicsMap := cache.NewSubTopicsMap()
			subtopicsMap.AppendSubtopicID("business", cache.Subtopic{
				ID:         "9999",
				Slug:       "business",
				ParentSlug: "",
			})
			testTopic := &cache.Topic{
				ID:   cache.TopicCacheKey,
				List: subtopicsMap,
			}
			topicCache.Set(cache.TopicCacheKey, testTopic)

			dataset, err := MapDatasetLandingPageToDatasetAPI(ctx, "test-dataset-id", pageData, topicCache)

			Convey("Then an error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "no topics found for dataset - datasets must have at least one topic")
			})

			Convey("Then the dataset is nil", func() {
				So(dataset, ShouldBeNil)
			})
		})

		Convey("When Zebedee data contains existing topics and canonical topic", func() {
			topicCache, _ := cache.NewTopicCache(ctx, nil)

			// Populate cache with test topics
			subtopicsMap := cache.NewSubTopicsMap()
			subtopicsMap.AppendSubtopicID("economy", cache.Subtopic{
				ID:         "1234",
				Slug:       "economy",
				ParentSlug: "",
			})
			subtopicsMap.AppendSubtopicID("inflationandpriceindices", cache.Subtopic{
				ID:         "5678",
				Slug:       "inflationandpriceindices",
				ParentSlug: "economy",
			})
			subtopicsMap.AppendSubtopicID("businessindustryandtrade", cache.Subtopic{
				ID:         "2468",
				Slug:       "businessindustryandtrade",
				ParentSlug: "",
			})
			subtopicsMap.AppendSubtopicID("employmentandlabourmarket", cache.Subtopic{
				ID:         "1357",
				Slug:       "employmentandlabourmarket",
				ParentSlug: "",
			})
			testTopic := &cache.Topic{
				ID:   cache.TopicCacheKey,
				List: subtopicsMap,
			}
			topicCache.Set(cache.TopicCacheKey, testTopic)

			// Create page data with existing topics in Zebedee
			pageDataWithTopics := zebedee.DatasetLandingPage{
				Type: zebedee.PageTypeDatasetLandingPage,
				URI:  "/economy/inflationandpriceindices/datasets/consumerpriceinflation",
				Description: zebedee.Description{
					Title:   "Consumer Price Inflation",
					Summary: "Price indices for different product groupings.",
					Contact: zebedee.Contact{
						Name:  "Test Contact",
						Email: "test@example.com",
					},
					Topics:         []string{"2468"}, // Existing secondary topic (4-digit ID)
					CanonicalTopic: "1357",           // Canonical topic (4-digit ID)
				},
			}

			dataset, err := MapDatasetLandingPageToDatasetAPI(ctx, "test-dataset-id", pageDataWithTopics, topicCache)

			Convey("Then no error is returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then topics from Zebedee and URI are merged without duplicates", func() {
				So(dataset, ShouldNotBeNil)
				So(dataset.Topics, ShouldNotBeEmpty)

				// Build a map for easier checking
				topicMap := make(map[string]bool)
				for _, topicID := range dataset.Topics {
					topicMap[topicID] = true
				}

				// Should contain URI-derived topics
				So(topicMap["1234"], ShouldBeTrue)
				So(topicMap["5678"], ShouldBeTrue)

				// Should contain existing Zebedee topics
				So(topicMap["2468"], ShouldBeTrue)

				// Should contain canonical topic
				So(topicMap["1357"], ShouldBeTrue)

				// Should have exactly 4 unique topics
				So(len(dataset.Topics), ShouldEqual, 4)
			})
		})
	})

	Convey("Given a mock topic cache (feature flag disabled)", t, func() {
		ctx := context.Background()
		mockCache, _ := cache.NewMockTopicCache(ctx)

		Convey("When mapping a dataset with URI that has no matching topics", func() {
			pageData := getTestDatasetLandingPage()
			pageData.URI = "/some/path/with/no/matching/topics"

			dataset, err := MapDatasetLandingPageToDatasetAPI(ctx, "test-dataset-123", pageData, mockCache)

			Convey("Then no error is returned as validation is skipped for mock cache", func() {
				So(err, ShouldBeNil)
				So(dataset, ShouldNotBeNil)
			})

			Convey("And the dataset has an empty topics list", func() {
				So(len(dataset.Topics), ShouldEqual, 0)
			})
		})
	})
}

func getTestDatasetLandingPage() zebedee.DatasetLandingPage {
	return zebedee.DatasetLandingPage{
		Type: zebedee.PageTypeDatasetLandingPage,
		URI:  "/economy/datasets/test-dataset",
		Description: zebedee.Description{
			Title:   "Test Dataset Title",
			Summary: "This is a summary of the test dataset.",
			Contact: zebedee.Contact{
				Name:      "John Doe",
				Email:     "john.doe@example.com",
				Telephone: "123-456-7890",
			},
			Keywords:    []string{"test", "dataset", "sample"},
			NextRelease: "2024-12-31",
		},
		RelatedMethodology: []zebedee.Link{
			{
				URI:     "/methodology/qmi/test-qmi",
				Title:   "Test QMI Title",
				Summary: "This is a summary of the test QMI.",
			},
		},
		Section: zebedee.Section{
			Title:    "Usage Notes",
			Markdown: "These are the usage notes for the dataset.",
		},
	}
}
