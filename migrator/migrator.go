package migrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out mock/migrator.go -pkg mock . Migrator
type Migrator interface {
	Migrate(ctx context.Context, job *domain.Job)
	Shutdown(ctx context.Context) error
}

type migrator struct {
	executor TaskExecutor
	wg       sync.WaitGroup
}

func NewDefaultMigrator() Migrator {
	return &migrator{
		executor: &defaultExecutor{},
	}
}

func (mig *migrator) Migrate(ctx context.Context, job *domain.Job) {
	mig.wg.Add(1)
	go func() {
		defer mig.wg.Done()
		err := mig.executor.Migrate(ctx, job)
		if err != nil {
			log.Error(ctx, "error executing job migration", err, log.Data{"jobID": job.ID})
		}
	}()
}

func (mig *migrator) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		mig.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		err := fmt.Errorf("timed out waiting for background tasks to complete")
		log.Error(ctx, "error shutting down migrator", err)
		return err
	}
}
