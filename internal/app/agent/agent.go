package agent

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/config"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/logger"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/metrics"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

var rt runtime.MemStats
var app config.AppConfig

func GetMetric(m *metrics.Metric) {
	pollDuration := time.Duration(app.PollInterval) * time.Second

	for {
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

		time.Sleep(pollDuration)
	}
}

func SendMetric(metric metrics.Metric) {
	var metricSlice []metrics.RequestMetric

	val := reflect.ValueOf(&metric)
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

	go func() {
		err := Send(fmt.Sprintf("http://%s/updates/", app.ServerAddress), metricSlice)
		if err != nil {
			logger.Log.Error(err)
		}
	}()
}

func Send(url string, request interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request=%v", request)
	}

	// Compress data
	if err := Compress(&data); err != nil {
		return fmt.Errorf("failed Compress()=%v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
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
