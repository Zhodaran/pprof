package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	NetPprof "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	_ "studentgit.kata.academy/Zhodaran/go-kata/docs"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"studentgit.kata.academy/Zhodaran/go-kata/controller"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/auth"
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

type GeocodeRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type RequestAddressSearch struct {
	Query string `json:"query"`
}

// TokenResponse представляет ответ с токеном

// LoginResponse представляет ответ при успешном входе

type Server struct {
	http.Server
	cache *Cache
}

type Cache struct {
	data  map[string]interface{}
	mutex sync.RWMutex
	ttl   time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		data: make(map[string]interface{}),
		ttl:  ttl,
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = value
	time.AfterFunc(c.ttl, func() {
		c.Remove(key)
	})
}

// Get получает значение из кэша по ключу
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

func (s *Server) Serve() {
	log.Println("Starting server...")
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: &v", err)
	}
}

func (c *Cache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

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
	cache := NewCache(5 * time.Minute) // Создаем кэш с TTL 5 минут

	r := router(resp, geoService, cache)
	srv := &Server{
		Server: http.Server{
			Addr:         ":8080",
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}

	go func() {
		err := http.ListenAndServe(":6060", nil) // исправлено на ":6060"
		if err != nil {
			panic(err) // обработка ошибки
		}
	}()
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

func proxyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			next.ServeHTTP(w, r)
			return
		}
		proxyURL, _ := url.Parse("http://hugo:1313")
		proxy := httputil.NewSingleHostReverseProxy(proxyURL)
		proxy.ServeHTTP(w, r)
	})
}

func TokenAuthMiddleware(resp controller.Responder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				resp.ErrorUnauthorized(w, errors.New("missing authorization token"))
				return
			}

			token = strings.TrimPrefix(token, "Bearer ")

			_, err := auth.TokenAuth.Decode(token)
			if err != nil {
				resp.ErrorUnauthorized(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func geocodeHandler(resp controller.Responder, geoService service.GeoProvider, cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req GeocodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			resp.ErrorBadRequest(w, err)
			return
		}

		cacheKey := fmt.Sprintf("geocode:%f:%f", req.Lat, req.Lng)

		// Проверяем кэш
		if cachedGeo, found := cache.Get(cacheKey); found {
			resp.OutputJSON(w, cachedGeo)
			return
		}

		geo, err := geoService.GetGeoCoordinatesGeocode(req.Lat, req.Lng)
		if err != nil {
			resp.ErrorInternal(w, err)
			return
		}
		cache.Set(cacheKey, geo)
		resp.OutputJSON(w, geo)
	}
}

func searchHandler(resp controller.Responder, geoService service.GeoProvider, cache *Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RequestAddressSearch
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			resp.ErrorBadRequest(w, err)
			return
		}

		cacheKey := fmt.Sprintf("search:%s", req.Query)

		if cachedGeo, found := cache.Get(cacheKey); found {
			resp.OutputJSON(w, cachedGeo)
			return
		}

		geo, err := geoService.GetGeoCoordinatesAddress(req.Query)
		if err != nil {
			resp.ErrorInternal(w, err)
			return
		}
		cache.Set(cacheKey, geo)
		resp.OutputJSON(w, geo)
	}
}

func router(resp controller.Responder, geoService service.GeoProvider, cache *Cache) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(proxyMiddleware)
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Post("/api/register", auth.Register)
	r.Post("/api/login", auth.Login)
	r.With(TokenAuthMiddleware(resp)).Post("/api/address/geocode", geocodeHandler(resp, geoService, cache))
	r.With(TokenAuthMiddleware(resp)).Post("/api/address/search", searchHandler(resp, geoService, cache))

	r.Mount("/debug/pprof/", http.HandlerFunc(NetPprof.Index))
	r.Handle("/debug/pprof/cmdline", http.HandlerFunc(NetPprof.Cmdline))
	r.Handle("/debug/pprof/profile", http.HandlerFunc(NetPprof.Profile))
	r.Handle("/debug/pprof/symbol", http.HandlerFunc(NetPprof.Symbol))
	r.Handle("/debug/pprof/trace", http.HandlerFunc(NetPprof.Trace))
	r.Handle("/debug/pprof/allocs", http.HandlerFunc(NetPprof.Handler("allocs").ServeHTTP))
	r.Handle("/debug/pprof/block", http.HandlerFunc(NetPprof.Handler("block").ServeHTTP))
	r.Handle("/debug/pprof/goroutine", http.HandlerFunc(NetPprof.Handler("goroutine").ServeHTTP))
	r.Handle("/debug/pprof/heap", http.HandlerFunc(NetPprof.Handler("heap").ServeHTTP))
	r.Handle("/debug/pprof/threadcreate", http.HandlerFunc(NetPprof.Handler("threadcreate").ServeHTTP))
	r.Handle("/debug/pprof/mutex", http.HandlerFunc(NetPprof.Handler("mutex").ServeHTTP))

	return r
}
