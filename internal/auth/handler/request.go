package handler

type RegisterUserRequest struct {
	Login string	`json:"login" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

