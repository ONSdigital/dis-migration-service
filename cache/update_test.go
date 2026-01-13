package cache

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ONSdigital/dp-topic-api/models"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	apiError "github.com/ONSdigital/dp-topic-api/sdk/errors"
	topicMock "github.com/ONSdigital/dp-topic-api/sdk/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

const testServiceAuthToken = "test-token"

func TestUpdateTopicCache(t *testing.T) {
	Convey("Given a mock topic API client that returns topics", t, func() {
		ctx := context.Background()
		serviceAuthToken := testServiceAuthToken

		// Create test topic data
		rootTopicID := "root-topic-1"
		subTopicID := "sub-topic-1"
		rootTopicTitle := "Economy"
		subTopicTitle := "Inflation"

		rootTopicItems := []models.TopicResponse{
			{
				ID: rootTopicID,
			},
		}

		subtopicIDs := []string{subTopicID}
		rootTopic := models.Topic{
			ID:          rootTopicID,
			Title:       rootTopicTitle,
			Slug:        "economy",
			SubtopicIds: &subtopicIDs,
		}

		subTopic := models.Topic{
			ID:    subTopicID,
			Title: subTopicTitle,
			Slug:  "inflation",
		}

		mockTopicClient := &topicMock.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders topicCli.Headers) (*models.PrivateSubtopics, apiError.Error) {
				return &models.PrivateSubtopics{
					PrivateItems: &rootTopicItems,
				}, nil
			},
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders topicCli.Headers, id string) (*models.TopicResponse, apiError.Error) {
				if id == rootTopicID {
					return &models.TopicResponse{
						ID:      rootTopicID,
						Current: &rootTopic,
					}, nil
				}
				if id == subTopicID {
					return &models.TopicResponse{
						ID:      subTopicID,
						Current: &subTopic,
					}, nil
				}
				return nil, nil
			},
		}

		Convey("When UpdateTopicCache is called", func() {
			updateFunc := UpdateTopicCache(ctx, serviceAuthToken, mockTopicClient)
			result := updateFunc()

			Convey("Then it should return a populated topic cache", func() {
				So(result, ShouldNotBeNil)
				So(result.ID, ShouldEqual, TopicCacheKey)
				So(result.List, ShouldNotBeNil)

				Convey("And the cache should contain the root topic", func() {
					economy, exists := result.List.Get("economy")
					So(exists, ShouldBeTrue)
					So(economy.ID, ShouldEqual, rootTopicID)
					So(economy.Slug, ShouldEqual, "economy")
				})

				Convey("And the cache should contain the subtopic", func() {
					inflation, exists := result.List.Get("inflation")
					So(exists, ShouldBeTrue)
					So(inflation.ID, ShouldEqual, subTopicID)
					So(inflation.Slug, ShouldEqual, "inflation")
					So(inflation.ParentSlug, ShouldEqual, "economy")
				})
			})
		})
	})

	Convey("Given a mock topic API client that returns an error", t, func() {
		ctx := context.Background()
		serviceAuthToken := testServiceAuthToken

		mockTopicClient := &topicMock.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders topicCli.Headers) (*models.PrivateSubtopics, apiError.Error) {
				return nil, apiError.StatusError{Err: context.DeadlineExceeded}
			},
		}

		Convey("When UpdateTopicCache is called", func() {
			updateFunc := UpdateTopicCache(ctx, serviceAuthToken, mockTopicClient)
			result := updateFunc()

			Convey("Then it should return an empty topic cache", func() {
				So(result, ShouldNotBeNil)
				So(result.List, ShouldNotBeNil)
				So(len(result.List.GetSubtopics()), ShouldEqual, 0)
			})
		})
	})
}

func TestTopicCachePeriodicRefresh(t *testing.T) {
	Convey("Given a topic cache configured with a short update interval", t, func() {
		ctx := context.Background()
		serviceAuthToken := testServiceAuthToken

		// Use a very short interval for testing (100ms)
		updateInterval := 100 * time.Millisecond

		// Counter to track how many times the update function is called
		var mu sync.Mutex
		updateCount := 0
		var updateTimes []time.Time

		// Create mock topic client that tracks calls
		mockTopicClient := &topicMock.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders topicCli.Headers) (*models.PrivateSubtopics, apiError.Error) {
				mu.Lock()
				updateCount++
				updateTimes = append(updateTimes, time.Now())
				mu.Unlock()

				rootTopicItems := []models.TopicResponse{
					{ID: "test-topic"},
				}
				return &models.PrivateSubtopics{
					PrivateItems: &rootTopicItems,
				}, nil
			},
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders topicCli.Headers, id string) (*models.TopicResponse, apiError.Error) {
				topic := models.Topic{
					ID:    id,
					Title: "Test Topic",
					Slug:  "test-topic",
				}
				return &models.TopicResponse{
					ID:      id,
					Current: &topic,
				}, nil
			},
		}

		Convey("When the cache is started with periodic updates", func() {
			// Create topic cache
			topicCache, err := NewTopicCache(ctx, &updateInterval)
			So(err, ShouldBeNil)

			// Add update function
			topicCache.AddUpdateFunc(
				topicCache.GetTopicCacheKey(),
				UpdateTopicCache(ctx, serviceAuthToken, mockTopicClient),
			)

			// Start updates in background
			cacheErrorChan := make(chan error, 1)
			updateCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			go topicCache.StartUpdates(updateCtx, cacheErrorChan)

			// Wait for multiple update cycles
			time.Sleep(350 * time.Millisecond)

			// Cancel the updates
			cancel()
			time.Sleep(50 * time.Millisecond) // Give time for cleanup

			Convey("Then the update function should be called multiple times", func() {
				mu.Lock()
				count := updateCount
				times := make([]time.Time, len(updateTimes))
				copy(times, updateTimes)
				mu.Unlock()

				So(count, ShouldBeGreaterThanOrEqualTo, 3)
				So(count, ShouldBeLessThanOrEqualTo, 5)

				Convey("And the updates should occur at regular intervals", func() {
					So(len(times), ShouldBeGreaterThanOrEqualTo, 3)

					// Check intervals between updates (should be ~100ms)
					for i := 1; i < len(times); i++ {
						interval := times[i].Sub(times[i-1])
						// Allow 50ms tolerance due to scheduling
						So(interval, ShouldBeGreaterThanOrEqualTo, 50*time.Millisecond)
						So(interval, ShouldBeLessThanOrEqualTo, 200*time.Millisecond)
					}
				})
			})

			Convey("And the cache should contain the latest data", func() {
				cachedTopic, err := topicCache.GetData(ctx, TopicCacheKey)
				So(err, ShouldBeNil)
				So(cachedTopic, ShouldNotBeNil)

				testTopic, exists := cachedTopic.List.Get("test-topic")
				So(exists, ShouldBeTrue)
				So(testTopic.ID, ShouldEqual, "test-topic")
			})
		})
	})
}

func TestTopicCacheRefreshWithChangingData(t *testing.T) {
	Convey("Given a topic cache that receives different data on each update", t, func() {
		ctx := context.Background()
		serviceAuthToken := testServiceAuthToken
		updateInterval := 100 * time.Millisecond

		var mu sync.Mutex
		callCount := 0

		mockTopicClient := &topicMock.ClienterMock{
			GetRootTopicsPrivateFunc: func(ctx context.Context, reqHeaders topicCli.Headers) (*models.PrivateSubtopics, apiError.Error) {
				mu.Lock()
				callCount++
				currentCount := callCount
				mu.Unlock()
				// Return different topic on each call
				topicID := "topic-" + string(rune('0'+currentCount))
				rootTopicItems := []models.TopicResponse{
					{ID: topicID},
				}
				return &models.PrivateSubtopics{
					PrivateItems: &rootTopicItems,
				}, nil
			},
			GetTopicPrivateFunc: func(ctx context.Context, reqHeaders topicCli.Headers, id string) (*models.TopicResponse, apiError.Error) {
				topic := models.Topic{
					ID:    id,
					Title: "Topic " + id,
					Slug:  id,
				}
				return &models.TopicResponse{
					ID:      id,
					Current: &topic,
				}, nil
			},
		}

		Convey("When the cache refreshes periodically", func() {
			topicCache, err := NewTopicCache(ctx, &updateInterval)
			So(err, ShouldBeNil)

			topicCache.AddUpdateFunc(
				topicCache.GetTopicCacheKey(),
				UpdateTopicCache(ctx, serviceAuthToken, mockTopicClient),
			)

			cacheErrorChan := make(chan error, 1)
			updateCtx, cancel := context.WithCancel(ctx)
			defer cancel()

			go topicCache.StartUpdates(updateCtx, cacheErrorChan)

			// Wait for first update
			time.Sleep(150 * time.Millisecond)

			// Get data after first update
			firstData, err := topicCache.GetData(ctx, TopicCacheKey)
			So(err, ShouldBeNil)
			firstTopics := firstData.List.GetSubtopics()
			firstTopicSlug := firstTopics[0].Slug

			// Wait for second update
			time.Sleep(150 * time.Millisecond)

			// Get data after second update
			secondData, err := topicCache.GetData(ctx, TopicCacheKey)
			So(err, ShouldBeNil)
			secondTopics := secondData.List.GetSubtopics()
			secondTopicSlug := secondTopics[0].Slug

			cancel()

			Convey("Then the cache should contain different data after refresh", func() {
				So(firstTopicSlug, ShouldNotEqual, secondTopicSlug)
				mu.Lock()
				count := callCount
				mu.Unlock()
				So(count, ShouldBeGreaterThanOrEqualTo, 2)
			})
		})
	})
}
