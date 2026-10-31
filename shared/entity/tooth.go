package entity

import "github.com/google/uuid"

type ToothType string
type ToothState string

const (
	ToothTypeMolar   ToothType = "Molar"
	ToothTypeMissing ToothType = "Missing"
	ToothTypeMilk    ToothType = "Milk"

	ToothStateHealthy     ToothState = "Healthy"
	ToothStateProblematic ToothState = "Problematic"
	ToothStateMissing     ToothState = "Missing"
)

type Tooth struct {
	UserUUID   uuid.UUID
	ToothIndex int16
	ToothType  ToothType
	ToothState ToothState
	ToothNote  *string
}
