package executor

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/domain"
)

// JobExecutor defines the contract for job migration operations
//
//go:generate moq -out mock/job_executor.go -pkg mock . JobExecutor
type JobExecutor interface {
	Migrate(ctx context.Context, job *domain.Job) error
	Publish(ctx context.Context, job *domain.Job) error
	PostPublish(ctx context.Context, job *domain.Job) error
	Revert(ctx context.Context, job *domain.Job) error
}

// TaskExecutor defines the contract for task migration operations
//
//go:generate moq -out mock/task_executor.go -pkg mock . TaskExecutor
type TaskExecutor interface {
	Migrate(ctx context.Context, task *domain.Task) error
	Publish(ctx context.Context, task *domain.Task) error
	PostPublish(ctx context.Context, task *domain.Task) error
	Revert(ctx context.Context, task *domain.Task) error
}
