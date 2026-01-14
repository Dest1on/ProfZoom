package linking

import (
	"context"
	"errors"
	"time"

	"github.com/profzoom/otp_bot/internal/telegram"
)

// BotLinkStore адаптирует TelegramLinkStore для использования ботом.
type BotLinkStore struct {
	store TelegramLinkStore
	clock func() time.Time
}

// NewBotLinkStore создает адаптер хранилища связей для бота.
func NewBotLinkStore(store TelegramLinkStore) *BotLinkStore {
	return &BotLinkStore{
		store: store,
		clock: time.Now,
	}
}

// GetByPhone возвращает связь по телефону.
func (s *BotLinkStore) GetByPhone(ctx context.Context, phone string) (telegram.LinkInfo, error) {
	link, err := s.store.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, ErrTelegramLinkNotFound) {
			return telegram.LinkInfo{}, telegram.ErrLinkNotFound
		}
		return telegram.LinkInfo{}, err
	}
	return telegram.LinkInfo{Phone: link.Phone, ChatID: link.TelegramChatID}, nil
}

// GetByChatID возвращает связь по ID чата.
func (s *BotLinkStore) GetByChatID(ctx context.Context, chatID int64) (telegram.LinkInfo, error) {
	link, err := s.store.GetByChatID(ctx, chatID)
	if err != nil {
		if errors.Is(err, ErrTelegramLinkNotFound) {
			return telegram.LinkInfo{}, telegram.ErrLinkNotFound
		}
		return telegram.LinkInfo{}, err
	}
	return telegram.LinkInfo{Phone: link.Phone, ChatID: link.TelegramChatID}, nil
}

// LinkChat связывает телефон с ID чата.
func (s *BotLinkStore) LinkChat(ctx context.Context, phone string, chatID int64) error {
	link := TelegramLink{
		UserID:         phone,
		Phone:          phone,
		TelegramChatID: chatID,
		VerifiedAt:     s.clock(),
	}
	return s.store.LinkChat(ctx, link)
}
