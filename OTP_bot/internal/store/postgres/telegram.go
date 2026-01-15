package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"
	"otp_bot/internal/linking"
)

// TelegramLinkStore хранит привязки Telegram в Postgres.
type TelegramLinkStore struct {
	db *sql.DB
}

// NewTelegramLinkStore создает новый TelegramLinkStore.
func NewTelegramLinkStore(db *sql.DB) *TelegramLinkStore {
	return &TelegramLinkStore{db: db}
}

func (s *TelegramLinkStore) GetByPhone(ctx context.Context, phone string) (linking.TelegramLink, error) {
	const query = `
		SELECT user_id, phone, chat_id, verified_at
		FROM telegram_links
		WHERE phone = $1
	`
	var link linking.TelegramLink
	if err := s.db.QueryRowContext(ctx, query, phone).Scan(&link.UserID, &link.Phone, &link.TelegramChatID, &link.VerifiedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return linking.TelegramLink{}, linking.ErrTelegramLinkNotFound
		}
		return linking.TelegramLink{}, err
	}
	return link, nil
}

func (s *TelegramLinkStore) GetByUserID(ctx context.Context, userID string) (linking.TelegramLink, error) {
	const query = `
		SELECT user_id, phone, chat_id, verified_at
		FROM telegram_links
		WHERE user_id = $1
	`
	var link linking.TelegramLink
	if err := s.db.QueryRowContext(ctx, query, userID).Scan(&link.UserID, &link.Phone, &link.TelegramChatID, &link.VerifiedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return linking.TelegramLink{}, linking.ErrTelegramLinkNotFound
		}
		return linking.TelegramLink{}, err
	}
	return link, nil
}

func (s *TelegramLinkStore) GetByChatID(ctx context.Context, chatID int64) (linking.TelegramLink, error) {
	const query = `
		SELECT user_id, phone, chat_id, verified_at
		FROM telegram_links
		WHERE chat_id = $1
	`
	var link linking.TelegramLink
	if err := s.db.QueryRowContext(ctx, query, chatID).Scan(&link.UserID, &link.Phone, &link.TelegramChatID, &link.VerifiedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return linking.TelegramLink{}, linking.ErrTelegramLinkNotFound
		}
		return linking.TelegramLink{}, err
	}
	return link, nil
}

func (s *TelegramLinkStore) LinkChat(ctx context.Context, link linking.TelegramLink) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM telegram_links WHERE user_id = $1 OR phone = $2 OR chat_id = $3`, link.UserID, link.Phone, link.TelegramChatID); err != nil {
		_ = tx.Rollback()
		return err
	}
	const query = `
		INSERT INTO telegram_links (user_id, phone, chat_id, verified_at)
		VALUES ($1, $2, $3, $4)
	`
	if _, err := tx.ExecContext(ctx, query, link.UserID, link.Phone, link.TelegramChatID, link.VerifiedAt); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// TelegramLinkTokenStore хранит токены привязки в Postgres.
type TelegramLinkTokenStore struct {
	db *sql.DB
}

// NewTelegramLinkTokenStore создает новое хранилище токенов.
func NewTelegramLinkTokenStore(db *sql.DB) *TelegramLinkTokenStore {
	return &TelegramLinkTokenStore{db: db}
}

func (s *TelegramLinkTokenStore) Save(ctx context.Context, token linking.LinkToken) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM telegram_link_tokens WHERE phone = $1`, token.Phone); err != nil {
		_ = tx.Rollback()
		return err
	}
	const query = `
		INSERT INTO telegram_link_tokens (token_hash, user_id, phone, expires_at, consumed_at)
		VALUES ($1, $2, $3, $4, NULL)
		ON CONFLICT (token_hash)
		DO UPDATE SET user_id = EXCLUDED.user_id,
			phone = EXCLUDED.phone,
			expires_at = EXCLUDED.expires_at,
			consumed_at = NULL
	`
	if _, err := tx.ExecContext(ctx, query, token.TokenHash, token.UserID, token.Phone, token.ExpiresAt); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *TelegramLinkTokenStore) Consume(ctx context.Context, tokenHash []byte) (linking.LinkToken, error) {
	const query = `
		UPDATE telegram_link_tokens
		SET consumed_at = NOW()
		WHERE token_hash = $1
			AND consumed_at IS NULL
			AND expires_at > NOW()
		RETURNING user_id, phone, expires_at
	`
	var link linking.LinkToken
	var expiresAt time.Time
	if err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(&link.UserID, &link.Phone, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return linking.LinkToken{}, linking.ErrLinkTokenNotFound
		}
		return linking.LinkToken{}, err
	}
	link.TokenHash = tokenHash
	link.ExpiresAt = expiresAt
	return link, nil
}
