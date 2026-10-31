package dto

type SendMessageBody struct {
	Message string `json:"message" validate:"required,min=1,max=2000"`
}

type GreetBody struct {
	Language string `json:"language" validate:"required,min=2,max=10"`
}
