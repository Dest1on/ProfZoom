package app

import (
	"context"
	"encoding/json"

	"github.com/Dest1on/ProfZoom-backend/internal/common"
)

func analyticsPayload(ctx context.Context, payload map[string]string) []byte {
	if payload == nil {
		payload = map[string]string{}
	}
	if requestID, ok := common.RequestIDFromContext(ctx); ok {
		payload["request_id"] = requestID
	}
	value, _ := json.Marshal(payload)
	return value
}
