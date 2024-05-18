package handlers

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/repositories/store"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewServer(middleware(routes))
	defer ts.Close()

	tests := []struct {
		name               string
		url                string
		method             string
		expectedStatusCode int
	}{
		{"Positive test: counter", "/update/counter/someMetric/123", http.MethodPost, http.StatusOK},
		{"positive test: gauge", "/update/gauge/someMetric/123", http.MethodPost, http.StatusOK},
		{"positive test: gauge", "/update/gauge/someMetric/123.123", http.MethodPost, http.StatusOK},
		{"positive test: default", "/", http.MethodGet, http.StatusOK},
		{"positive test: counter value", "/value/counter/someMetric", http.MethodGet, http.StatusOK},
		{"positive test: gauge value", "/value/gauge/someMetric", http.MethodGet, http.StatusOK},
		{"negative test: post batch", "/updates/", http.MethodPost, http.StatusBadRequest},
		{"negative test: counter", "/update/counter/someMetric/123.123", http.MethodPost, http.StatusBadRequest},
		{"negative test: counter", "/update/counter/someMetric/none", http.MethodPost, http.StatusBadRequest},
		{"negative test: counter", "/update/counter/someMetric/none", http.MethodPost, http.StatusBadRequest},
		{"negative test: gauge", "/update/gauge/someMetric/none", http.MethodPost, http.StatusBadRequest},
		{"negative test: none", "/update/none/none/none", http.MethodPost, http.StatusBadRequest},
		{"negative test: metric name", "/update/counter//123", http.MethodPost, http.StatusNotFound},
		{"nagative test: http method", "/update/counter/someMetric/123", http.MethodGet, http.StatusMethodNotAllowed},
		{"nagative test: wrong url", "/someurl/", http.MethodPost, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name+":"+tt.url, func(t *testing.T) {
			switch tt.method {
			case http.MethodPost:
				req, err := http.NewRequestWithContext(context.Background(), "POST", ts.URL+tt.url, nil)
				assert.NoError(t, err)
				req.Header.Set("Content-Type", "text/plain")
				client := &http.Client{}
				resp, err := client.Do(req)
				defer resp.Body.Close()
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			case http.MethodGet:
				req, err := http.NewRequestWithContext(context.Background(), "GET", ts.URL+tt.url, nil)
				assert.NoError(t, err)
				client := &http.Client{}
				resp, err := client.Do(req)
				defer resp.Body.Close()
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
			}
		})
	}

	testsContentTypeJSON := []struct {
		name               string
		url                string
		body               string
		expectedStatusCode int
	}{
		{"negative: batch", "/updates/", ``, http.StatusBadRequest},
		{"negative: PostMetric: error reading body", "/update/", `errBody`, http.StatusBadRequest},
		{"positive: batch", "/updates/", `[{"id":"someMetric","type":"counter","delta":10}]`, http.StatusOK},
		{"positive: gauge", "/update/", `{"id":"someMetric","type":"gauge","value":1.23}`, http.StatusOK},
		{"positive: counter", "/update/", `{"id":"someMetric","type":"counter","delta":123}`, http.StatusOK},
	}

	for _, tt := range testsContentTypeJSON {
		t.Run(tt.name+":"+tt.url, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), "POST", ts.URL+tt.url, bytes.NewReader([]byte(tt.body)))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
		})
	}
}

func TestHandlersWithBadStorage(t *testing.T) {
	testStorage := store.NewFakeBadStorage()
	repo := NewRepo(testStorage)
	NewHandlers(repo, app)

	routes := getRoutes()
	ts := httptest.NewServer(middleware(routes))
	defer ts.Close()

	tests := []struct {
		name               string
		url                string
		method             string
		expectedStatusCode int
	}{
		{"bad storage: counter", "/update/counter/someMetric/123", http.MethodPost, http.StatusInternalServerError},
		{"bad storage: gauge", "/update/gauge/someMetric/123.123", http.MethodPost, http.StatusInternalServerError},
		{"bad storage: default", "/", http.MethodGet, http.StatusInternalServerError},
		{"bad storage: counter value", "/value/counter/someMetric", http.MethodGet, http.StatusNotFound},
		{"bad storage: gauge value", "/value/gauge/someMetric", http.MethodGet, http.StatusNotFound},
		{"bad storage: update batch", "/updates/", "Other:json", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name+":"+tt.url, func(t *testing.T) {
			switch tt.method {
			case http.MethodPost:
				req, err := http.NewRequestWithContext(context.Background(), "POST", ts.URL+tt.url, nil)
				assert.NoError(t, err)
				req.Header.Set("Content-Type", "text/plain")
				client := &http.Client{}
				resp, err := client.Do(req)
				defer resp.Body.Close()
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			case http.MethodGet:
				req, err := http.NewRequestWithContext(context.Background(), "GET", ts.URL+tt.url, nil)
				assert.NoError(t, err)
				client := &http.Client{}
				resp, err := client.Do(req)
				defer resp.Body.Close()
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)

			case "Other:json":
				req, err := http.NewRequestWithContext(context.Background(), "POST", ts.URL+tt.url, bytes.NewReader([]byte(`[{"id":"someMetric","type":"counter","delta":10}]`)))
				assert.NoError(t, err)
				req.Header.Set("Content-Type", "application/json")
				client := &http.Client{}
				resp, err := client.Do(req)
				defer resp.Body.Close()
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
			}
		})
	}
}
