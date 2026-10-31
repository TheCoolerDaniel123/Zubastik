package entity

import (
	"time"

	"github.com/google/uuid"
)

type ToothNote struct {
	ID          uuid.UUID
	DoctorUUID  uuid.UUID
	PatientUUID uuid.UUID
	ToothIndex  int16
	Note        string
	CreatedAt   time.Time
}
