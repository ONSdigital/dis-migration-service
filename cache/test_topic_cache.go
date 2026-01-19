package cache

import "context"

// NewPopulatedTopicCacheForTest creates a topic cache with basic
// test data. This ensures tests never use an empty cache, which
// would be invalid in production.
func NewPopulatedTopicCacheForTest(ctx context.Context) (*TopicCache, error) {
	topicCache, err := NewTopicCache(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Populate cache with common test topics
	subtopicsMap := NewSubTopicsMap()

	// Add root-level topics
	subtopicsMap.AppendSubtopicID("economy", Subtopic{
		ID:         "1234",
		Slug:       "economy",
		ParentSlug: "",
	})

	subtopicsMap.AppendSubtopicID("business", Subtopic{
		ID:         "5678",
		Slug:       "business",
		ParentSlug: "",
	})

	subtopicsMap.AppendSubtopicID("peoplepopulationandcommunity", Subtopic{
		ID:         "9012",
		Slug:       "peoplepopulationandcommunity",
		ParentSlug: "",
	})

	// Add some child topics under economy
	subtopicsMap.AppendSubtopicID("inflationandpriceindices", Subtopic{
		ID:         "3456",
		Slug:       "inflationandpriceindices",
		ParentSlug: "economy",
	})

	subtopicsMap.AppendSubtopicID("grossdomesticproductgdp", Subtopic{
		ID:         "7890",
		Slug:       "grossdomesticproductgdp",
		ParentSlug: "economy",
	})

	testTopic := &Topic{
		ID:   TopicCacheKey,
		List: subtopicsMap,
	}

	topicCache.Set(TopicCacheKey, testTopic)

	return topicCache, nil
}
