package telegram

import (
	"context"
	"errors"
	"log/slog"
	"testing"
)

type fakeSender struct {
	lastChatID int64
	lastText   string
}

func (f *fakeSender) SendMessage(ctx context.Context, chatID int64, text string) error {
	f.lastChatID = chatID
	f.lastText = text
	return nil
}

type fakeVerifier struct {
	phone string
	err   error
}

func (f fakeVerifier) VerifyAndLink(ctx context.Context, token string, chatID int64) (string, error) {
	return f.phone, f.err
}

func TestBotHandleStartSuccess(t *testing.T) {
	sender := &fakeSender{}
	verifier := fakeVerifier{phone: "+15550001111"}
	bot := NewBot(sender, verifier, nil, slog.Default())

	update := Update{Message: &Message{Chat: Chat{ID: 42, Type: "private"}, Text: "/start token"}}
	if err := bot.HandleUpdate(context.Background(), update); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.lastChatID != 42 {
		t.Fatalf("expected chat id 42, got %d", sender.lastChatID)
	}
	if sender.lastText == "" {
		t.Fatalf("expected response text")
	}
}

func TestBotHandleStartInvalidToken(t *testing.T) {
	sender := &fakeSender{}
	verifier := fakeVerifier{err: ErrInvalidToken}
	bot := NewBot(sender, verifier, nil, slog.Default())

	update := Update{Message: &Message{Chat: Chat{ID: 99, Type: "private"}, Text: "/start token"}}
	if err := bot.HandleUpdate(context.Background(), update); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.lastText == "" {
		t.Fatalf("expected response text")
	}
}

func TestBotHandleStartBackendError(t *testing.T) {
	sender := &fakeSender{}
	verifier := fakeVerifier{err: errors.New("backend down")}
	bot := NewBot(sender, verifier, nil, slog.Default())

	update := Update{Message: &Message{Chat: Chat{ID: 7, Type: "private"}, Text: "/start token"}}
	if err := bot.HandleUpdate(context.Background(), update); err == nil {
		t.Fatalf("expected error")
	}
}
