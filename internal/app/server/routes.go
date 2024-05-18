package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/handlers"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/middleware"
	"net/http"
)

// Routes задаем маршруты для всего сервиса
func Routes() http.Handler {
	r := chi.NewRouter()
	// вариант подвключения middleware
	r.Use(middleware.TrustedSubnet)
	r.Use(middleware.WithLogging)
	r.Use(middleware.WithSign)
	r.Use(middleware.Gzip)
	// text/plain
	r.Group(func(r chi.Router) {
		r.Use(middleware.TextPlain)
		r.Get("/", handlers.Repo.Default)
		r.Post("/update/{metric}/{name}/{value}", handlers.Repo.PostMetrics)
		r.Get("/value/{metric}/{name}", handlers.Repo.GetMetric)
	})
	// application/json
	r.Group(func(r chi.Router) {
		r.With(middleware.Decrypt).Post("/updates/", handlers.Repo.PostBatchMetrics)
		r.Post("/update/", handlers.Repo.PostMetrics)
		r.Post("/value/", handlers.Repo.GetMetric)
	})
	// ping PostgreSQL
	r.Group(func(r chi.Router) {
		r.Use(middleware.TextPlain)
		r.Get("/ping", handlers.Repo.PingPostgreSQL)
	})

	return r
}
