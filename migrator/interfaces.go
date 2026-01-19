package migrator

import (
	"context"
)

// Migrator defines the contract for migration operations
//
//go:generate moq -out mock/migrator.go -pkg mock . Migrator
type Migrator interface {
	Shutdown(ctx context.Context) error
	Start(ctx context.Context)
}
