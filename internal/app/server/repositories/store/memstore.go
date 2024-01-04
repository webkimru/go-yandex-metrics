package store

import (
	"fmt"
	"sync"
)

type Counter int64
type Gauge float64

// MemStorage описываем структуру хранилища в памяти
type MemStorage struct {
	Counter map[string]Counter
	Gauge   map[string]Gauge
	mu      sync.Mutex
}

// NewMemStorage конструктур типа MemStorage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Counter: make(map[string]Counter, 1),
		Gauge:   make(map[string]Gauge, 28),
	}
}

// UpdateCounter обновляем поле Counter
// описываем метод в соответствии с контактном интерфейсного типа StoreRepository
func (ms *MemStorage) UpdateCounter(name string, value int64) (int64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Counter[name] += Counter(value)
	return int64(ms.Counter[name]), nil
}

// UpdateGauge обновляем поле UpdateGauge
func (ms *MemStorage) UpdateGauge(name string, value float64) (float64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.Gauge[name] = Gauge(value)
	return float64(ms.Gauge[name]), nil
}

func (ms *MemStorage) GetCounter(metric string) (int64, error) {
	value, ok := ms.Counter[metric]
	if !ok {
		return 0, fmt.Errorf("%s does not exists", metric)
	}
	return int64(value), nil
}

func (ms *MemStorage) GetGauge(metric string) (float64, error) {
	value, ok := ms.Gauge[metric]
	if !ok {
		return 0, fmt.Errorf("%s does not exists", metric)
	}
	return float64(value), nil
}

func (ms *MemStorage) GetAllMetrics() (map[string]interface{}, error) {
	all := make(map[string]interface{}, 30)
	all["counter"] = ms.Counter
	all["gauge"] = ms.Gauge

	return all, nil
}
