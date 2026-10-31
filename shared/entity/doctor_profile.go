package entity

import "github.com/google/uuid"

type DoctorProfile struct {
	UserUUID        uuid.UUID
	Name            string
	Specialization  string
	ExperienceYears int
	Rating          float64
	PhotoURL        string
}
