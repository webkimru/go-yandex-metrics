package server

import (
	"flag"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/handlers"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/logger"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/repositories/store"
	"os"
)

// Setup будет полезна при инициализации зависимостей сервера перед запуском
func Setup() (*string, error) {

	// указываем имя флага, значение по умолчанию и описание
	serverAddress := flag.String("a", "localhost:8080", "server address")
	// разбор командной строки
	flag.Parse()
	// определяем переменные окружения
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		serverAddress = &envRunAddr
	}

	if err := logger.Initialize("info"); err != nil {
		return nil, err
	}

	// задаем вариант хранения
	memStorage := store.NewMemStorage()
	// инициализируем репозиторий хендлеров с указанным вариантом хранения
	repo := handlers.NewRepo(memStorage)
	// инициализвруем хендлеры для работы с репозиторием
	handlers.NewHandlers(repo)

	return serverAddress, nil
}
