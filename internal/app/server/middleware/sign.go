package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/logger"
	"io"
	"net/http"
)

// WithSign проверяем полученный и вычесленный хеш
func WithSign(next http.Handler) http.Handler {
	// получаем Handler приведением типа http.HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.SecretKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		// При наличии ключа на этапе формирования ответа сервер должен вычислять хеш
		// и передавать его в HTTP-заголовке ответа с именем `HashSHA256`
		b, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Log.Errorln("failed to read body, ReadAll()=", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(b))
		h := hmac.New(sha256.New, []byte(app.SecretKey))
		h.Write(b)
		sign := h.Sum(nil)
		w.Header().Set("HashSHA256", hex.EncodeToString(sign))

		//  При наличии ключа во время обработки запроса сервер должен проверять
		//  соответствие полученного и вычисленного хеша.
		receivedHash := r.Header.Get("HashSHA256")
		if receivedHash == "" {
			next.ServeHTTP(w, r)
			return
		}
		hash, err := hex.DecodeString(receivedHash)
		if err != nil {
			logger.Log.Errorln("failed to decode, DecodeString()=", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// При несовпадении сервер должен отбрасывать полученные данные и возвращать `http.StatusBadRequest`.
		if !hmac.Equal(sign, hash) {
			logger.Log.Infoln("Wrong sign")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
