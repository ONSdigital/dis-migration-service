package cache

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSubTopicsMap(t *testing.T) {
	Convey("When creating a new subtopics map", t, func() {
		subtopics := NewSubTopicsMap()

		Convey("Then it should be initialized with empty map", func() {
			So(subtopics, ShouldNotBeNil)
			So(subtopics.mutex, ShouldNotBeNil)
			So(subtopics.subtopicsMap, ShouldNotBeNil)
			So(len(subtopics.subtopicsMap), ShouldEqual, 0)
		})
	})
}

func TestSubtopicsAppendSubtopicID(t *testing.T) {
	Convey("Given an empty subtopics map", t, func() {
		subtopics := NewSubTopicsMap()

		Convey("When appending a subtopic", func() {
			testSubtopic := Subtopic{
				ID:              "test-id",
				LocaliseKeyName: "Test Topic",
				Slug:            "test-topic",
				ParentID:        "parent-id",
				ParentSlug:      "parent-slug",
			}
			subtopics.AppendSubtopicID("test-topic", testSubtopic)

			Convey("Then the subtopic should be stored", func() {
				retrieved, exists := subtopics.Get("test-topic")
				So(exists, ShouldBeTrue)
				So(retrieved.ID, ShouldEqual, "test-id")
				So(retrieved.Slug, ShouldEqual, "test-topic")
				So(retrieved.ParentID, ShouldEqual, "parent-id")
			})
		})

		Convey("When appending multiple subtopics", func() {
			subtopic1 := Subtopic{ID: "id1", Slug: "slug1"}
			subtopic2 := Subtopic{ID: "id2", Slug: "slug2"}
			subtopic3 := Subtopic{ID: "id3", Slug: "slug3"}

			subtopics.AppendSubtopicID("slug1", subtopic1)
			subtopics.AppendSubtopicID("slug2", subtopic2)
			subtopics.AppendSubtopicID("slug3", subtopic3)

			Convey("Then all subtopics should be retrievable", func() {
				retrieved1, exists1 := subtopics.Get("slug1")
				retrieved2, exists2 := subtopics.Get("slug2")
				retrieved3, exists3 := subtopics.Get("slug3")

				So(exists1, ShouldBeTrue)
				So(exists2, ShouldBeTrue)
				So(exists3, ShouldBeTrue)
				So(retrieved1.ID, ShouldEqual, "id1")
				So(retrieved2.ID, ShouldEqual, "id2")
				So(retrieved3.ID, ShouldEqual, "id3")
			})
		})

		Convey("When overwriting an existing subtopic", func() {
			original := Subtopic{ID: "original-id", Slug: "test-slug"}
			updated := Subtopic{ID: "updated-id", Slug: "test-slug"}

			subtopics.AppendSubtopicID("test-slug", original)
			subtopics.AppendSubtopicID("test-slug", updated)

			Convey("Then the subtopic should be updated", func() {
				retrieved, exists := subtopics.Get("test-slug")
				So(exists, ShouldBeTrue)
				So(retrieved.ID, ShouldEqual, "updated-id")
			})
		})
	})

	Convey("Given a nil subtopics map", t, func() {
		subtopics := &Subtopics{
			mutex:        &sync.RWMutex{},
			subtopicsMap: nil,
		}

		Convey("When appending a subtopic", func() {
			testSubtopic := Subtopic{ID: "test-id", Slug: "test-slug"}
			subtopics.AppendSubtopicID("test-slug", testSubtopic)

			Convey("Then it should initialize the map and store the subtopic", func() {
				retrieved, exists := subtopics.Get("test-slug")
				So(exists, ShouldBeTrue)
				So(retrieved.ID, ShouldEqual, "test-id")
			})
		})
	})
}

func TestSubtopicsGet(t *testing.T) {
	Convey("Given a subtopics map with data", t, func() {
		subtopics := NewSubTopicsMap()
		testSubtopic := Subtopic{
			ID:              "test-id",
			LocaliseKeyName: "Test Topic",
			Slug:            "test-slug",
		}
		subtopics.AppendSubtopicID("test-slug", testSubtopic)

		Convey("When getting an existing subtopic", func() {
			retrieved, exists := subtopics.Get("test-slug")

			Convey("Then it should return the subtopic and true", func() {
				So(exists, ShouldBeTrue)
				So(retrieved.ID, ShouldEqual, "test-id")
				So(retrieved.LocaliseKeyName, ShouldEqual, "Test Topic")
			})
		})

		Convey("When getting a non-existent subtopic", func() {
			retrieved, exists := subtopics.Get("non-existent")

			Convey("Then it should return false", func() {
				So(exists, ShouldBeFalse)
				So(retrieved.ID, ShouldBeEmpty)
			})
		})
	})
}

func TestSubtopicsGetSubtopics(t *testing.T) {
	Convey("Given a subtopics map with multiple entries", t, func() {
		subtopics := NewSubTopicsMap()
		releaseDate := time.Now()

		subtopic1 := Subtopic{
			ID:              "id1",
			Slug:            "slug1",
			LocaliseKeyName: "Topic 1",
			ReleaseDate:     &releaseDate,
		}
		subtopic2 := Subtopic{
			ID:              "id2",
			Slug:            "slug2",
			LocaliseKeyName: "Topic 2",
		}
		subtopic3 := Subtopic{
			ID:              "id3",
			Slug:            "slug3",
			LocaliseKeyName: "Topic 3",
			ParentID:        "parent-id",
			ParentSlug:      "parent-slug",
		}

		subtopics.AppendSubtopicID("slug1", subtopic1)
		subtopics.AppendSubtopicID("slug2", subtopic2)
		subtopics.AppendSubtopicID("slug3", subtopic3)

		Convey("When getting all subtopics", func() {
			allSubtopics := subtopics.GetSubtopics()

			Convey("Then it should return all subtopics", func() {
				So(len(allSubtopics), ShouldEqual, 3)

				// Check that all IDs are present
				ids := make(map[string]bool)
				for _, st := range allSubtopics {
					ids[st.ID] = true
				}
				So(ids["id1"], ShouldBeTrue)
				So(ids["id2"], ShouldBeTrue)
				So(ids["id3"], ShouldBeTrue)
			})
		})
	})

	Convey("Given an empty subtopics map", t, func() {
		subtopics := NewSubTopicsMap()

		Convey("When getting all subtopics", func() {
			allSubtopics := subtopics.GetSubtopics()

			Convey("Then it should return an empty slice", func() {
				So(len(allSubtopics), ShouldEqual, 0)
			})
		})
	})

	Convey("Given a nil subtopics map", t, func() {
		subtopics := &Subtopics{
			mutex:        &sync.RWMutex{},
			subtopicsMap: nil,
		}

		Convey("When getting all subtopics", func() {
			allSubtopics := subtopics.GetSubtopics()

			Convey("Then it should return nil", func() {
				So(allSubtopics, ShouldBeNil)
			})
		})
	})
}
