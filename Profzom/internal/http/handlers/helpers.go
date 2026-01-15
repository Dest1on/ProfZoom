package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"profzom/internal/common"
)

func errUnauthorized() error {
	return common.NewError(common.CodeUnauthorized, "unauthorized", nil)
}

func idFromPath(r *http.Request, positionFromEnd int) (common.UUID, error) {
	trimmed := strings.Trim(r.URL.Path, "/")
	if trimmed == "" {
		return "", fmt.Errorf("missing path")
	}
	segments := strings.Split(trimmed, "/")
	if len(segments) < positionFromEnd {
		return "", fmt.Errorf("invalid path")
	}
	return common.ParseUUID(segments[len(segments)-positionFromEnd])
}

func decodeJSON(r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return common.NewError(common.CodeValidation, "invalid request body", err)
	}
	return nil
}
