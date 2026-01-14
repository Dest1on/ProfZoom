package handlers

import (
	"net/http"

	"github.com/Dest1on/ProfZoom-backend/internal/http/metrics"
	"github.com/Dest1on/ProfZoom-backend/internal/http/response"
)

type MetricsHandler struct {
	collector *metrics.Collector
}

func NewMetricsHandler(collector *metrics.Collector) *MetricsHandler {
	return &MetricsHandler{collector: collector}
}

func (h *MetricsHandler) Get(w http.ResponseWriter, _ *http.Request) {
	requests, errors := h.collector.Snapshot()
	response.JSON(w, http.StatusOK, map[string]uint64{
		"requests": requests,
		"errors":   errors,
	})
}
