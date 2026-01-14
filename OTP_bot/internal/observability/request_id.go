package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type requestIDKey struct{}

// NewRequestID генерирует случайный ID запроса.
func NewRequestID() string {
	buffer := make([]byte, 12)
	_, err := rand.Read(buffer)
	if err != nil {
		return "unknown"
	}
	return hex.EncodeToString(buffer)
}

// WithRequestID сохраняет ID запроса в контексте.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestIDFromContext возвращает ID запроса из контекста.
func RequestIDFromContext(ctx context.Context) string {
	value := ctx.Value(requestIDKey{})
	if value == nil {
		return ""
	}
	if id, ok := value.(string); ok {
		return id
	}
	return ""
}
