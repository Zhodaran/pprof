package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/core/entity"
)

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
	var user entity.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if _, exists := entity.Users[user.Username]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	entity.Users[user.Username] = entity.User{
		Username: user.Username,
		Password: string(hashedPassword),
	}

	// Отправляем ответ о успешной регистрации
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
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
	var user entity.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Получаем хешированный пароль пользователя из мапы users
	storedUser, exists := entity.Users[user.Username]
	if !exists || bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Если авторизация успешна, возвращаем статус 200 OK
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entity.LoginResponse{Message: "Login successful"})
	claims := map[string]interface{}{
		"user_id": user.Username, // Используем username как user_id
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	_, tokenString, err := entity.TokenAuth.Encode(claims)
	if err != nil {
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	entity.Tokens[tokenString] = struct{}{}

	// Сохраняем токены в файл
	if err := SaveTokens(); err != nil {
		http.Error(w, "Could not save token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entity.TokenResponse{Token: tokenString})
	fmt.Println(tokenString)
}

const tokenFilePath = "tokens.json"

func SaveTokens() error {
	file, err := json.MarshalIndent(entity.Tokens, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(tokenFilePath, file, 0644)
}

// LoadTokens загружает токены из файла
func LoadTokens() error {
	file, err := os.ReadFile(tokenFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Если файл не существует, просто возвращаем nil
			return nil
		}
		return err
	}

	return json.Unmarshal(file, &entity.Tokens)
}
