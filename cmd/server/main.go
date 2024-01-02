package main

import (
	"github.com/webkimru/go-yandex-metrics/internal/app/server"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/middleware"
	"log"
	"net/http"
)

// main начало приложения
func main() {

	// настраиваем/инициализируем приложение
	serverAddress, err := server.Setup()
	if err != nil {
		log.Fatal(err)
	}

	// стартуем сервер
	err = http.ListenAndServe(*serverAddress, middleware.TextPlain(server.Routes()))
	panic(err)
}
