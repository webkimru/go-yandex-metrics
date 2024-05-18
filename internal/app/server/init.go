package server

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/config"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/file"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/file/async"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/grpc"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/handlers"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/logger"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/middleware"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/repositories"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/repositories/store"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/repositories/store/pg"
	"github.com/webkimru/go-yandex-metrics/internal/security"
	"net/http"
	"os"
	"strconv"
)

var app config.AppConfig

const (
	HTTP = "HTTP"
	GRPC = "GRPC"
)

// Setup будет полезна при инициализации зависимостей сервера перед запуском
func Setup(ctx context.Context) (*string, error) {
	// указываем имя флага, значение по умолчанию и описание
	serverAddress := flag.String("a", "", "server address")
	// интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск
	// (по умолчанию 300 секунд, значение 0 делает запись синхронной)
	storeInterval := flag.Int("i", 0, "store interval")
	storeFilePath := flag.String("f", "", "file storage path")
	storeRestore := flag.Bool("r", false, "restore saved data")
	databaseDSN := flag.String("d", "", "database dsn")
	secretKey := flag.String("k", "", "secret key")
	cryptoKey := flag.String("crypto-key", "", "path to pem private key file")
	trustedSubnet := flag.String("t", "", "trusted subnet")
	serverProtocol := flag.String("s", "", "protocol: HTTP, GRPC")
	configuration := flag.String("c", "", "path to json configuration file")
	// разбор командной строки
	flag.Parse()
	// определяем переменные окружения
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		serverAddress = &envRunAddr
	}
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		si, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			return nil, err
		}
		storeInterval = &si
	}
	if envStoreFilePath := os.Getenv("FILE_STORAGE_PATH"); envStoreFilePath != "" {
		storeFilePath = &envStoreFilePath
	}
	if envStoreRestore := os.Getenv("RESTORE"); envStoreRestore != "" {
		sr, err := strconv.ParseBool(envStoreRestore)
		if err != nil {
			return nil, err
		}
		storeRestore = &sr
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		databaseDSN = &envDatabaseDSN
	}
	if envSecretKey := os.Getenv("KEY"); envSecretKey != "" {
		secretKey = &envSecretKey
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cryptoKey = &envCryptoKey
	}
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		trustedSubnet = &envTrustedSubnet
	}
	if envServerProtocol := os.Getenv("SERVER_PROTOCOL"); envServerProtocol != "" {
		serverProtocol = &envServerProtocol
	}
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configuration = &envConfig
	}

	// инициализируем логер
	if err := logger.Initialize("info"); err != nil {
		return nil, err
	}

	// читаем конфиг из файла
	if *configuration != "" {
		configFile, err := os.ReadFile(*configuration)
		if err != nil {
			return nil, fmt.Errorf("failed loading config from file=%s: %w", *configuration, err)
		}
		if err = json.Unmarshal(configFile, &app); err != nil {
			return nil, fmt.Errorf("failed unmarshaling config from file=%s: %w", *configuration, err)
		}

		logger.Log.Infof("configuration loaded successfully from file=%s", *configuration)
	}
	// переопределяем значения конфига значениями из envs / flags:
	if *serverAddress != "" {
		app.ServerAddress = *serverAddress
	}
	if *storeInterval != 0 {
		app.FileStore.Interval = *storeInterval
	}
	if *storeFilePath != "" {
		app.FileStore.FilePath = *storeFilePath
	}
	if *storeRestore {
		app.FileStore.Restore = *storeRestore
	}
	if *databaseDSN != "" {
		app.DatabaseDSN = *databaseDSN
	}
	if *secretKey != "" {
		app.SecretKey = *secretKey
	}
	if *cryptoKey != "" {
		app.CryptoKey = *cryptoKey
	}
	if *trustedSubnet != "" {
		app.TrustedSubnet = *trustedSubnet
	}
	if *serverProtocol != "" {
		app.ServerProtocol = *serverProtocol
	}
	// обязательные настройки
	if app.ServerAddress == "" {
		return nil, fmt.Errorf("server address is not defined, it must be specified, for example, localhost:8080")
	}
	if app.FileStore.FilePath == "" {
		app.FileStore.FilePath = "/tmp/metrics-db.json" // silent default
		logger.Log.Infof("storage file is automatically set = %s", app.FileStore.FilePath)
	}
	if app.ServerProtocol == "" {
		app.ServerProtocol = HTTP
		logger.Log.Infof("default server protocol is automatically set = %s", app.ServerProtocol)
	}

	logger.Log.Infoln(
		"Starting configuration:",
		"ADDRESS", app.ServerAddress,
		"STORE_INTERVAL", app.FileStore.Interval,
		"FILE_STORAGE_PATH", app.FileStore.FilePath,
		"RESTORE", app.FileStore.Restore,
		"DATABASE_DSN", app.DatabaseDSN,
		"KEY", app.SecretKey,
		"CRYPTO_KEY", app.CryptoKey,
		"TRUSTED_SUBNET", app.TrustedSubnet,
	)

	// инициализация ключей шифрования
	privateKey, err := security.GetPrivateKeyPEM(app.CryptoKey)
	if err != nil {
		logger.Log.Errorf("faild GetPrivateKeyPEM()=%v", err)
		logger.Log.Warn("Service starting without encryption")
	}
	app.PrivateKeyPEM = privateKey

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
	var db repositories.StoreRepository
	switch {
	case app.DatabaseDSN != "": // DB
		storePriority = config.Database
		db = &pg.Store{}

	default: // in memory
		storePriority = config.Memory
		db = &store.MemStorage{}
		// загружать ранее сохранённые значения из указанного файла при старте сервера
		if app.FileStore.Restore {
			res, err := file.Reader()
			if err != nil {
				return nil, err
			}
			// если не пустой файл
			if res != nil {
				db = &store.MemStorage{Counter: res.Counter, Gauge: res.Gauge}
			}
		}
	}

	if err = db.Initialize(ctx, app); err != nil {
		return nil, err
	}

	// инициализируем репозиторий хендлеров с указанным вариантом хранения
	repo := handlers.NewRepo(db)
	// запоминаем вариант хранения
	app.StorePriority = storePriority
	// инициализируем
	middleware.NewMiddleware(&app)
	// инициализвруем хендлеры для работы с репозиторием
	handlers.NewHandlers(repo, &app)

	repoGRPC := grpc.NewRepo(db)
	grpc.NewMetricHandlers(repoGRPC)

	return &app.ServerAddress, nil
}

func Shutdown(ctx context.Context, srv *http.Server) {
	if app.StorePriority == config.Database {
		err := pg.DB.Conn.Close()
		if err != nil {
			logger.Log.Errorf("Faild pg.DB.Conn.Close(): %v", err)
		}
	}

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatalf("Server shutdown failed: %v", err)
	}
}
