package middleware

import (
	"net/http"

	"github.com/Dest1on/ProfZoom-backend/internal/http/metrics"
)

func Metrics(collector *metrics.Collector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}
			collector.IncRequests()
			next.ServeHTTP(w, r)
		})
	}
}
