package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/config"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithSign(t *testing.T) {
	a := config.AppConfig{}
	app = &a
	NewMiddleware(app)

	handler := WithSign(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("valid sign", func(t *testing.T) {
		app.SecretKey = "secret"
		r := httptest.NewRequest("POST", srv.URL, bytes.NewReader([]byte("")))
		r.RequestURI = ""

		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		r.Body = io.NopCloser(bytes.NewReader(b))
		h := hmac.New(sha256.New, []byte(app.SecretKey))
		h.Write(b)
		sign := h.Sum(nil)
		r.Header.Set("HashSHA256", hex.EncodeToString(sign))

		resp, err := http.DefaultClient.Do(r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		defer resp.Body.Close()
	})

	t.Run("invalid sign hex", func(t *testing.T) {
		app.SecretKey = "secret"
		r := httptest.NewRequest("POST", srv.URL, bytes.NewReader([]byte("")))
		r.RequestURI = ""
		r.Header.Set("HashSHA256", "test")
		resp, err := http.DefaultClient.Do(r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		defer resp.Body.Close()
	})

	t.Run("without secret key", func(t *testing.T) {
		app.SecretKey = ""
		r := httptest.NewRequest("POST", srv.URL, bytes.NewReader([]byte("")))
		r.RequestURI = ""
		resp, err := http.DefaultClient.Do(r)
		assert.NoError(t, err)
		defer resp.Body.Close()
	})

	t.Run("empty hash", func(t *testing.T) {
		app.SecretKey = "secret"
		r := httptest.NewRequest("POST", srv.URL, bytes.NewReader([]byte("")))
		r.RequestURI = ""
		r.Header.Set("HashSHA256", "")
		resp, err := http.DefaultClient.Do(r)
		assert.NoError(t, err)
		defer resp.Body.Close()
	})

	t.Run("wrong sign", func(t *testing.T) {
		app.SecretKey = "secret"
		r := httptest.NewRequest("POST", srv.URL, bytes.NewReader([]byte("")))
		r.RequestURI = ""
		r.Header.Set("HashSHA256", "00")

		resp, err := http.DefaultClient.Do(r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		defer resp.Body.Close()
	})
}

func TestWithSignWithBadReader(t *testing.T) {
	a := config.AppConfig{}
	app = &a
	NewMiddleware(app)

	handler := WithSign(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("error reading body", func(t *testing.T) {
		app.SecretKey = "secret"
		r := httptest.NewRequest("POST", srv.URL, errReader(0))
		r.RequestURI = ""

		resp, err := http.DefaultClient.Do(r)
		assert.Error(t, err)
		if resp != nil {
			defer resp.Body.Close()
		}

	})

}

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}
