package agent

import (
	"github.com/webkimru/go-yandex-metrics/internal/app/agent/metrics"
	"net/http"
	"testing"
	"time"
)

func TestGetMetric(t *testing.T) {
	m := metrics.Metric{}
	go GetMetric(&m, 1)

	time.Sleep(2 * time.Second)

	if m.Alloc == 0 {
		t.Error("sdfsdf")
	}
}

func TestSendMetric(t *testing.T) {
	tests := []struct {
		name               string
		metric             metrics.Metric
		expectedStatusCode int
	}{
		{
			name: "positive test",
			metric: metrics.Metric{
				RandomValue: 123.123,
				PollCount:   1,
			},
			expectedStatusCode: http.StatusOK,
		},
		//{"positive test: gauge", "http://localhost:8080/update/gauge/someMetric/123", http.StatusOK},
		//{"positive test: gauge", "http://localhost:8080/update/gauge/someMetric/123.123", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendMetric(tt.metric, "localhost:8080")
		})
	}
}
