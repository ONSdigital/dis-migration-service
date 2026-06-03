package api

import (
	"context"
	"fmt"

	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dp-authorisation/v2/authorisation"
	"github.com/ONSdigital/log.go/v2/log"
)

// logAuditEvent produces protective monitoring logging for API endpoints.
// Failed request audit events include the reason for the failure.
func logAuditEvent(ctx context.Context, message string, authEntityData *authorisation.AuthEntityData, action domain.Action,
	endpoint string, outcome domain.Outcome, errReason string, auditEventParams *domain.AuditEventParams) {
	identityType := log.USER
	data := log.Data{
		"action":   action,
		"endpoint": endpoint,
		"outcome":  outcome,
	}

	if errReason != "" {
		data["reason"] = errReason
	}

	if authEntityData != nil {
		if authEntityData.IsServiceAuth {
			identityType = log.SERVICE
		}
		log.Info(
			ctx,
			message,
			log.Classification(log.ProtectiveMonitoring),
			log.Auth(identityType, authEntityData.EntityData.UserID),
			data,
		)
		return
	}

	if auditEventParams != nil {
		for _, value := range *auditEventParams {
			auditEventParamsString := fmt.Sprint(value)
			log.Info(
				ctx,
				message,
				log.Classification(log.ProtectiveMonitoring),
				log.Auth(identityType, auditEventParamsString),
				data,
			)
		}
	}
}
