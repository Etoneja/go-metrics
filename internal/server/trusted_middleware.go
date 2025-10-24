package server

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

func (bmw *BaseMiddleware) TrustedIPMiddleware(allowedSubnet string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if allowedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}

			realIP := r.Header.Get("X-Real-IP")
			if realIP == "" {
				bmw.logger.Warn("X-Real-IP header is missing")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(realIP)
			if ip == nil {
				bmw.logger.Warn("Invalid IP address in X-Real-IP header", zap.String("ip", realIP))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			_, subnet, err := net.ParseCIDR(allowedSubnet)
			if err != nil {
				bmw.logger.Error("Invalid subnet configuration",
					zap.String("subnet", allowedSubnet),
					zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !subnet.Contains(ip) {
				bmw.logger.Warn("IP address not in allowed subnet",
					zap.String("ip", realIP),
					zap.String("subnet", allowedSubnet))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
