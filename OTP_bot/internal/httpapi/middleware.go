package httpapi

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/profzoom/otp_bot/internal/observability"
)

const (
	internalAuthHeader    = "Authorization"
	internalAuthAltHeader = "X-Internal-Key"
)

func requireInternalAuth(w http.ResponseWriter, r *http.Request, internalKey string, logger *slog.Logger) bool {
	key := strings.TrimSpace(internalKey)
	if key == "" {
		if logger == nil {
			logger = slog.Default()
		}
		logger.Error("internal auth key missing", slog.String("request_id", observability.RequestIDFromContext(r.Context())))
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}

	altValue := strings.TrimSpace(r.Header.Get(internalAuthAltHeader))
	value := strings.TrimSpace(r.Header.Get(internalAuthHeader))
	if altValue == key || value == "Bearer "+key {
		return true
	}

	if logger == nil {
		logger = slog.Default()
	}
	logger.Warn("invalid internal auth", slog.String("request_id", observability.RequestIDFromContext(r.Context())), slog.String("path", r.URL.Path))
	writeError(w, http.StatusUnauthorized, "unauthorized")
	return false
}

func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}
