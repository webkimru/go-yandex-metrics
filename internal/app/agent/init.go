package agent

import (
	"flag"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/config"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/logger"
	"log"
	"os"
	"strconv"
)

func Setup() (int, error) {
	// задаем флаги для агента
	serverAddress := flag.String("a", "localhost:8080", "server address")
	reportInterval := flag.Int("r", 10, "report interval (in seconds)")
	pollInterval := flag.Int("p", 2, "poll interval (in seconds)")
	secretKey := flag.String("k", "", "secret key")
	rateLimit := flag.Int("l", 1, "rate limit (a number of workers)")

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

	// конфигурация приложения
	a := config.AppConfig{
		ServerAddress:  *serverAddress,
		ReportInterval: *reportInterval,
		PollInterval:   *pollInterval,
		SecretKey:      *secretKey,
		RateLimit:      *rateLimit,
	}
	app = a

	// инициализируем логер
	if err := logger.Initialize("info"); err != nil {
		return 0, err
	}

	logger.Log.Infoln(
		"Starting configuration:",
		"ADDRESS", app.ServerAddress,
		"REPORT_INTERVAL", app.ReportInterval,
		"POLL_INTERVAL", app.PollInterval,
		"KEY", app.SecretKey,
	)

	return app.RateLimit, nil
}
