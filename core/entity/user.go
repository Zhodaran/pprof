package entity

import (
	"net/http"

	"github.com/go-chi/jwtauth"
)

type Responder interface {
	OutputJSON(w http.ResponseWriter, responseData interface{})
	ErrorUnauthorized(w http.ResponseWriter, err error)
	ErrorBadRequest(w http.ResponseWriter, err error)
	ErrorForbidden(w http.ResponseWriter, err error)
	ErrorInternal(w http.ResponseWriter, err error)
}

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
