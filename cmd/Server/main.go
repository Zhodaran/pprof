package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"runtime/trace"
	"syscall"
	"time"

	"go.uber.org/zap"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/controller"
	myhttp "studentgit.kata.academy/Zhodaran/go-kata/internal/delivery/http"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/entity"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/service"
)

// @title Address API
// @version 1.0
// @description API для поиска
// @host localhost:8080
// @BasePath
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @RequestAddressSearch представляет запрос для поиска
// @Description Этот эндпоинт позволяет получить адрес по наименованию
// @Param address body ResponseAddress true "Географические координаты"

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cpuProfileFile, err := os.Create("profile.prof")
	if err != nil {
		log.Fatal("Не удалось создать файл для CPU-профиля:", err)
	}
	defer cpuProfileFile.Close()

	// Начинаем профилирование CPU
	if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
		log.Fatal("Не удалось начать CPU-профилирование:", err)
	}
	defer pprof.StopCPUProfile()

	traceFile, err := os.Create("trace.out")
	if err != nil {
		log.Fatal("Не удалось создать файл для трассировки:", err)
	}
	defer traceFile.Close()

	// Начинаем трассировку
	if err := trace.Start(traceFile); err != nil {
		log.Fatal("Не удалось начать трассировку:", err)
	}
	defer trace.Stop()

	geoService := service.NewGeoService("d9e0649452a137b73d941aa4fb4fcac859372c8c", "ec99b849ebf21277ec821c63e1a2bc8221900b1d") // Создаем новый экземпляр GeoService
	resp := controller.NewResponder(logger)
	cache := entity.NewCache(5 * time.Minute) // Создаем кэш с TTL 5 минут

	r := myhttp.Router(resp, geoService, cache)
	srv := &entity.Server{
		Server: http.Server{
			Addr:         ":8080",
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

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

	// --- Конец Health Check Endpoint ---

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
