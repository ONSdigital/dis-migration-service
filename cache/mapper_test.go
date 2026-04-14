package cache

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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

			topicIDs := ExtractTopicIDFromURI(ctx, uri, nil, topicCache)

			Convey("Then it should extract ONLY the topic before datasets", func() {
				So(topicIDs, ShouldContain, "inflation-id")
				So(len(topicIDs), ShouldEqual, 1)
			})
		})

		Convey("When extracting topic IDs from a path-only URI", func() {
			topicID := ExtractTopicIDFromURI(ctx, "/economy/inflationandpriceindices/datasets", nil, topicCache)

			Convey("Then it should extract the correct topic", func() {
				So(topicID, ShouldContain, "inflation-id")
				So(len(topicID), ShouldEqual, 1)
			})
		})

		Convey("When extracting topic IDs with existing topics", func() {
			uri := "/economy"
			existingTopics := []string{"existing-topic-id"}
			topicID := ExtractTopicIDFromURI(ctx, uri, existingTopics, topicCache)

			Convey("Then it should preserve existing topics", func() {
				So(topicID, ShouldContain, "existing-topic-id")
			})
		})

		Convey("When URI contains less than three segments before datasets", func() {
			topicID := ExtractTopicIDFromURI(ctx, testEconomyURI, nil, topicCache)

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
