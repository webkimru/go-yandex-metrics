package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"github.com/stretchr/testify/assert"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/config"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecrypt(t *testing.T) {
	a := config.AppConfig{}
	app = &a
	NewMiddleware(app)

	handler := Decrypt(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Run("wrong decrypt", func(t *testing.T) {
		cryptoKey := `-----BEGIN RSA PRIVATE KEY-----
MCQCAQACAwDZhwIDAQABAgIIAQICAOkCAgDvAgIAwQICAJECASc=
-----END RSA PRIVATE KEY-----`
		block, _ := pem.Decode([]byte(cryptoKey))

		privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		app.PrivateKeyPEM = privateKey
		assert.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, bytes.NewReader([]byte("")))
		r.RequestURI = ""

		b, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		data, err := hex.DecodeString(string(b))
		assert.NoError(t, err)
		data, err = rsa.DecryptPKCS1v15(rand.Reader, privateKey, data)
		assert.Error(t, err)

		r.Body = io.NopCloser(bytes.NewReader(data))

		resp, err := http.DefaultClient.Do(r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		defer resp.Body.Close()
	})

	t.Run("without crypto key", func(t *testing.T) {
		app.PrivateKeyPEM = nil
		r := httptest.NewRequest("POST", srv.URL, bytes.NewReader([]byte("")))
		r.RequestURI = ""

		_, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		defer resp.Body.Close()
	})

	t.Run("valid decrypt", func(t *testing.T) {
		privatePEM := `-----BEGIN RSA PRIVATE KEY-----
MIGqAgEAAiEArrViGshBTVoDvVTWb8BvdAvVduuWkRMJ36iF6a3GJh8CAwEAAQIg
ORyyZW7xagfzEQGa2A1gYVT14hp67sh+G3DB8LUKloECEQDVw3DBbl7gfaO3XutL
Dl8bAhEA0Tp4rn6AixrQIDc6JNRRTQIQS2Dds/f9kN/9CT55bkAlHQIRAJH1f3ED
cPsZtm1y+Y3ty9UCEHDFk500N7dGTK3+AKTuTDw=
-----END RSA PRIVATE KEY-----`
		block, _ := pem.Decode([]byte(privatePEM))
		privateKeyPEM, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		assert.NoError(t, err)
		app.PrivateKeyPEM = privateKeyPEM

		publicPEM := `-----BEGIN RSA PUBLIC KEY-----
MCgCIQCutWIayEFNWgO9VNZvwG90C9V265aREwnfqIXprcYmHwIDAQAB
-----END RSA PUBLIC KEY-----`
		// encrypt
		body := []byte("test")
		block, _ = pem.Decode([]byte(publicPEM))
		publicKeyPEM, _ := x509.ParsePKCS1PublicKey(block.Bytes)
		body, _ = rsa.EncryptPKCS1v15(rand.Reader, publicKeyPEM, body)
		body = []byte(hex.EncodeToString(body))

		// request
		r := httptest.NewRequest("POST", srv.URL, bytes.NewReader(body))
		r.RequestURI = ""

		// response
		resp, err := http.DefaultClient.Do(r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		defer resp.Body.Close()
	})

	t.Run("wrong body - no hex encode", func(t *testing.T) {
		publicPEM := `-----BEGIN RSA PUBLIC KEY-----
MCgCIQCutWIayEFNWgO9VNZvwG90C9V265aREwnfqIXprcYmHwIDAQAB
-----END RSA PUBLIC KEY-----`
		// encrypt
		body := []byte("test")
		block, _ := pem.Decode([]byte(publicPEM))
		publicKeyPEM, _ := x509.ParsePKCS1PublicKey(block.Bytes)
		body, _ = rsa.EncryptPKCS1v15(rand.Reader, publicKeyPEM, body)

		// request
		r := httptest.NewRequest("POST", srv.URL, bytes.NewReader(body))
		r.RequestURI = ""

		// response
		resp, err := http.DefaultClient.Do(r)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		defer resp.Body.Close()
	})

	t.Run("error reading body to decrypt", func(t *testing.T) {
		privatePEM := `-----BEGIN RSA PRIVATE KEY-----
MCQCAQACAwDZhwIDAQABAgIIAQICAOkCAgDvAgIAwQICAJECASc=
-----END RSA PRIVATE KEY-----`
		block, _ := pem.Decode([]byte(privatePEM))
		privateKeyPEM, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
		app.PrivateKeyPEM = privateKeyPEM

		r := httptest.NewRequest("POST", srv.URL, errReader(0))
		r.RequestURI = ""

		resp, err := http.DefaultClient.Do(r)
		assert.Error(t, err)
		if resp != nil {
			defer resp.Body.Close()
		}
	})
}
