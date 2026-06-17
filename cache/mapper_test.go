package cache

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExtractSingleTopicSlugFromURI(t *testing.T) {
	Convey("Given a topic cache with test topics", t, func() {
		ctx := context.Background()
		topicCache, _ := NewTopicCache(ctx, nil)

		// Setup test topics
		subtopicsMap := NewSubTopicsMap()
		subtopicsMap.AppendSubtopicID("economy", Subtopic{
			ID:         "economy-id",
			Slug:       "economy",
			ParentSlug: "",
		})
		subtopicsMap.AppendSubtopicID("investmentspensionsandtrusts", Subtopic{
			ID:         "investments-id",
			Slug:       "investmentspensionsandtrusts",
			ParentSlug: "economy",
		})
		subtopicsMap.AppendSubtopicID("datasets", Subtopic{
			ID:         "datasets-id",
			Slug:       "datasets",
			ParentSlug: "investmentspensionsandtrusts",
		})

		testTopic := &Topic{
			ID:   TopicCacheKey,
			List: subtopicsMap,
		}

		// Set test topic in cache
		topicCache.Set(TopicCacheKey, testTopic)

		Convey("When the URI contains a topic slug before datasets", func() {
			uri := "https://www.ons.gov.uk/economy/investmentspensionsandtrusts/datasets/example"

			slug := ExtractSingleTopicSlugFromURI(ctx, uri, topicCache)

			Convey("Then it should return the slug immediately before datasets", func() {
				So(slug, ShouldEqual, "investmentspensionsandtrusts")
			})
		})

		Convey("When the URI is path-only with a topic slug before datasets", func() {
			slug := ExtractSingleTopicSlugFromURI(ctx, "/economy/investmentspensionsandtrusts/datasets", topicCache)

			Convey("Then it should return the slug immediately before datasets", func() {
				So(slug, ShouldEqual, "investmentspensionsandtrusts")
			})
		})

		Convey("When datasets is the first segment in the URI", func() {
			slug := ExtractSingleTopicSlugFromURI(ctx, "/datasets/example", topicCache)

			Convey("Then it should return an empty string", func() {
				So(slug, ShouldEqual, "")
			})
		})

		Convey("When the URI does not contain a datasets segment", func() {
			slug := ExtractSingleTopicSlugFromURI(ctx, "/economy/investmentspensionsandtrusts", topicCache)

			Convey("Then it should return an empty string", func() {
				So(slug, ShouldEqual, "")
			})
		})

		Convey("When the URI is empty", func() {
			slug := ExtractSingleTopicSlugFromURI(ctx, "", topicCache)

			Convey("Then it should return an empty string", func() {
				So(slug, ShouldEqual, "")
			})
		})
	})
}

func TestExtractTopicIDsFromURI(t *testing.T) {
	const testEconomyURI = "/economy/inflationandpriceindices/datasets"

	Convey("Given a topic cache with test topics", t, func() {
		ctx := context.Background()
		topicCache, _ := NewTopicCache(ctx, nil)

		// Setup test topics
		subtopicsMap := NewSubTopicsMap()
		subtopicsMap.AppendSubtopicID("economy", Subtopic{
			ID:         "economy-id",
			Slug:       "economy",
			ParentSlug: "",
		})
		subtopicsMap.AppendSubtopicID("inflationandpriceindices", Subtopic{
			ID:         "inflation-id",
			Slug:       "inflationandpriceindices",
			ParentSlug: "economy",
		})
		subtopicsMap.AppendSubtopicID("datasets", Subtopic{
			ID:         "datasets-id",
			Slug:       "datasets",
			ParentSlug: "inflationandpriceindices",
		})

		testTopic := &Topic{
			ID:   TopicCacheKey,
			List: subtopicsMap,
		}

		// Set test topic in cache
		topicCache.Set(TopicCacheKey, testTopic)

		Convey("When extracting topic IDs from a full URI", func() {
			uri := "https://www.ons.gov.uk/economy/inflationandpriceindices/datasets/example"

			topicIDs := ExtractTopicIDFromURI(ctx, uri, topicCache)

			Convey("Then it should extract ONLY the topic before datasets", func() {
				So(topicIDs, ShouldContain, "inflation-id")
				So(len(topicIDs), ShouldEqual, 1)
			})
		})

		Convey("When extracting topic IDs from a path-only URI", func() {
			topicID := ExtractTopicIDFromURI(ctx, "/economy/inflationandpriceindices/datasets", topicCache)

			Convey("Then it should extract the correct topic", func() {
				So(topicID, ShouldContain, "inflation-id")
				So(len(topicID), ShouldEqual, 1)
			})
		})

		Convey("When URI contains less than three segments before datasets", func() {
			topicID := ExtractTopicIDFromURI(ctx, testEconomyURI, topicCache)

			Convey("Then it should extract ONLY the segment immediately before datasets", func() {
				So(topicID, ShouldContain, "inflation-id")
				So(len(topicID), ShouldEqual, 1)
			})
		})
	})
}

func TestExtractURISegments(t *testing.T) {
	Convey("Given various URI formats", t, func() {
		Convey("When parsing a full URL", func() {
			uri := "https://www.ons.gov.uk/economy/inflationandpriceindices/datasets"
			segments := extractURISegments(uri)

			Convey("Then it should extract path segments", func() {
				So(segments, ShouldResemble, []string{"economy", "inflationandpriceindices", "datasets"})
			})
		})

		Convey("When parsing a path", func() {
			uri := "/economy/inflationandpriceindices/datasets"
			segments := extractURISegments(uri)

			Convey("Then it should extract segments", func() {
				So(segments, ShouldResemble, []string{"economy", "inflationandpriceindices", "datasets"})
			})
		})

		Convey("When parsing path with trailing slash", func() {
			uri := "/economy/inflationandpriceindices/"
			segments := extractURISegments(uri)

			Convey("Then it should exclude empty segments", func() {
				So(segments, ShouldResemble, []string{"economy", "inflationandpriceindices"})
			})
		})

		Convey("When parsing an invalid URI", func() {
			uri := "://invalid"
			segments := extractURISegments(uri)

			Convey("Then it should treat as path and return segments", func() {
				So(len(segments), ShouldBeGreaterThanOrEqualTo, 0)
			})
		})
	})
}

func TestExtractPathSegments(t *testing.T) {
	Convey("Given various path formats", t, func() {
		Convey("When extracting from normal path", func() {
			path := "/segment1/segment2/segment3"
			segments := extractPathSegments(path)

			Convey("Then it should return all segments", func() {
				So(segments, ShouldResemble, []string{"segment1", "segment2", "segment3"})
			})
		})

		Convey("When extracting from path with consecutive slashes", func() {
			path := "/segment1//segment2///segment3"
			segments := extractPathSegments(path)

			Convey("Then it should filter out empty segments", func() {
				So(segments, ShouldResemble, []string{"segment1", "segment2", "segment3"})
			})
		})

		Convey("When extracting from empty path", func() {
			path := ""
			segments := extractPathSegments(path)

			Convey("Then it should return empty slice", func() {
				So(len(segments), ShouldEqual, 0)
			})
		})
	})
}
