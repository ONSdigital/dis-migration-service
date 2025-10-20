package migrator

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/log.go/v2/log"
)

//go:generate moq -out mock/executor.go -pkg mock . TaskExecutor
type TaskExecutor interface {
	Migrate(ctx context.Context, job *domain.Job) error
}

type defaultExecutor struct{}

func (e *defaultExecutor) Migrate(ctx context.Context, job *domain.Job) error {
	logData := log.Data{
		"id": job.ID,
	}

	log.Info(ctx, "starting migration for job", logData)

	// TODO: implement real migration steps here.

	log.Info(ctx, "finished migration for job", logData)

	return nil
}
