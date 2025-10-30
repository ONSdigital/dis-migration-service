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

func NewDefaultMigrator(jobService application.JobService, appClients *clients.ClientList) Migrator {
	return &migrator{
		JobService: jobService,
		Clients:    appClients,
	}
}

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
