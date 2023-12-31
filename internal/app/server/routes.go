package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/handlers"
	"net/http"
)

// Routes задаем маршруты для всего сервиса
func Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", handlers.Repo.Default)
	r.Post("/update/{metric}/{name}/{value}", handlers.Repo.PostMetrics)
	r.Get("/value/{metric}/{name}", handlers.Repo.GetMetric)

	return r
}
