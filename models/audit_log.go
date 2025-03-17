package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdminID   string             `bson:"admin_id" json:"admin_id"`
	Action    string             `bson:"action" json:"action"`
	TargetID  string             `bson:"target_id" json:"target_id"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
}
