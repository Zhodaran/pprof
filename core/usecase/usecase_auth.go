package usecase

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"studentgit.kata.academy/Zhodaran/go-kata/core/entity"
)

const tokenFilePath = "tokens.json"

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
func Register(user *entity.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	entity.Users[user.Username] = entity.User{
		Username: user.Username,
		Password: string(hashedPassword),
	}
	return nil
	// Отправляем ответ о успешной регистрации
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
func Login(user *entity.User) (string, error) {
	// Получаем хешированный пароль пользователя из мапы users
	storedUser, exists := entity.Users[user.Username]
	if !exists || bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)) != nil {
		return "", nil // Возвращаем пустую строку, если пользователь не найден или пароль неверный
	}

	// Если авторизация успешна, создаем токен
	claims := map[string]interface{}{
		"user_id": user.Username, // Используем username как user_id
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	_, tokenString, err := entity.TokenAuth.Encode(claims)
	if err != nil {
		return "", err
	}

	entity.Tokens[tokenString] = struct{}{}

	// Сохраняем токены в файл
	if err := SaveTokens(); err != nil {
		return "", err
	}

	fmt.Println(tokenString)
	return tokenString, nil // Возвращаем токен
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

func SaveTokens() error {
	file, err := json.MarshalIndent(entity.Tokens, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(tokenFilePath, file, 0644)
}
