package healthpoint

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"studentgit.kata.academy/Zhodaran/go-kata/adapters/adapter"
	"studentgit.kata.academy/Zhodaran/go-kata/adapters/repository"
)

func Healthpoint(cache *adapter.Cache, geoService *repository.GeoService) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// 1. Проверка geoService
		if geoService == nil {
			log.Println("geoService is nil")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "geoService is nil")
			return
		}

		// 3.  Более сложная проверка (пример: проверка кэша)
		if cache == nil {
			log.Println("Cache is nil")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Cache is nil")
			return
		}

		//Попытка записи и чтения из кэша
		testKey := "healthcheck_test_key"
		testValue := "healthcheck_test_value"
		cache.Set(testKey, testValue)
		_, found := cache.Get(testKey)
		if !found {
			log.Println("Cache test failed")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintln(w, "Cache test failed")
			return
		}

		// Все проверки прошли успешно
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})
}

// --- Конец Health Check Endpoint ---

func Geopoint(srv *adapter.Server) {
	http.HandleFunc("/geocode", func(w http.ResponseWriter, r *http.Request) {
		// Имитация работы
		time.Sleep(100 * time.Millisecond)
		w.Write([]byte("Geocode response"))
	})

	go srv.Serve()
	for i := 0; i < 100; i++ {
		_, err := http.Get("http://localhost:8080/geocode")
		if err != nil {
			log.Println("Ошибка при запросе:", err)
		}
		time.Sleep(10 * time.Millisecond) // Небольшая задержка между запросами
	}
	time.Sleep(2 * time.Second)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Ошибка при завершении работы: %v\n", err)
	} else {
		log.Println("Server stopped gracefully")
	}
}
