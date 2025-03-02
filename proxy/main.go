package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/http/pprof"
	"net/url"
	"os"
	"os/signal"
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

	go srv.Serve()

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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

	// Используем обработчики с middleware
	r.With(TokenAuthMiddleware(resp)).Post("/api/address/geocode", geocodeHandler(resp, geoService, cache))
	r.With(TokenAuthMiddleware(resp)).Post("/api/address/search", searchHandler(resp, geoService, cache))
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/", http.HandlerFunc(pprof.Index))
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))

	// @Summary Профилирование CPU
	// @Description Получить профиль CPU
	// @Tags pprof
	// @Produce json
	// @Success 200 {object} string
	// @Router /mycustompath/pprof/profile [get]
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/profile", http.HandlerFunc(pprof.Profile))
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/trace", http.HandlerFunc(pprof.Trace))
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/allocs", http.HandlerFunc(pprof.Handler("allocs").ServeHTTP))
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/block", http.HandlerFunc(pprof.Handler("block").ServeHTTP))

	// @Summary Профилирование горутин
	// @Description Получить профиль горутин
	// @Tags pprof
	// @Produce json
	// @Success 200 {object} string
	// @Router /mycustompath/pprof/goroutine [get]
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/goroutine", http.HandlerFunc(pprof.Handler("goroutine").ServeHTTP))

	// @Summary Профилирование памяти
	// @Description Получить профиль памяти
	// @Tags pprof
	// @Produce json
	// @Success 200 {object} string
	// @Router /mycustompath/pprof/heap [get]
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/heap", http.HandlerFunc(pprof.Handler("heap").ServeHTTP))
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/threadcreate", http.HandlerFunc(pprof.Handler("threadcreate").ServeHTTP))
	r.With(TokenAuthMiddleware(resp)).Get("/mycustompath/pprof/mutex", http.HandlerFunc(pprof.Handler("mutex").ServeHTTP))
	return r
}
