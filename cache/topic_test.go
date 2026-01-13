package cache

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewTopicCache(t *testing.T) {
	Convey("Given a context and update interval", t, func() {
		ctx := context.Background()

		Convey("When creating a topic cache with an update interval", func() {
			interval := 10 * time.Minute
			topicCache, err := NewTopicCache(ctx, &interval)

			Convey("Then it should be created successfully", func() {
				So(err, ShouldBeNil)
				So(topicCache, ShouldNotBeNil)
				So(topicCache.Cache, ShouldNotBeNil)
			})
		})

		Convey("When creating a topic cache with nil update interval", func() {
			topicCache, err := NewTopicCache(ctx, nil)

			Convey("Then it should be created successfully", func() {
				So(err, ShouldBeNil)
				So(topicCache, ShouldNotBeNil)
			})
		})
	})
}

func TestTopicCacheGetData(t *testing.T) {
	Convey("Given a topic cache with data", t, func() {
		ctx := context.Background()
		topicCache, _ := NewTopicCache(ctx, nil)

		testTopic := &Topic{
			ID:              "test-topic-id",
			LocaliseKeyName: "Test Topic",
			Slug:            "test-slug",
			List:            NewSubTopicsMap(),
		}

		topicCache.Set("test-key", testTopic)

		Convey("When getting existing data", func() {
			retrieved, err := topicCache.GetData(ctx, "test-key")

			Convey("Then it should return the topic", func() {
				So(err, ShouldBeNil)
				So(retrieved, ShouldNotBeNil)
				So(retrieved.ID, ShouldEqual, "test-topic-id")
				So(retrieved.LocaliseKeyName, ShouldEqual, "Test Topic")
			})
		})

		Convey("When getting non-existent data", func() {
			retrieved, err := topicCache.GetData(ctx, "non-existent")

			Convey("Then it should return an error and empty topic", func() {
				So(err, ShouldNotBeNil)
				So(retrieved, ShouldNotBeNil)
				So(retrieved.ID, ShouldBeEmpty)
			})
		})

		Convey("When getting data with wrong type", func() {
			topicCache.Set("wrong-type", "not-a-topic")
			retrieved, err := topicCache.GetData(ctx, "wrong-type")

			Convey("Then it should return an error and empty topic", func() {
				So(err, ShouldNotBeNil)
				So(retrieved, ShouldNotBeNil)
				So(retrieved.ID, ShouldBeEmpty)
			})
		})

		Convey("When getting data with nil value", func() {
			topicCache.Set("nil-key", nil)
			retrieved, err := topicCache.GetData(ctx, "nil-key")

			Convey("Then it should return an error and empty topic", func() {
				So(err, ShouldNotBeNil)
				So(retrieved, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a nil topic cache", t, func() {
		ctx := context.Background()
		var topicCache *TopicCache

		Convey("When trying to get data", func() {
			retrieved, err := topicCache.GetData(ctx, "test-key")

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(retrieved, ShouldBeNil)
			})
		})
	})

	Convey("Given a topic cache with nil underlying cache", t, func() {
		ctx := context.Background()
		topicCache := &TopicCache{Cache: nil}

		Convey("When trying to get data", func() {
			retrieved, err := topicCache.GetData(ctx, "test-key")

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(retrieved, ShouldBeNil)
			})
		})
	})
}

func TestTopicCacheAddUpdateFunc(t *testing.T) {
	Convey("Given a topic cache", t, func() {
		ctx := context.Background()
		topicCache, _ := NewTopicCache(ctx, nil)

		Convey("When adding an update function", func() {
			called := false
			updateFunc := func() *Topic {
				called = true
				return &Topic{
					ID:   "updated-topic",
					List: NewSubTopicsMap(),
				}
			}

			topicCache.AddUpdateFunc("test-update", updateFunc)

			Convey("Then the function should be stored", func() {
				So(topicCache.UpdateFuncs, ShouldContainKey, "test-update")
			})

			Convey("And when the update function is called", func() {
				updateFuncInterface := topicCache.UpdateFuncs["test-update"]
				result, err := updateFuncInterface()

				Convey("Then it should execute successfully", func() {
					So(err, ShouldBeNil)
					So(called, ShouldBeTrue)
					So(result, ShouldNotBeNil)

					topic, ok := result.(*Topic)
					So(ok, ShouldBeTrue)
					So(topic.ID, ShouldEqual, "updated-topic")
				})
			})
		})
	})
}

func TestGetEmptyTopic(t *testing.T) {
	Convey("When getting an empty topic", t, func() {
		emptyTopic := GetEmptyTopic()

		Convey("Then it should return a valid empty topic", func() {
			So(emptyTopic, ShouldNotBeNil)
			So(emptyTopic.ID, ShouldBeEmpty)
			So(emptyTopic.LocaliseKeyName, ShouldBeEmpty)
			So(emptyTopic.Slug, ShouldBeEmpty)
			So(emptyTopic.ReleaseDate, ShouldBeNil)
			So(emptyTopic.List, ShouldNotBeNil)
			So(len(emptyTopic.List.GetSubtopics()), ShouldEqual, 0)
		})
	})
}

func TestTopicCacheGetTopic(t *testing.T) {
	Convey("Given a topic cache with subtopics", t, func() {
		ctx := context.Background()
		topicCache, _ := NewTopicCache(ctx, nil)

		// Create topic with subtopics
		subtopicsMap := NewSubTopicsMap()
		subtopicsMap.AppendSubtopicID("economy", Subtopic{
			ID:         "economy-id",
			Slug:       "economy",
			ParentSlug: "",
		})
		subtopicsMap.AppendSubtopicID("inflation", Subtopic{
			ID:         "inflation-id",
			Slug:       "inflation",
			ParentSlug: "economy",
		})

		testTopic := &Topic{
			ID:   TopicCacheKey,
			List: subtopicsMap,
		}
		topicCache.Set(TopicCacheKey, testTopic)

		Convey("When getting an existing topic by slug", func() {
			retrieved, err := topicCache.GetTopic(ctx, "economy")

			Convey("Then it should return the topic", func() {
				So(err, ShouldBeNil)
				So(retrieved, ShouldNotBeNil)
				So(retrieved.ID, ShouldEqual, "economy-id")
				So(retrieved.Slug, ShouldEqual, "economy")
			})
		})

		Convey("When getting a non-existent topic by slug", func() {
			retrieved, err := topicCache.GetTopic(ctx, "non-existent")

			Convey("Then it should return an error", func() {
				So(err, ShouldNotBeNil)
				So(retrieved, ShouldBeNil)
			})
		})
	})
}

func TestGetTopicCacheKey(t *testing.T) {
	Convey("Given a topic cache", t, func() {
		ctx := context.Background()
		topicCache, _ := NewTopicCache(ctx, nil)

		Convey("When getting the topic cache key", func() {
			key := topicCache.GetTopicCacheKey()

			Convey("Then it should return the constant value", func() {
				So(key, ShouldEqual, TopicCacheKey)
			})
		})
	})
}
