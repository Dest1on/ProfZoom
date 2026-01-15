package middleware

import (
	"log"
	"net/http"
	"time"

	"profzom/internal/common"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		requestID, _ := common.RequestIDFromContext(r.Context())
		log.Printf("request_id=%s method=%s path=%s status=%d duration=%s", requestID, r.Method, r.URL.Path, rw.status, time.Since(start))
	})
}
