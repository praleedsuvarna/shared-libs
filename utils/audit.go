package utils

import (
	"context"
	"time"

	"github.com/praleedsuvarna/shared-libs/config"
	"github.com/praleedsuvarna/shared-libs/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Log an admin action
func LogAudit(adminID, action, targetID string) {
	collection := config.GetCollection("oms_audit_logs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log := models.AuditLog{
		ID:        primitive.NewObjectID(),
		AdminID:   adminID,
		Action:    action,
		TargetID:  targetID,
		Timestamp: time.Now(),
	}

	_, err := collection.InsertOne(ctx, log)
	if err != nil {
		println("Failed to log audit:", err.Error())
	}
}

// GetAuditLogs retrieves audit logs with optional filtering
func GetAuditLogs(filter bson.M) ([]models.AuditLog, error) {
	collection := config.GetCollection("oms_audit_logs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []models.AuditLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}
