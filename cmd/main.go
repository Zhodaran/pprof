package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"

	"go.uber.org/zap"
	healthpoint "studentgit.kata.academy/Zhodaran/go-kata/adapters/Healthpoint"
	"studentgit.kata.academy/Zhodaran/go-kata/adapters/adapter"
	"studentgit.kata.academy/Zhodaran/go-kata/adapters/pprof"
	"studentgit.kata.academy/Zhodaran/go-kata/adapters/repository"
	myhttp "studentgit.kata.academy/Zhodaran/go-kata/adapters/router/http"
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
	geoService := repository.NewGeoService("d9e0649452a137b73d941aa4fb4fcac859372c8c", "ec99b849ebf21277ec821c63e1a2bc8221900b1d")
	resp := repository.NewResponder(logger)
	cache := adapter.NewCache(5 * time.Minute) // Создаем кэш с TTL 5 минут

	r := myhttp.Router(resp, geoService, cache)

	// Создаем экземпляр entity.Server
	srv := &adapter.Server{
		Server: http.Server{
			Addr:         ":8080",
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
	pprof.CreatePprof()
	// Запускаем сервер в горутине
	go srv.Serve()
	gracefulShutdown(srv, logger)
	// Передаем экземпляр entity.Server в функции healthpoint
	healthpoint.Healthpoint(cache, geoService)
	healthpoint.Geopoint(srv) // Теперь передаем srv как *entity.Server

	// Блокируем основной поток, чтобы сервер продолжал работать

}

func gracefulShutdown(server *adapter.Server, logger *zap.Logger) {
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
