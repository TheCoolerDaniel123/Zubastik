package entity

import (
	"time"

	"github.com/google/uuid"
)

type Recommendation struct {
	ID          uuid.UUID
	DoctorUUID  uuid.UUID
	PatientUUID uuid.UUID
	Text        string
	CreatedAt   time.Time
}
