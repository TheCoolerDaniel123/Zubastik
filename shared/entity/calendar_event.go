package entity

import "github.com/google/uuid"

type CalendarEventStatus string

const (
	CalendarEventStatusVisitedDoctor           CalendarEventStatus = "VISITED_DOCTOR"
	CalendarEventStatusConditionChangeProblem  CalendarEventStatus = "CONDITION_CHANGE_PROBLEM"
	CalendarEventStatusConditionChangeSolution CalendarEventStatus = "CONDITION_CHANGE_SOLUTION"
	CalendarEventStatusBrushChange             CalendarEventStatus = "BRUSH_CHANGE"
	CalendarEventStatusPlannedVisit            CalendarEventStatus = "PLANNED_VISIT"
)

type CalendarEvent struct {
	ID          uuid.UUID
	UserUUID    uuid.UUID
	Date        string
	Time        string
	Description string
	Status      CalendarEventStatus
}
