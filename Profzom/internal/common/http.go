package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func ErrUnauthorized() error {
	return NewError(CodeUnauthorized, "unauthorized", nil)
}

func IDFromPath(r *http.Request, positionFromEnd int) (UUID, error) {
	trimmed := strings.Trim(r.URL.Path, "/")
	if trimmed == "" {
		return "", fmt.Errorf("missing path")
	}
	segments := strings.Split(trimmed, "/")
	if len(segments) < positionFromEnd {
		return "", fmt.Errorf("invalid path")
	}
	return ParseUUID(segments[len(segments)-positionFromEnd])
}

func DecodeJSON(r *http.Request, dst interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return NewError(CodeValidation, "invalid request body", err)
	}
	return nil
}
