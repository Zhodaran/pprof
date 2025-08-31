package entity

import (
	"github.com/go-chi/jwtauth"
)

type LoginResponse struct {
	Message string `json:"message"`
}

var (
	TokenAuth = jwtauth.New("HS256", []byte("your_secret_key"), nil)
	Users     = make(map[string]User) // Хранение пользователей
	Tokens    = make(map[string]struct{})
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
