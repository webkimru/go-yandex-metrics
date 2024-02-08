package middleware

import "github.com/webkimru/go-yandex-metrics/internal/app/server/config"

var app *config.AppConfig

func NewMiddleware(a *config.AppConfig) {
	app = a
}
