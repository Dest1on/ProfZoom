package linking

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/profzoom/otp_bot/internal/telegram"
)

func TestLinkTokenLifecycle(t *testing.T) {
	store := NewMemoryLinkTokenStore()
	linkStore := NewMemoryTelegramLinkStore()
	registrar := NewLinkTokenRegistrar(store, time.Minute, []byte("secret"))
	linker := NewTelegramLinker(store, linkStore, []byte("secret"))

	token := "token-123"
	if _, err := registrar.Register(context.Background(), token, "+15550001111"); err != nil {
		t.Fatalf("register token: %v", err)
	}

	phone, err := linker.VerifyAndLink(context.Background(), token, 101)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if phone != "+15550001111" {
		t.Fatalf("unexpected phone: %s", phone)
	}

	if _, err := linker.VerifyAndLink(context.Background(), token, 101); !errors.Is(err, telegram.ErrInvalidToken) {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func TestLinkTokenExpired(t *testing.T) {
	store := NewMemoryLinkTokenStore()
	linkStore := NewMemoryTelegramLinkStore()
	registrar := NewLinkTokenRegistrar(store, time.Minute, []byte("secret"))
	linker := NewTelegramLinker(store, linkStore, []byte("secret"))
	registrar.clock = func() time.Time { return time.Now().Add(-2 * time.Minute) }
	linker.clock = func() time.Time { return time.Now().Add(2 * time.Minute) }

	token := "token-123"
	if _, err := registrar.Register(context.Background(), token, "+15550001111"); err != nil {
		t.Fatalf("register token: %v", err)
	}

	if _, err := linker.VerifyAndLink(context.Background(), token, 101); !errors.Is(err, telegram.ErrInvalidToken) {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}
