package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth"
	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	Message string `json:"message"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	BadRequest      string `json:"400"`
	DadataBad       string `json:"500"`
	SuccefulRequest string `json:"200"`
}

var (
	TokenAuth = jwtauth.New("HS256", []byte("your_secret_key"), nil)
	users     = make(map[string]User) // Хранение пользователей
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// @Summary Register a new user
// @Description This endpoint allows you to register a new user with a username and password.
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "User registration details"
// @Success 201 {object} TokenResponse "User registered successfully"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 409 {object} ErrorResponse "User already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if _, exists := users[user.Username]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	users[user.Username] = User{
		Username: user.Username,
		Password: string(hashedPassword),
	}

	// Используем логин пользователя в качестве user_id

}

// @Summary Login a user
// @Description This endpoint allows a user to log in with their username and password.
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "User login details"
// @Success 200 {object} LoginResponse "Login successful"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Invalid credentials"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/login [post]
func Login(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Получаем хешированный пароль пользователя из мапы users
	storedUser, exists := users[user.Username]
	if !exists || bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Если авторизация успешна, возвращаем статус 200 OK
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{Message: "Login successful"})
	claims := map[string]interface{}{
		"user_id": user.Username, // Используем username как user_id
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	_, tokenString, err := TokenAuth.Encode(claims)
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(TokenResponse{Token: tokenString})
	fmt.Println(tokenString)
}
