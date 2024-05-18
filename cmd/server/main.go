package main

import (
	"context"
	"fmt"
	"github.com/webkimru/go-yandex-metrics/internal/app/server"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/file/async"
	mygrpc "github.com/webkimru/go-yandex-metrics/internal/app/server/grpc"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/logger"
	pb "github.com/webkimru/go-yandex-metrics/internal/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// main начало приложения
func main() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	ctx, cancel := context.WithCancel(context.Background())
	// при штатном завершении сервера все накопленные данные должны сохраняться
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// настраиваем/инициализируем приложение
	serverAddress, err := server.Setup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	// HTTP SERVER
	srv := &http.Server{
		Addr:    *serverAddress,
		Handler: server.Routes(),
	}
	// gRPC Server
	var gRPC *grpc.Server
	go func() {
		// определяем порт для сервера
		listen, err := net.Listen("tcp", ":3200")
		if err != nil {
			log.Fatal(err)
		}
		// создаём gRPC-сервер без зарегистрированной службы
		gRPC = grpc.NewServer()
		// регистрируем сервис
		pb.RegisterMetricsServer(gRPC, mygrpc.Repo)
		// получаем запросы gRPC
		fmt.Println("Starting gRPC server on port 3200")
		if err = gRPC.Serve(listen); err != nil {
			log.Fatal(err)
		}

	}()

	// gracefully shutdown
	go func() {
		<-c
		async.SaveData(ctx)
		logger.Log.Infoln("Successful shutdown")
		server.Shutdown(ctx, srv)
		gRPC.Stop()
		cancel()
	}()

	// асинхронная запись метрик
	async.FileWriter(ctx)

	// стартуем сервер
	logger.Log.Infof("Starting metric server on %s", *serverAddress)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal(err)
	}

	<-ctx.Done()
}
