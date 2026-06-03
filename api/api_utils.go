package api

import (
	"context"

	"github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/log.go/v2/log"
)

func handleAuthEntityDataError(ctx context.Context, err error, logData log.Data) error {
	errors.NewAuditEventError(ctx, err, errors.GetAuthEntityDataError, errors.GetAuthEntityDataErrorDescription, logData)
	return errors.ErrFailedToParseAuthEntityData
}
