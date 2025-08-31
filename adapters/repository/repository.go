package repository

import (
	"encoding/json"
	"log"
	"net/http"

	"studentgit.kata.academy/Zhodaran/go-kata/core/entity"
	"studentgit.kata.academy/Zhodaran/go-kata/core/usecase"
)

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
	if err := usecase.Register(&user); err != nil {
		log.Printf("Error registering user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user entity.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Вызов функции usecase.Login и обработка результата
	tokenString, err := usecase.Login(&user)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if tokenString == "" {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Успешный вход
	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(entity.TokenResponse{Token: tokenString})
}
