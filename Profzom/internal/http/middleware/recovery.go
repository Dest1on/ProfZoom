package middleware

import (
	"log"
	"net/http"

	"profzom/internal/common"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := common.RequestIDFromContext(r.Context())
				log.Printf("panic request_id=%s err=%v", requestID, err)
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
