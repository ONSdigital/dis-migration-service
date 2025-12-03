package migrator

import (
	"context"
	"fmt"
	"sync"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/log.go/v2/log"
)

// Migrator defines the contract for migration operations
//
//go:generate moq -out mock/migrator.go -pkg mock . Migrator
type Migrator interface {
	Migrate(ctx context.Context, job *domain.Job)
	Shutdown(ctx context.Context) error
}

type migrator struct {
	JobService application.JobService
	Clients    *clients.ClientList
	wg         sync.WaitGroup
}

// NewDefaultMigrator creates a new default migrator with the
// provided job service and clients
func NewDefaultMigrator(jobService application.JobService, appClients *clients.ClientList) Migrator {
	//ctx := context.Background()
	//err := jobService.CreateJobNumberCounter(ctx)
	//if err != nil {
	//	log.Error(ctx, "error creating job number counter", err)
	//}

	return &migrator{
		JobService: jobService,
		Clients:    appClients,
	}
}

// Migrate starts the migration process for the given job
func (mig *migrator) Migrate(ctx context.Context, job *domain.Job) {
	mig.wg.Add(1)
	go func() {
		defer mig.wg.Done()
		// TODO: account for job type here.
		err := mig.migrateStaticDataset(ctx, job)
		if err != nil {
			log.Error(ctx, "error executing job migration", err, log.Data{"jobID": job.ID})
		}
	}()
}

// Shutdown waits for all ongoing migrations to complete or times out
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
