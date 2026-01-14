package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/profzoom/otp_bot/internal/linking"
	"github.com/profzoom/otp_bot/internal/ratelimit"
)

type otpSender struct {
	called bool
	err    error
	text   string
}

func (o *otpSender) SendMessage(ctx context.Context, chatID int64, text string) error {
	o.called = true
	o.text = text
	return o.err
}

type denyLimiter struct{}

func (denyLimiter) Allow(string) bool { return false }

func TestOTPSendUnauthorized(t *testing.T) {
	sender := &otpSender{}
	linkStore := linking.NewMemoryTelegramLinkStore()
	handler := NewOTPHandler(sender, "secret", linkStore, ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/otp/send", bytes.NewBufferString(`{"phone":"+15550001111","code":"123456"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Result().StatusCode)
	}
	if sender.called {
		t.Fatalf("expected otp not sent")
	}
}

func TestOTPSendInvalidPayload(t *testing.T) {
	sender := &otpSender{}
	linkStore := linking.NewMemoryTelegramLinkStore()
	handler := NewOTPHandler(sender, "secret", linkStore, ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/otp/send", bytes.NewBufferString(`{"phone":"","code":""}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Result().StatusCode)
	}
	if sender.called {
		t.Fatalf("expected otp not sent")
	}
}

func TestOTPSendNotLinked(t *testing.T) {
	sender := &otpSender{}
	linkStore := linking.NewMemoryTelegramLinkStore()
	handler := NewOTPHandler(sender, "secret", linkStore, ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/otp/send", bytes.NewBufferString(`{"phone":"+15550001111","code":"123456"}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Result().StatusCode)
	}
	if sender.called {
		t.Fatalf("expected otp not sent")
	}
}

func TestOTPSendRateLimited(t *testing.T) {
	sender := &otpSender{}
	linkStore := linking.NewMemoryTelegramLinkStore()
	_ = linkStore.LinkChat(context.Background(), linking.TelegramLink{
		UserID:         "+15550001111",
		Phone:          "+15550001111",
		TelegramChatID: 3,
		VerifiedAt:     time.Now(),
	})
	handler := NewOTPHandler(sender, "secret", linkStore, denyLimiter{}, ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/otp/send", bytes.NewBufferString(`{"phone":"+15550001111","code":"123456"}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Result().StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Result().StatusCode)
	}
	if sender.called {
		t.Fatalf("expected otp not sent")
	}
}

func TestOTPSendSuccess(t *testing.T) {
	sender := &otpSender{}
	linkStore := linking.NewMemoryTelegramLinkStore()
	_ = linkStore.LinkChat(context.Background(), linking.TelegramLink{
		UserID:         "+15550001111",
		Phone:          "+15550001111",
		TelegramChatID: 4,
		VerifiedAt:     time.Now(),
	})
	handler := NewOTPHandler(sender, "secret", linkStore, ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, ratelimit.NoopLimiter{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/otp/send", bytes.NewBufferString(`{"phone":"+15550001111","code":"123456"}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	if rec.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Result().StatusCode)
	}
	if !sender.called {
		t.Fatalf("expected otp sent")
	}
	if sender.text != "ProfZoom login code: 123456" {
		t.Fatalf("expected otp message sent, got %q", sender.text)
	}
	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if sent, ok := response["sent"].(bool); !ok || !sent {
		t.Fatalf("expected sent=true, got %v", response["sent"])
	}
}
