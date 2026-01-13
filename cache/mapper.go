package cache

import (
	"context"
	"net/url"
	"strings"

	"github.com/ONSdigital/log.go/v2/log"
)

// ExtractTopicIDsFromURI extracts topic IDs from a URI by matching
// URI segments to topics in the cache. It breaks the URI into
// segments, matches them against the topic cache, and returns a
// deduplicated list of topic IDs. Existing topic IDs can be provided
// and will be preserved in the result.
func ExtractTopicIDsFromURI(ctx context.Context, uri string, existingTopicIDs []string, topicCache *TopicCache) []string {
	// Set to track unique topic IDs
	uniqueTopics := make(map[string]struct{})

	// Add existing topic IDs
	for _, topicID := range existingTopicIDs {
		uniqueTopics[topicID] = struct{}{}
	}

	// Parse and extract URI segments
	uriSegments := extractURISegments(uri)

	logData := log.Data{
		"uri":          uri,
		"uri_segments": uriSegments,
	}
	log.Info(ctx, "extracting topics from URI segments", logData)

	// Add topics based on URI segments (up to 3 segments)
	parentSlug := ""
	for i, segment := range uriSegments {
		if i >= 3 {
			// Only process up to the first 3 segments as per requirements
			break
		}
		addTopicWithParents(ctx, segment, parentSlug, topicCache, uniqueTopics)
		parentSlug = segment // Update parentSlug for the next iteration
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

// addTopicWithParents adds a topic and its parents to uniqueTopics
// map if they don't already exist. It recursively adds parent topics
// until it reaches the root topic.
func addTopicWithParents(ctx context.Context, slug, parentSlug string, topicCache *TopicCache, uniqueTopics map[string]struct{}) {
	topic, err := topicCache.GetTopic(ctx, slug)
	if err != nil {
		// Topic not found in cache, skip
		log.Info(ctx, "topic not found in cache, skipping", log.Data{
			"slug":        slug,
			"parent_slug": parentSlug,
		})
		return
	}

	// Check if this topic has already been processed
	if _, alreadyProcessed := uniqueTopics[topic.ID]; alreadyProcessed {
		return
	}

	// Verify the parent slug matches (or is empty for root topics)
	if parentSlug != "" && topic.ParentSlug != parentSlug {
		log.Info(ctx, "parent slug mismatch, skipping topic", log.Data{
			"slug":               slug,
			"expected_parent":    parentSlug,
			"actual_parent_slug": topic.ParentSlug,
		})
		return
	}

	uniqueTopics[topic.ID] = struct{}{}
	log.Info(ctx, "added topic from URI segment", log.Data{
		"topic_id":    topic.ID,
		"slug":        slug,
		"parent_slug": topic.ParentSlug,
	})

	// Recursively add the parent topic if it exists
	if topic.ParentSlug != "" {
		addTopicWithParents(ctx, topic.ParentSlug, "", topicCache, uniqueTopics)
	}
}
