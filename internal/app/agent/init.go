package agent

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/logger"
	"github.com/webkimru/go-yandex-metrics/internal/security"
	"log"
	"os"
	"strconv"
)

const (
	HTTP = "HTTP"
	GRPC = "GRPC"
)

func Setup() (string, int, error) {
	// задаем флаги для агента
	serverAddress := flag.String("a", "", "server address")
	reportInterval := flag.Int("r", 0, "report interval (in seconds)")
	pollInterval := flag.Int("p", 0, "poll interval (in seconds)")
	secretKey := flag.String("k", "", "secret key")
	rateLimit := flag.Int("l", 0, "rate limit (a number of workers)")
	cryptoKey := flag.String("crypto-key", "", "path to pem public key file")
	realIP := flag.String("i", "", "real ip")
	serverProtocol := flag.String("s", "", "protocol: HTTP, GRPC")
	configuration := flag.String("c", "", "path to json configuration file")

	// разбор командой строки
	flag.Parse()

	// определение переменных окружения
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		serverAddress = &envRunAddr
	}
	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		ri, err := strconv.Atoi(envReportInterval)
		if err != nil {
			log.Fatal(err)
		}
		reportInterval = &ri
	}
	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		pi, err := strconv.Atoi(envPollInterval)
		if err != nil {
			log.Fatal(err)
		}
		pollInterval = &pi
	}
	if envSecretKey := os.Getenv("KEY"); envSecretKey != "" {
		secretKey = &envSecretKey
	}
	if envRateLimit := os.Getenv("RATE_LIMIT"); envRateLimit != "" {
		pi, err := strconv.Atoi(envRateLimit)
		if err != nil {
			log.Fatal(err)
		}
		rateLimit = &pi
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cryptoKey = &envCryptoKey
	}
	if envRealIP := os.Getenv("REAL_IP"); envRealIP != "" {
		realIP = &envRealIP
	}
	if envServerProtocol := os.Getenv("SERVER_PROTOCOL"); envServerProtocol != "" {
		serverProtocol = &envServerProtocol
	}
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configuration = &envConfig
	}

	// инициализируем логер
	if err := logger.Initialize("info"); err != nil {
		return "", 0, err
	}

	// читаем конфиг из файла
	if *configuration != "" {
		configFile, err := os.ReadFile(*configuration)
		if err != nil {
			return "", 0, fmt.Errorf("failed loading config from file=%s: %w", *configuration, err)
		}
		// определяем для всего сервиса конфигурацию из файла
		if err = json.Unmarshal(configFile, &app); err != nil {
			return "", 0, fmt.Errorf("failed unmarshaling config from file=%s: %w", *configuration, err)
		}

		logger.Log.Infof("configuration loaded successfully from file=%s", *configuration)
	}
	// переопределяем значения конфига значениями из envs / flags:
	if *serverAddress != "" {
		app.ServerAddress = *serverAddress
	}
	if *reportInterval != 0 {
		app.ReportInterval = *reportInterval
	}
	if *pollInterval != 0 {
		app.PollInterval = *pollInterval
	}
	if *secretKey != "" {
		app.SecretKey = *secretKey
	}
	if *rateLimit != 0 {
		app.RateLimit = *rateLimit
	}
	if *cryptoKey != "" {
		app.CryptoKey = *cryptoKey
	}
	if *realIP != "" {
		app.RealIP = *realIP
	}
	if *serverProtocol != "" {
		app.ServerProtocol = *serverProtocol
	}
	// обязательные настройки
	if app.ServerAddress == "" {
		return "", 0, fmt.Errorf("destionation server address is not defined, it must be specified, for example, localhost:8080")
	}
	if app.RateLimit == 0 {
		app.RateLimit = 1 // silent default
		logger.Log.Infof("default rate limit is automatically set = %d", app.RateLimit)
	}
	if app.PollInterval == 0 {
		app.PollInterval = 2 // silent default
		logger.Log.Infof("default poll interval is automatically set = %d", app.PollInterval)
	}
	if app.ReportInterval == 0 {
		app.ReportInterval = 10 // silent default
		logger.Log.Infof("default report interval is automatically set = %d", app.ReportInterval)
	}
	if app.RealIP == "" {
		app.RealIP = "127.0.0.1"
		logger.Log.Infof("default real ip is automatically set = %d", app.RealIP)
	}
	if app.ServerProtocol == "" {
		app.ServerProtocol = HTTP
		logger.Log.Infof("default server protocol is automatically set = %s", app.ServerProtocol)
	}

	logger.Log.Infoln(
		"Starting configuration:",
		"ADDRESS", app.ServerAddress,
		"REPORT_INTERVAL", app.ReportInterval,
		"POLL_INTERVAL", app.PollInterval,
		"KEY", app.SecretKey,
		"CRYPTO_KEY", app.CryptoKey,
		"RATE_LIMIT", app.RateLimit,
		"REAL_IP", app.RealIP,
		"SERVER_PROTOCOL", app.ServerProtocol,
	)

	// инициализация ключей ассиметричного шифрования
	publicKey, err := security.GetPublicKeyPEM(app.CryptoKey)
	if err != nil {
		logger.Log.Errorf("faild GetPublicKeyPEM()=%v", err)
	}
	app.PublicKeyPEM = publicKey

	return app.ServerProtocol, app.RateLimit, nil
}
