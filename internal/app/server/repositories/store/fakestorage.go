package store

type FakeStorage struct{}

// NewFakeStorage конструктур типа FakeStorage
func NewFakeStorage() *FakeStorage {
	return &FakeStorage{}
}

func (f *FakeStorage) UpdateCounter(_ string, _ int64) (int64, error) {
	return 0, nil
}

func (f *FakeStorage) UpdateGauge(_ string, _ float64) (float64, error) {
	return 0, nil
}

func (f *FakeStorage) GetCounter(_ string) (int64, error) {
	return 0, nil
}

func (f *FakeStorage) GetGauge(_ string) (float64, error) {
	return 0, nil
}

func (f *FakeStorage) GetAllMetrics() (map[string]interface{}, error) {
	return nil, nil
}
