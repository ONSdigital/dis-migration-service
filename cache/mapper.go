package cache

import (
	"context"
	"net/url"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"
)

// ExtractSingleTopicSlugFromURI extracts the singular topic slug before
// "datasets". It breaks the URI into segements, and returns the topic slug.
func ExtractSingleTopicSlugFromURI(ctx context.Context, uri string, topicCache *TopicCache) string {
	uriSegments := extractURISegments(uri)
	for i, segment := range uriSegments {
		if segment == "datasets" && i > 0 {
			// This is to validate that the topic slug exists in the topic cache.
			topic, err := topicCache.GetTopicBySlug(ctx, uriSegments[i-1])
			if err != nil {
				log.Info(ctx, "topic slug not found in cache", log.Data{
					"slug": uriSegments[i-1],
				})
				return ""
			}
			return topic.Slug
		}
	}
	log.Info(ctx, "no topic slug found in URI", log.Data{"uri": uri})
	return ""
}

// ExtractTopicIDFromURI extracts the singular topic ID before
// "datasets". It breaks the URI into segements, and matches
// the topic against the topic cache, and returns the topic
// ID.
func ExtractTopicIDFromURI(ctx context.Context, uri string, topicCache *TopicCache) []string {
	// Set to track unique topic IDs
	uniqueTopics := make(map[string]struct{})

	// Parse and extract URI segments
	uriSegments := extractURISegments(uri)

	logData := log.Data{
		"uri":          uri,
		"uri_segments": uriSegments,
	}
	log.Info(ctx, "extracting topics from URI segments", logData)

	// Add topic from the single URI segment immediately before "datasets"
	topicSlug := ExtractSingleTopicSlugFromURI(ctx, uri, topicCache)
	if topicSlug != "" {
		topic, err := topicCache.GetTopicBySlug(ctx, topicSlug)
		if err != nil {
			log.Info(ctx, "topic slug not found in cache, skipping", log.Data{
				"slug": topicSlug,
			})
		} else {
			uniqueTopics[topic.ID] = struct{}{}
			log.Info(ctx, "added topic from URI segment", log.Data{
				"topic_id": topic.ID,
			})
		}
	}

	// Convert set to slice
	topicIDs := make([]string, 0, len(uniqueTopics))
	for topicID := range uniqueTopics {
		topicIDs = append(topicIDs, topicID)
	}

	log.Info(ctx, "extracted topic IDs from URI", log.Data{
		"uri":             uri,
		"topic_ids":       topicIDs,
		"topic_ids_count": len(topicIDs),
	})

	return topicIDs
}

// extractURISegments parses a URI and returns the path segments
// For example: "https://www.ons.gov.uk/economy/..." returns:
// ["economy", "inflationandpriceindices", "datasets"]
func extractURISegments(uri string) []string {
	// Parse the URL
	parsedURL, err := url.Parse(uri)
	if err != nil {
		// If URL parsing fails, try treating it as a path
		return extractPathSegments(uri)
	}

	return extractPathSegments(parsedURL.Path)
}

// extractPathSegments splits a path into segments, filtering out empty strings
func extractPathSegments(path string) []string {
	segments := strings.Split(path, "/")
	var result []string

	for _, segment := range segments {
		if segment != "" {
			result = append(result, segment)
		}
	}

	return result
}
