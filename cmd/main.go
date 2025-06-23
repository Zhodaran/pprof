package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"

	"go.uber.org/zap"
	healthpoint "studentgit.kata.academy/Zhodaran/go-kata/internal/Healthpoint"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/controller"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/entity"
	myhttp "studentgit.kata.academy/Zhodaran/go-kata/internal/router/http"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/usecase"
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
	geoService := usecase.NewGeoService("d9e0649452a137b73d941aa4fb4fcac859372c8c", "ec99b849ebf21277ec821c63e1a2bc8221900b1d")
	resp := controller.NewResponder(logger)
	cache := entity.NewCache(5 * time.Minute) // Создаем кэш с TTL 5 минут

	r := myhttp.Router(resp, geoService, cache)

	// Создаем экземпляр entity.Server
	srv := &entity.Server{
		Server: http.Server{
			Addr:         ":8080",
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

	// Запускаем сервер в горутине
	go srv.Serve()
	gracefulShutdown(srv, logger)
	// Передаем экземпляр entity.Server в функции healthpoint
	healthpoint.Healthpoint(cache, geoService)
	healthpoint.Geopoint(srv) // Теперь передаем srv как *entity.Server

	// Блокируем основной поток, чтобы сервер продолжал работать

}

func gracefulShutdown(server *entity.Server, logger *zap.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Info("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Graceful shutdown failed", zap.Error(err))
	} else {
		logger.Info("Server stopped gracefully")
	}
}
