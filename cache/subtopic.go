package cache

import (
	"sync"
	"time"
)

// Subtopics contains a list of subtopics in map form with mutex
// locking. The subtopicsMap is used to keep a record of subtopics
// to be later used to generate the subtopics id query for a topic
// and to check if the subtopic id given by a user exists
type Subtopics struct {
	mutex        *sync.RWMutex
	subtopicsMap map[string]Subtopic
}

// Subtopic represents the data which is cached for a subtopic to be
// used by the dis-migration-service
type Subtopic struct {
	ID              string
	LocaliseKeyName string
	Slug            string
	ReleaseDate     *time.Time
	// This is a reference to the parent topic
	ParentID   string
	ParentSlug string
}

// NewSubTopicsMap creates a new subtopics id map to store subtopic ids
// with mutex locking
func NewSubTopicsMap() *Subtopics {
	return &Subtopics{
		mutex:        &sync.RWMutex{},
		subtopicsMap: make(map[string]Subtopic),
	}
}

// Get returns subtopic for given key (slug)
func (sts *Subtopics) Get(key string) (Subtopic, bool) {
	sts.mutex.RLock()
	defer sts.mutex.RUnlock()

	subtopic, exists := sts.subtopicsMap[key]
	return subtopic, exists
}

// GetSubtopics returns an array of subtopics
func (sts *Subtopics) GetSubtopics() (subtopics []Subtopic) {
	if sts.subtopicsMap == nil {
		return
	}

	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	for _, subtopic := range sts.subtopicsMap {
		subtopics = append(subtopics, subtopic)
	}

	return subtopics
}

// AppendSubtopicID appends the subtopic to the map stored in
// Subtopics with consideration to mutex locking
func (sts *Subtopics) AppendSubtopicID(slug string, subtopic Subtopic) {
	sts.mutex.Lock()
	defer sts.mutex.Unlock()

	if sts.subtopicsMap == nil {
		sts.subtopicsMap = make(map[string]Subtopic)
	}

	sts.subtopicsMap[slug] = subtopic
}
