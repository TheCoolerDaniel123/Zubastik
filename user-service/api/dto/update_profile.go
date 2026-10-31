package dto

type UpdateProfileBody struct {
	Name   *string `json:"name" validate:"omitempty,min=1,max=100"`
	Age    *int    `json:"age" validate:"omitempty,min=0,max=120"`
	Gender *string `json:"gender" validate:"omitempty,oneof=Male Female"`
}

type ToothUpdateItem struct {
	ToothState string  `json:"tooth_state" validate:"required,oneof=Healthy Problematic Missing"`
	ToothNote  *string `json:"tooth_note"`
}

type UpdateTeethBody struct {
	Teeth []ToothUpdateItem `json:"teeth" validate:"required,len=32,dive"`
}

type UpdateBrushScheduleBody struct {
	TimesBrush []string `json:"times_brush" validate:"required,dive,datetime=15:04"`
}
