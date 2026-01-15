package response

import (
	"encoding/json"
	"net/http"

	"profzom/internal/common"
	"profzom/internal/http/metrics"
)

type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

var errorCollector *metrics.Collector

func SetErrorCollector(collector *metrics.Collector) {
	errorCollector = collector
}

func Error(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := common.CodeInternal
	message := "internal error"
	var fields map[string]string
	if appErr, ok := err.(*common.AppError); ok {
		code = appErr.Code
		message = appErr.Message
		fields = appErr.Fields
	}
	switch code {
	case common.CodeValidation:
		status = http.StatusBadRequest
	case common.CodeTelegramNotLinked:
		status = http.StatusConflict
	case common.CodeRateLimited:
		status = http.StatusTooManyRequests
	case common.CodeDeliveryFailed:
		status = http.StatusBadGateway
	case common.CodeUnauthorized:
		status = http.StatusUnauthorized
	case common.CodeForbidden:
		status = http.StatusForbidden
	case common.CodeNotFound:
		status = http.StatusNotFound
	case common.CodeConflict:
		status = http.StatusConflict
	}
	if errorCollector != nil && status >= http.StatusBadRequest {
		errorCollector.IncErrors()
	}
	JSON(w, status, ErrorResponse{Error: string(code), Message: message, Fields: fields})
}
