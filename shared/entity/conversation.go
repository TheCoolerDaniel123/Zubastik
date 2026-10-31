package entity

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID          uuid.UUID
	DoctorUUID  uuid.UUID
	PatientUUID uuid.UUID
	CreatedAt   time.Time
}
