package store

import (
	"context"
	"fmt"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/config"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/models"
)

type FakeBadStorage struct{}

// NewFakeBadStorage конструктур типа FakeBadStorage
func NewFakeBadStorage() *FakeBadStorage {
	return &FakeBadStorage{}
}

func (f *FakeBadStorage) UpdateCounter(_ context.Context, _ string, _ int64) (int64, error) {
	return 0, fmt.Errorf("err")
}

func (f *FakeBadStorage) UpdateGauge(_ context.Context, _ string, _ float64) (float64, error) {
	return 0, fmt.Errorf("err")
}

func (f *FakeBadStorage) GetCounter(_ context.Context, _ string) (int64, error) {
	return 0, fmt.Errorf("err")
}

func (f *FakeBadStorage) GetGauge(_ context.Context, _ string) (float64, error) {
	return 0, fmt.Errorf("err")
}

func (f *FakeBadStorage) GetAllMetrics(_ context.Context) (map[string]interface{}, error) {
	return nil, fmt.Errorf("err")
}

func (f *FakeBadStorage) UpdateBatchMetrics(_ context.Context, _ []models.Metrics) error {
	return fmt.Errorf("err")
}

func (f *FakeBadStorage) Initialize(_ context.Context, _ config.AppConfig) error {
	return fmt.Errorf("err")
}
