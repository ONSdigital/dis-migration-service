package domain

import (
	"strconv"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
)

const (
	// CollectionNamePrefix is the prefix given to
	// collections created in zebedee, i.e. "Migration Collection for Job 1"
	CollectionNamePrefix = "Migration Collection for Job"
)

// NewMigrationCollection creates a new zebedee.Collection
// for a given job number.
func NewMigrationCollection(jobNumber int) zebedee.Collection {
	return zebedee.Collection{
		Name: CollectionNamePrefix + " " + strconv.Itoa(jobNumber),
		Type: zebedee.CollectionTypeAutomated,
	}
}
