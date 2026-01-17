package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"otp_bot/internal/linking"
	"otp_bot/internal/ratelimit"
)

func TestLinkTokenRequiresAuth(t *testing.T) {
	tokenStore := linking.NewMemoryLinkTokenStore()
	linkStore := linking.NewMemoryTelegramLinkStore()
	registrar := linking.NewLinkTokenRegistrar(tokenStore, time.Minute, []byte("secret"))
	api := NewAPI(registrar, linkStore, "secret", ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/telegram/link-token", bytes.NewReader([]byte(`{"user_id":"user-1","token":"token"}`)))
	recorder := httptest.NewRecorder()

	api.HandleLinkToken(recorder, req)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestLinkTokenInvalidPayload(t *testing.T) {
	tokenStore := linking.NewMemoryLinkTokenStore()
	linkStore := linking.NewMemoryTelegramLinkStore()
	registrar := linking.NewLinkTokenRegistrar(tokenStore, time.Minute, []byte("secret"))
	api := NewAPI(registrar, linkStore, "secret", ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/telegram/link-token", bytes.NewReader([]byte(`{"user_id":"","token":""}`)))
	req.Header.Set("Authorization", "Bearer secret")
	recorder := httptest.NewRecorder()

	api.HandleLinkToken(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestLinkTokenSuccess(t *testing.T) {
	tokenStore := linking.NewMemoryLinkTokenStore()
	linkStore := linking.NewMemoryTelegramLinkStore()
	registrar := linking.NewLinkTokenRegistrar(tokenStore, time.Minute, []byte("secret"))
	api := NewAPI(registrar, linkStore, "secret", ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/telegram/link-token", bytes.NewReader([]byte(`{"user_id":"user-1","token":"token"}`)))
	req.Header.Set("Authorization", "Bearer secret")
	recorder := httptest.NewRecorder()

	api.HandleLinkToken(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	var response map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if success, ok := response["success"].(bool); !ok || !success {
		t.Fatalf("expected success=true, got %v", response["success"])
	}
}

func TestStatusNotLinked(t *testing.T) {
	tokenStore := linking.NewMemoryLinkTokenStore()
	linkStore := linking.NewMemoryTelegramLinkStore()
	registrar := linking.NewLinkTokenRegistrar(tokenStore, time.Minute, []byte("secret"))
	api := NewAPI(registrar, linkStore, "secret", ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/telegram/status?user_id=user-1", nil)
	req.Header.Set("Authorization", "Bearer secret")
	recorder := httptest.NewRecorder()

	api.HandleStatus(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	var response map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if linked, ok := response["linked"].(bool); !ok || linked {
		t.Fatalf("expected linked=false, got %v", response["linked"])
	}
}

func TestStatusNotLinkedJSONBody(t *testing.T) {
	tokenStore := linking.NewMemoryLinkTokenStore()
	linkStore := linking.NewMemoryTelegramLinkStore()
	registrar := linking.NewLinkTokenRegistrar(tokenStore, time.Minute, []byte("secret"))
	api := NewAPI(registrar, linkStore, "secret", ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/telegram/status", bytes.NewBufferString(`{"user_id":"user-1"}`))
	req.Header.Set("Authorization", "Bearer secret")
	recorder := httptest.NewRecorder()

	api.HandleStatus(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	var response map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if linked, ok := response["linked"].(bool); !ok || linked {
		t.Fatalf("expected linked=false, got %v", response["linked"])
	}
}

func TestStatusLinked(t *testing.T) {
	tokenStore := linking.NewMemoryLinkTokenStore()
	linkStore := linking.NewMemoryTelegramLinkStore()
	registrar := linking.NewLinkTokenRegistrar(tokenStore, time.Minute, []byte("secret"))
	api := NewAPI(registrar, linkStore, "secret", ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	_ = linkStore.LinkChat(context.Background(), linking.TelegramLink{
		UserID:         "user-1",
		Phone:          "",
		TelegramChatID: 123,
		VerifiedAt:     time.Now(),
	})

	req := httptest.NewRequest(http.MethodGet, "/telegram/status?user_id=user-1", nil)
	req.Header.Set("Authorization", "Bearer secret")
	recorder := httptest.NewRecorder()

	api.HandleStatus(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	var response map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if linked, ok := response["linked"].(bool); !ok || !linked {
		t.Fatalf("expected linked=true, got %v", response["linked"])
	}
}
