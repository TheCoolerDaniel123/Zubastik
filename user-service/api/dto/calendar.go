package dto

type CreateCalendarEventBody struct {
	Date        string `json:"date" validate:"required"`
	Time        string `json:"time" validate:"required"`
	Description string `json:"description"`
	Status      string `json:"status" validate:"required,oneof=VISITED_DOCTOR CONDITION_CHANGE_PROBLEM CONDITION_CHANGE_SOLUTION BRUSH_CHANGE PLANNED_VISIT"`
}
