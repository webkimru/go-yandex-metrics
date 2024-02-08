package main

import (
	"context"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/logger"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/metrics"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var m metrics.Metric

func main() {
	// понадобится для ожидания всех горутин
	var wg sync.WaitGroup

	// задаем максимальное количество задач для воркеров
	const numJobs = 10
	// создаем буферизованный канал для принятия задач в воркер
	jobs := make(chan []metrics.RequestMetric, numJobs)
	// создаем буферизованный канал для результатов отправок
	results := make(chan agent.Result, numJobs)

	// при штатном завершении отменяем контекст для завершения работы всех горутин
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		<-c
		logger.Log.Infoln("Shutdown...")
		wg.Done()
		cancel()
	}()

	// настраиваем/инициализируем приложение
	rateLimit, err := agent.Setup()
	if err != nil {
		log.Fatal(err)
	}

	// получаем базовые метрики
	wg.Add(1)
	go agent.GetMetrics(ctx, &wg, &m)

	// получаем дополнительные метрики
	wg.Add(1)
	go agent.GetExtraMetrics(ctx, &wg, &m)

	// добавляем метрики в новую задачу
	wg.Add(1)
	go agent.AddMetricsToJob(ctx, &wg, &m, jobs)

	// запускаем rateLimit воркеров для наших задач
	for w := 1; w <= rateLimit; w++ {
		wg.Add(1)
		go agent.Worker(ctx, &wg, jobs, results)
	}

	// для контроля ошибок отправки метрик из основного потока
	// можно передать не только ошибку, но и данные, добавить их в новую задачу
	wg.Add(1)
	go func() {
		for {
			select {
			// ждем отмены контекста из main и выходим
			case <-ctx.Done():
				agent.ShutdownResults(results)
				wg.Done()
				return
			case res := <-results:
				if res.Err != nil {
					logger.Log.Errorln(res.Err)
				}
			}
		}
	}()

	wg.Wait()
	logger.Log.Infoln("Successful shutdown")
}
