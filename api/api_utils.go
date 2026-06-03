package api

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/log.go/v2/log"
)

func handleAuthEntityDataError(ctx context.Context, err error, logData log.Data) *errors.AuditEventError {
	return errors.NewAuditEventError(ctx, err, errors.GetAuthEntityDataError, errors.GetAuthEntityDataErrorDescription, logData)
}
