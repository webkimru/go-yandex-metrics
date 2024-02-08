package server

import (
	"context"
	"flag"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/config"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/file"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/file/async"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/handlers"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/logger"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/middleware"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/repositories/store"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/repositories/store/pg"
	"log"
	"os"
	"strconv"
)

var app config.AppConfig

// Setup будет полезна при инициализации зависимостей сервера перед запуском
func Setup(ctx context.Context) (*string, error) {

	// указываем имя флага, значение по умолчанию и описание
	serverAddress := flag.String("a", "localhost:8080", "server address")
	// интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск
	// (по умолчанию 300 секунд, значение 0 делает запись синхронной)
	storeInterval := flag.Int("i", 300, "store interval")
	storeFilePath := flag.String("f", "/tmp/metrics-db.json", "file storage path")
	storeRestore := flag.Bool("r", true, "restore saved data")
	databaseDSN := flag.String("d", "", "database dsn")
	secretKey := flag.String("k", "", "secret key")
	// разбор командной строки
	flag.Parse()
	// определяем переменные окружения
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		serverAddress = &envRunAddr
	}
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		si, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			log.Fatal(err)
		}
		storeInterval = &si
	}
	if envStoreFilePath := os.Getenv("FILE_STORAGE_PATH"); envStoreFilePath != "" {
		storeFilePath = &envStoreFilePath
	}
	if envStoreRestore := os.Getenv("RESTORE"); envStoreRestore != "" {
		sr, err := strconv.ParseBool(envStoreRestore)
		if err != nil {
			log.Fatal(err)
		}
		storeRestore = &sr
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		databaseDSN = &envDatabaseDSN
	}
	if envSecretKey := os.Getenv("KEY"); envSecretKey != "" {
		secretKey = &envSecretKey
	}

	// инициализируем логер
	if err := logger.Initialize("info"); err != nil {
		return nil, err
	}

	// конфигурация приложения
	a := config.AppConfig{
		ServerAddress: *serverAddress,
		SecretKey:     *secretKey,
		DatabaseDSN:   *databaseDSN,
		FileStore: config.RecorderConfig{
			Interval: *storeInterval,
			Restore:  *storeRestore,
			FilePath: *storeFilePath,
		},
	}
	app = a

	logger.Log.Infoln(
		"Starting configuration:",
		"ADDRESS", app.ServerAddress,
		"STORE_INTERVAL", app.FileStore.Interval,
		"FILE_STORAGE_PATH", app.FileStore.FilePath,
		"RESTORE", app.FileStore.Restore,
		"DATABASE_DSN", app.DatabaseDSN,
		"KEY", app.SecretKey,
	)

	// инициализируем хранение метрик в файле
	if err := file.Initialize(&app); err != nil {
		return nil, err
	}
	if err := async.WriterInitialize(&app); err != nil {
		return nil, err
	}

	// задаем варианты хранения
	// 1 - DB
	// 2 - File
	// 3 - Memory
	var storePriority config.Store
	var repo *handlers.Repository
	switch {
	case app.DatabaseDSN != "": // DB
		storePriority = config.Database
		conn, err := pg.ConnectToDB(app.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		if err := pg.Bootstrap(ctx, conn); err != nil {
			log.Fatal(err)
		}
		db := pg.NewStore(conn)
		pg.DB = db
		storage := db
		repo = handlers.NewRepo(storage)

	default: // in memory
		storePriority = config.Memory
		storage := store.NewMemStorage()
		// загружать ранее сохранённые значения из указанного файла при старте сервера
		if app.FileStore.Restore {
			res, err := file.Reader()
			if err != nil {
				return nil, err
			}
			// если не пустой файл
			if res != nil {
				storage.Counter = res.Counter
				storage.Gauge = res.Gauge
			}
		}
		// инициализируем репозиторий хендлеров с указанным вариантом хранения
		repo = handlers.NewRepo(storage)
	}

	// запоминаем вариант хранения
	app.StorePriority = storePriority

	// инициализируем
	middleware.NewMiddleware(&app)
	// инициализвруем хендлеры для работы с репозиторием
	handlers.NewHandlers(repo, &app)

	return serverAddress, nil
}

func ShutdownDB() {
	if app.StorePriority == config.Database {
		pg.DB.Conn.Close()
	}
}
