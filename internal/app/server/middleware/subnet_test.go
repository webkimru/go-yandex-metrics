package middleware

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustedSubnet(t *testing.T) {
	a := config.AppConfig{}
	app = &a
	NewMiddleware(app)

	handler := TrustedSubnet(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	tests := []struct {
		name               string
		trustedSubnet      string
		realIP             string
		expectedStatusCode int
	}{
		{"positive: valid ip", "127.0.0.0/8", "127.0.0.1", http.StatusOK},
		{"negative: invalid ip", "192.168.1.0/32", "192.168.0.1", http.StatusForbidden},
		{"positive: without subnet", "", "", http.StatusOK},
		{"negative: invalid subnet", "none", "", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.TrustedSubnet = tt.trustedSubnet
			r := httptest.NewRequest("POST", srv.URL, bytes.NewReader([]byte("")))
			if tt.realIP != "" {
				r.Header.Set("X-Real-IP", tt.realIP)
			}
			r.RequestURI = ""

			resp, err := http.DefaultClient.Do(r)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
			defer resp.Body.Close()
		})
	}
}
