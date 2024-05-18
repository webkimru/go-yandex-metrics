package agent

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"github.com/stretchr/testify/assert"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/config"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/metrics"
	grpc2 "github.com/webkimru/go-yandex-metrics/internal/app/server/grpc"
	"github.com/webkimru/go-yandex-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"sync"
	"testing"
	"time"
)

func TestGetMetrics(t *testing.T) {
	a := config.AppConfig{
		PollInterval: 1,
	}
	app = a

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	m := metrics.Metric{}
	wg.Add(1)
	go GetMetrics(ctx, &wg, &m)

	time.Sleep(3 * time.Second)
	cancel()

	if m.Alloc == 0 {
		t.Error("Expected Alloc value > 0, but got 0")
	}
}

func TestGetExtraMetrics(t *testing.T) {
	a := config.AppConfig{
		PollInterval: 1,
	}
	app = a

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	m := metrics.Metric{}
	wg.Add(1)
	go GetExtraMetrics(ctx, &wg, &m)

	time.Sleep(3 * time.Second)
	cancel()

	if m.TotalMemory == 0 {
		t.Error("Expected TotalMemory value > 0, but got 0")
	}
}

func TestSend(t *testing.T) {
	a := config.AppConfig{
		SecretKey: "123",
	}
	app = a

	publicPEM := `-----BEGIN RSA PUBLIC KEY-----
MIICCgKCAgEAxvtg3AK5F71BcBw1ofMp3osMkO6Hqstmr1hbW2Wrax3qgsqELRyM
Hy2aiQrKiewOcMC+xm5gtvHxn/2MfeMTUbmfB/UJ+H6NMK7QkwCxgb6qevLkS5JD
ntkdrmgVTaXlWUlWr1llChyAbJejUN2GcEYZFp8DFWj/e0k9OTEsZG3XVTnmk+84
OFKT2LD6l9lxmjC/emU/0+4WkTMkaDiun2eY71U+duVWadM47ZCVi0mHjYJkL+Xv
qbdqhz4JVAikITsz5Kyx7LwtRQ/CO2hTcK24eAjKYE+58ITZKrIPaXFfoOdJLSBK
xnbJD+kFL4uwcSzmdnSO99zMMD/TnNUIENCoQKkuK5bkE+icnP2IaBBnKCFCYEeH
AUQT28n4C9rmSIROxY9gejg3CjUlqI3C6xtGFM1APLQjYm6WKgYQKANVSMoWVcw7
Ny/O88es7CQ5nvg1bLaDJSP27mD3hOUVMQ5f7jRCQkY4wXbO2ieRn+3oxeMbkCkX
aeuKxnPddK6j5fbc/6AV5bFWafPLqVwyUSvg4z7Z3KiNtD/GkE7OuUYM8WuKYIH9
himgy2sTZlXx26uIlO2NvppE/sZas/igxIDXcZVBIWlbj46wmQn9I1ZWlTEJvVdO
DmD/4fpSztLiYzwu2F8UV+qx875El3UrI4u0SRMPH2SETiGJ/CtVTYUCAwEAAQ==
-----END RSA PUBLIC KEY-----`
	// encrypt
	block, _ := pem.Decode([]byte(publicPEM))
	publicKeyPEM, err := x509.ParsePKCS1PublicKey(block.Bytes)
	assert.NoError(t, err)
	app.PublicKeyPEM = publicKeyPEM

	var tests = []struct {
		name   string
		url    string
		metric metrics.RequestMetricSlice
	}{
		{
			name: "positive test",
			url:  "http://localhost:8080/update/",
			metric: metrics.RequestMetricSlice{
				{
					ID:    "someMetric",
					MType: "gauge",
					Value: 123.123,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Send(context.Background(), tt.url, tt.metric)
			assert.Error(t, err)
		})
	}
}

func TestAddMetricsToJob(t *testing.T) {
	t.Run("case: ticker", func(t *testing.T) {
		app.PollInterval = 1
		app.ReportInterval = 1
		var wg sync.WaitGroup
		var metric metrics.Metric
		ctx, cancel := context.WithCancel(context.Background())
		jobs := make(chan []metrics.RequestMetric, 1)
		wg.Add(1)
		go AddMetricsToJob(ctx, &wg, &metric, jobs)

		time.Sleep(3 * time.Second)
		cancel()
	})

	t.Run("case: context", func(t *testing.T) {
		app.PollInterval = 1
		app.ReportInterval = 5
		var wg sync.WaitGroup
		var metric metrics.Metric
		ctx, cancel := context.WithCancel(context.Background())
		jobs := make(chan []metrics.RequestMetric, 1)
		wg.Add(1)
		go AddMetricsToJob(ctx, &wg, &metric, jobs)

		time.Sleep(2 * time.Second)
		cancel()
	})
}

func TestSendThroughGRPC(t *testing.T) {
	listen := bufconn.Listen(1024 * 1024)
	defer listen.Close()
	srv := grpc.NewServer()
	defer srv.Stop()
	proto.RegisterMetricsServer(srv, grpc2.Repo)
	go func() {
		err := srv.Serve(listen)
		assert.NoError(t, err)
	}()

	resolver.SetDefaultScheme("passthrough")
	conn, err := grpc.NewClient("bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listen.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)

	client := proto.NewMetricsClient(conn)

	tests := []struct {
		name     string
		ctx      context.Context
		requests []metrics.RequestMetric
		c        proto.MetricsClient
	}{
		{name: "grpc test", ctx: context.Background(), requests: []metrics.RequestMetric{{ID: "someMetric", MType: "counter", Delta: 100}}, c: client},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = SendThroughGRPC(tt.ctx, tt.requests, tt.c)
			assert.NoError(t, err)
		})
	}
}

func TestWorker(t *testing.T) {
	app.PollInterval = 1
	app.ReportInterval = 1
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan []metrics.RequestMetric, 1)
	results := make(chan Result, 1)
	jobs <- []metrics.RequestMetric{
		{
			ID:    "someMetric",
			MType: "counter",
			Delta: 100,
		},
	}

	listen := bufconn.Listen(1024 * 1024)
	defer listen.Close()
	srv := grpc.NewServer()
	defer srv.Stop()
	proto.RegisterMetricsServer(srv, grpc2.Repo)
	go func() {
		err := srv.Serve(listen)
		assert.NoError(t, err)
	}()

	resolver.SetDefaultScheme("passthrough")
	conn, err := grpc.NewClient("bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listen.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)

	client := proto.NewMetricsClient(conn)

	wg.Add(1)
	go Worker(ctx, &wg, jobs, results, client)

	time.Sleep(3 * time.Second)
	cancel()
}
