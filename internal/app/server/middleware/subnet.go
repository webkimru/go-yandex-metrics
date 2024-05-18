package middleware

import (
	"github.com/webkimru/go-yandex-metrics/internal/app/server/logger"
	"net"
	"net/http"
)

// TrustedSubnet allow requests by client IP from trusted subnet
func TrustedSubnet(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.TrustedSubnet == "" {
			next.ServeHTTP(w, r)
			return
		}
		_, subnet, err := net.ParseCIDR(app.TrustedSubnet)
		if err != nil {
			logger.Log.Errorln("failed ParseCIDR()=", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ip := net.ParseIP(r.Header.Get("X-Real-IP"))

		if !subnet.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
