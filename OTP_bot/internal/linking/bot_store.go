package linking

import (
	"context"
	"errors"

	"otp_bot/internal/telegram"
)

// BotLinkStore адаптирует TelegramLinkStore для использования ботом.
type BotLinkStore struct {
	store TelegramLinkStore
}

// NewBotLinkStore создает адаптер хранилища связей для бота.
func NewBotLinkStore(store TelegramLinkStore) *BotLinkStore {
	return &BotLinkStore{
		store: store,
	}
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
	return telegram.LinkInfo{UserID: link.UserID, Phone: link.Phone, ChatID: link.TelegramChatID}, nil
}
