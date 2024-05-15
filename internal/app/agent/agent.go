package agent

import (
	"bytes"
	"context"
	"crypto/hmac"
	randcrypto "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/mailru/easyjson"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/config"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/logger"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/metrics"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"sync"
	"time"
)

var app config.AppConfig

func GetMetrics(ctx context.Context, wg *sync.WaitGroup, m *metrics.Metric) {
	defer wg.Done()

	var rt runtime.MemStats
	// будем собирать метрики каждые PollInterval секунд
	ticker := time.NewTicker(time.Duration(app.PollInterval) * time.Second)

	for {
		select {
		// ждем отмены контекста из main и выходим из функции
		case <-ctx.Done():
			return
		// ждем таймер
		case <-ticker.C:
			runtime.ReadMemStats(&rt)
			m.Alloc = metrics.Gauge(rt.Alloc)
			m.BuckHashSys = metrics.Gauge(rt.BuckHashSys)
			m.Frees = metrics.Gauge(rt.Frees)
			m.GCCPUFraction = metrics.Gauge(rt.GCCPUFraction)
			m.GCSys = metrics.Gauge(rt.GCSys)
			m.HeapAlloc = metrics.Gauge(rt.HeapAlloc)
			m.HeapIdle = metrics.Gauge(rt.HeapIdle)
			m.HeapInuse = metrics.Gauge(rt.HeapInuse)
			m.HeapObjects = metrics.Gauge(rt.HeapObjects)
			m.HeapReleased = metrics.Gauge(rt.HeapReleased)
			m.HeapSys = metrics.Gauge(rt.HeapSys)
			m.LastGC = metrics.Gauge(rt.LastGC)
			m.Lookups = metrics.Gauge(rt.Lookups)
			m.MCacheInuse = metrics.Gauge(rt.MCacheInuse)
			m.MCacheSys = metrics.Gauge(rt.MCacheSys)
			m.MSpanInuse = metrics.Gauge(rt.MSpanInuse)
			m.MSpanSys = metrics.Gauge(rt.MSpanSys)
			m.Mallocs = metrics.Gauge(rt.Mallocs)
			m.NextGC = metrics.Gauge(rt.NextGC)
			m.NumForcedGC = metrics.Gauge(rt.NumForcedGC)
			m.NumGC = metrics.Gauge(rt.NumGC)
			m.OtherSys = metrics.Gauge(rt.OtherSys)
			m.PauseTotalNs = metrics.Gauge(rt.PauseTotalNs)
			m.StackInuse = metrics.Gauge(rt.StackInuse)
			m.StackSys = metrics.Gauge(rt.StackSys)
			m.Sys = metrics.Gauge(rt.Sys)
			m.TotalAlloc = metrics.Gauge(rt.TotalAlloc)

			m.RandomValue = metrics.Gauge(rand.Float64())
			m.PollCount++
		}
	}
}

func GetExtraMetrics(ctx context.Context, wg *sync.WaitGroup, m *metrics.Metric) {
	defer wg.Done()

	v, err := mem.VirtualMemory()
	if err != nil {
		logger.Log.Errorln(err)
	}
	c, err := cpu.Percent(0, true)
	if err != nil {
		logger.Log.Errorln(err)
	}
	// будем собирать метрики каждые PollInterval секунд
	ticker := time.NewTicker(time.Duration(app.PollInterval) * time.Second)

	for {
		select {
		// ждем отмены контекста из main и выходим из функции
		case <-ctx.Done():
			return
		// ждем таймер
		case <-ticker.C:
			m.TotalMemory = metrics.Gauge(v.Total)
			m.FreeMemory = metrics.Gauge(v.Free)
			m.CPUutilization1 = metrics.Gauge(c[0])
		}
	}
}

// Result структура, в которую добавили ошибку
type Result struct {
	Err error
}

// Worker принимает два канала:
// jobs - канал задач для отправки метрик
// results - канал результатов работы
func Worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan []metrics.RequestMetric, results chan<- Result) {
	defer wg.Done()

	for {
		select {
		// ждем отмены контекста из main и выходим
		case <-ctx.Done():
			return
		// или читаем задачи
		case job := <-jobs:
			err := Send(ctx, fmt.Sprintf("http://%s/updates/", app.ServerAddress), job)
			if err != nil {
				result := Result{
					Err: err,
				}
				results <- result
			}
		}
	}
}

func AddMetricsToJob(ctx context.Context, wg *sync.WaitGroup, metric *metrics.Metric, jobs chan []metrics.RequestMetric) {
	defer wg.Done()

	// будем добавлять задачи с метриками каждые app.ReportInterval секунд = отравка с данным интервалом
	ticker := time.NewTicker(time.Duration(app.ReportInterval) * time.Second)

	for {
		select {
		// ждем отмены контекста из main и выходим
		case <-ctx.Done():
			// где пишем, там и закрываем канал
			close(jobs)
			ShutdownJobs(ctx, jobs)
			return
		// ждем таймер
		case <-ticker.C:
			var metricSlice []metrics.RequestMetric

			val := reflect.ValueOf(metric)
			val = val.Elem()
			for fieldIndex := 0; fieldIndex < val.NumField(); fieldIndex++ {
				field := val.Field(fieldIndex)
				f := val.FieldByName(val.Type().Field(fieldIndex).Name)

				switch f.Kind() {
				case reflect.Int64:
					metricSlice = append(metricSlice, metrics.RequestMetric{
						ID:    val.Type().Field(fieldIndex).Name,
						MType: "counter",
						Delta: field.Int(),
					})

				case reflect.Float64:
					metricSlice = append(metricSlice, metrics.RequestMetric{
						ID:    val.Type().Field(fieldIndex).Name,
						MType: "gauge",
						Value: field.Float(),
					})
				}
			}
			// пишем новую задачу в виде слайса метрик
			jobs <- metricSlice
		}
	}
}

func Send(ctx context.Context, url string, request metrics.RequestMetricSlice) error {
	data, err := easyjson.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request=%v, err=%w", request, err)
	}

	// Encrypt request data
	if app.PublicKeyPEM != nil {
		data, err = rsa.EncryptPKCS1v15(randcrypto.Reader, app.PublicKeyPEM, data)
		if err != nil {
			return fmt.Errorf("failed EncryptPKCS1v15()=%w", err)
		}
		data = []byte(hex.EncodeToString(data))
	}

	// Compress data
	if err = Compress(&data); err != nil {
		return fmt.Errorf("failed Compress()=%w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("X-Real-IP", app.RealIP)
	// Encrypt data
	if app.SecretKey != "" {
		// подписываем алгоритмом HMAC, используя SHA-256
		h := hmac.New(sha256.New, []byte(app.SecretKey))
		h.Write(data)
		sign := h.Sum(nil)
		req.Header.Set("HashSHA256", hex.EncodeToString(sign))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status code 200, but got %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	return nil
}

func ShutdownJobs(ctx context.Context, jobs chan []metrics.RequestMetric) {
	logger.Log.Infof("Sending %d metric jobs...", len(jobs))

	for job := range jobs {
		err := Send(ctx, fmt.Sprintf("http://%s/updates/", app.ServerAddress), job)
		if err != nil {
			logger.Log.Errorln(err)
		}
	}
}

func ShutdownResults(results chan Result) {
	logger.Log.Infof("Writting %d logs of the results...", len(results))

	close(results)

	for res := range results {
		if res.Err != nil {
			logger.Log.Errorln(res.Err)
		}
	}
}
