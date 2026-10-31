package dto

type CreateToothNoteBody struct {
	ToothIndex int16  `json:"tooth_index" validate:"min=0,max=31"`
	Note       string `json:"note" validate:"required,min=1,max=1000"`
}

type UpdateToothNoteBody struct {
	Note string `json:"note" validate:"required,min=1,max=1000"`
}

type CreateRecommendationBody struct {
	Text string `json:"text" validate:"required,min=1,max=2000"`
}
