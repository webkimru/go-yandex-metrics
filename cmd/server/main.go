package main

import (
	"github.com/webkimru/go-yandex-metrics/internal/app/server"
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
	err = http.ListenAndServe(*serverAddress, server.Routes())
	panic(err)
}
