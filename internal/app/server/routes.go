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
	r.Use(middleware.WithLogging)
	// описание маршрутов
	r.Get("/", handlers.Repo.Default)
	r.Post("/update/{metric}/{name}/{value}", handlers.Repo.PostMetrics)
	r.Get("/value/{metric}/{name}", handlers.Repo.GetMetric)

	return r
}
