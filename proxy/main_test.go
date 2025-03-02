package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Мок-функция для тестирования
func mockGeocodeAddress(lat, lng string) ([]*Address, error) {
	return []*Address{
		{Street: "Example Street", City: "Example City", State: "Example State", ZipCode: "12345", Country: "Example Country"},
	}, nil
}

func TestGeocodeAddress(t *testing.T) {
	// Создаем тестовый сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что запрос был POST
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Проверяем, что тело запроса содержит правильные данные
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if body["lat"] != "55.7558" || body["lng"] != "37.6173" {
			t.Errorf("Expected lat=55.7558 and lng=37.6173, got lat=%s and lng=%s", body["lat"], body["lng"])
		}

		// Возвращаем успешный ответ
		response := struct {
			Data []Address `json:"data"`
		}{
			Data: []Address{
				{Street: "Example Street", City: "Example City", State: "Example State", ZipCode: "12345", Country: "Example Country"},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer ts.Close()

	// Заменяем URL на тестовый сервер
	dadataAPIkey = "mocked_api_key" // Убедитесь, что ваш ключ не используется в тестах
	originalURL := "https://dadata.ru/api/v2/geocode"
	defer func() { originalURL = "https://dadata.ru/api/v2/geocode" }() // Восстанавливаем оригинальный URL
	originalURL = ts.URL // Заменяем на тестовый сервер

	// Тестируем функцию
	addresses, err := geocodeAddress("55.7558", "37.6173")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(addresses) == 0 {
		t.Fatal("Expected at least one address, got none")
	}

	expectedAddress := addresses[0]
	if expectedAddress.Street != "Example Street" {
		t.Errorf("Expected street to be 'Example Street', got '%s'", expectedAddress.Street)
	}
}
