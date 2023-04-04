package models

type SignUp struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Mobile    string `json:"mobile" validate:"numeric,len=10"`
	Password  string `json:"password" validate:"required,min=8,max=20"`
}

type SignIn struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}