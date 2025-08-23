package http

import (
	"net/http"

	NetPprof "net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/adapter"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/controller"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/core/usecase"
)

func Router(resp controller.Responder, geoService usecase.GeoProvider, cache *adapter.Cache) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes (без авторизации)
	r.Get("/swagger/*", httpSwagger.WrapHandler) // Swagger остаётся публичным

	// API routes
	r.Post("/api/register", usecase.Register)
	r.Post("/api/login", usecase.Login)

	// Protected routes (требуют авторизации)
	r.Group(func(r chi.Router) {
		r.Use(TokenAuthMiddleware(resp))

		// API endpoints
		r.Post("/api/address/geocode", geocodeHandler(resp, geoService, cache))
		r.Post("/api/address/search", searchHandler(resp, geoService, cache))

		// Pprof endpoints
		r.Handle("/mycustompath/pprof/*", http.HandlerFunc(NetPprof.Index))
		r.Handle("/mycustompath/pprof/cmdline", http.HandlerFunc(NetPprof.Cmdline))
		r.Handle("/mycustompath/pprof/profile", http.HandlerFunc(NetPprof.Profile))
		r.Handle("/mycustompath/pprof/symbol", http.HandlerFunc(NetPprof.Symbol))
		r.Handle("/mycustompath/pprof/trace", http.HandlerFunc(NetPprof.Trace))
		r.Handle("/mycustompath/pprof/allocs", NetPprof.Handler("allocs"))
		r.Handle("/mycustompath/pprof/block", NetPprof.Handler("block"))
		r.Handle("/mycustompath/pprof/goroutine", NetPprof.Handler("goroutine"))
		r.Handle("/mycustompath/pprof/heap", NetPprof.Handler("heap"))
		r.Handle("/mycustompath/pprof/threadcreate", NetPprof.Handler("threadcreate"))
		r.Handle("/mycustompath/pprof/mutex", NetPprof.Handler("mutex"))
	})

	return r
}
