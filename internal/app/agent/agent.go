package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/metrics"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
	"time"
)

var rt runtime.MemStats

func GetMetric(m *metrics.Metric, pollInterval int) {
	pollDuration := time.Duration(pollInterval) * time.Second

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

		//log.Println(m.PollCount)
		time.Sleep(pollDuration)
	}
}

func SendMetric(metric metrics.Metric, path string) {

	val := reflect.ValueOf(&metric)
	val = val.Elem()
	for fieldIndex := 0; fieldIndex < val.NumField(); fieldIndex++ {
		field := val.Field(fieldIndex)
		f := val.FieldByName(val.Type().Field(fieldIndex).Name)

		switch f.Kind() {
		case reflect.Int64:
			go func(fieldIndex int) {
				err := SendCounterJSON(fmt.Sprintf("http://%s/update/", path), val.Type().Field(fieldIndex).Name, field.Int())
				if err != nil {
					log.Println(err)
				}

			}(fieldIndex)

		case reflect.Float64:
			go func(fieldIndex int) {
				err := SendGaugeJSON(fmt.Sprintf("http://%s/update/", path), val.Type().Field(fieldIndex).Name, field.Float())
				if err != nil {
					log.Println(err)
				}
			}(fieldIndex)
		}
	}
}

func SendCounterJSON(url string, metric string, value int64) error {
	request := struct {
		ID    string `json:"id"`
		MType string `json:"type"`
		Delta int64  `json:"delta"`
	}{
		ID:    metric,
		MType: "counter",
		Delta: value,
	}

	if err := Send(url, request); err != nil {
		return err
	}

	return nil
}

func SendGaugeJSON(url string, metric string, value float64) error {
	request := struct {
		ID    string  `json:"id"`
		MType string  `json:"type"`
		Value float64 `json:"value"`
	}{
		ID:    metric,
		MType: "gauge",
		Value: value,
	}

	if err := Send(url, request); err != nil {
		return err
	}

	return nil
}

func Send(url string, request interface{}) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request=%v", request)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

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
