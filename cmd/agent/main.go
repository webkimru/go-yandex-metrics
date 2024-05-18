package main

import (
	"context"
	"fmt"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/logger"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/metrics"
	pb "github.com/webkimru/go-yandex-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	_ "net/http/pprof" // подключаем пакет pprof
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

var m metrics.Metric

func main() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

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
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	wg.Add(1)
	go func() {
		<-c
		logger.Log.Infoln("Shutdown...")
		wg.Done()
		cancel()
	}()

	go func() {
		err := http.ListenAndServe(":8000", nil)
		if err != nil {
			logger.Log.Errorln("ListenAndServe for pprof doesn't work:", err)
		}
	}()

	// настраиваем/инициализируем приложение
	serverProtocol, rateLimit, err := agent.Setup()
	if err != nil {
		log.Fatal(err)
	}

	// GRPC
	var clientGRPC pb.MetricsClient
	if serverProtocol == agent.GRPC {
		// устанавливаем соединение с сервером GRPC
		conn, err := grpc.NewClient("localhost:3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		// получаем переменную интерфейсного типа MetricsClient,
		// через которую будем отправлять сообщения
		clientGRPC = pb.NewMetricsClient(conn)
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
		go agent.Worker(ctx, &wg, jobs, results, clientGRPC)
	}

	// для контроля ошибок отправки метрик из основного потока
	// можно передать не только ошибку, но и данные, добавить их в новую задачу
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			// ждем отмены контекста из main и выходим
			case <-ctx.Done():
				agent.ShutdownResults(results)
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
