package main

import (
	"github.com/webkimru/go-yandex-metrics/internal/app/agent"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/metrics"
	"log"
	"time"
)

var m metrics.Metric

func main() {
	// настраиваем/инициализируем приложение
	reportInterval, err := agent.Setup()
	if err != nil {
		log.Fatal(err)
	}

	// получаем метрики
	go agent.GetMetric(&m)

	// отдаем метрики
	reportDuration := time.Duration(*reportInterval) * time.Second
	for {
		time.Sleep(reportDuration)
		agent.SendMetric(m)
	}
}
