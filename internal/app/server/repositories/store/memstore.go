// Package store обеспечивает хранение данных в памяти.
package store

import (
	"context"
	"fmt"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/config"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/models"
	"sync"
)

type Counter int64
type Gauge float64

// MemStorage описывает структуру хранилища в памяти.
type MemStorage struct {
	Counter map[string]Counter
	Gauge   map[string]Gauge
	mu      sync.Mutex
}

// NewMemStorage конструктур типа MemStorage.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Counter: make(map[string]Counter, 1),
		Gauge:   make(map[string]Gauge, 31),
	}
}

// UpdateCounter обновляет поле Counter.
// Описывает метод в соответствии с контактном интерфейсного типа StoreRepository.
func (ms *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) (int64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Counter[name] += Counter(value)
	return int64(ms.Counter[name]), nil
}

// UpdateGauge обновляет поле Gauge.
func (ms *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) (float64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Gauge[name] = Gauge(value)
	return float64(ms.Gauge[name]), nil
}

// GetCounter возращает значение счетчика Counter.
func (ms *MemStorage) GetCounter(ctx context.Context, metric string) (int64, error) {
	value, ok := ms.Counter[metric]
	if !ok {
		return 0, fmt.Errorf("%s does not exists", metric)
	}
	return int64(value), nil
}

// GetGauge возращает значение счетчика Gauge.
func (ms *MemStorage) GetGauge(ctx context.Context, metric string) (float64, error) {
	value, ok := ms.Gauge[metric]
	if !ok {
		return 0, fmt.Errorf("%s does not exists", metric)
	}
	return float64(value), nil
}

// GetAllMetrics возращает мапку счетчиков Counter и Gauge.
func (ms *MemStorage) GetAllMetrics(ctx context.Context) (map[string]interface{}, error) {
	all := make(map[string]interface{}, 30)
	all["counter"] = ms.Counter
	all["gauge"] = ms.Gauge

	return all, nil
}

// UpdateBatchMetrics обновляет значение метрик Gauge и Counter по входящему батчу.
func (ms *MemStorage) UpdateBatchMetrics(ctx context.Context, metrics []models.Metrics) error {
	for i := range metrics {
		switch metrics[i].MType {
		case "gauge":
			ms.Gauge[metrics[i].ID] = Gauge(*metrics[i].Value)

		case "counter":
			ms.Counter[metrics[i].ID] += Counter(*metrics[i].Delta)
		}
	}

	return nil
}

func (ms *MemStorage) Initialize(ctx context.Context, _ config.AppConfig) error {
	return nil
}
