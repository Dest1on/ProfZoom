package linking

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"github.com/profzoom/otp_bot/internal/telegram"
)

// ErrRateLimited сообщает о превышении лимита запросов.
var ErrRateLimited = errors.New("rate_limited")

// RateLimiter ограничивает число запросов на привязку для одного субъекта.
type RateLimiter interface {
	Allow(key string) bool
}

// MemoryRateLimiter — лимитер token bucket для разработки.
type MemoryRateLimiter struct {
	mu       sync.Mutex
	capacity int
	refill   time.Duration
	buckets  map[string]*bucket
}

type bucket struct {
	remaining int
	resetAt   time.Time
}

// NewMemoryRateLimiter разрешает не более capacity запросов за интервал пополнения.
func NewMemoryRateLimiter(capacity int, refill time.Duration) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		capacity: capacity,
		refill:   refill,
		buckets:  make(map[string]*bucket),
	}
}

func (l *MemoryRateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		l.buckets[key] = &bucket{remaining: l.capacity - 1, resetAt: now.Add(l.refill)}
		return true
	}

	if now.After(b.resetAt) {
		b.remaining = l.capacity - 1
		b.resetAt = now.Add(l.refill)
		return true
	}

	if b.remaining <= 0 {
		return false
	}

	b.remaining--
	return true
}

// LinkTokenIssuer создает одноразовые токены для привязки аккаунтов Telegram.
type LinkTokenIssuer struct {
	store      LinkTokenStore
	limiter    RateLimiter
	ttl        time.Duration
	hashSecret []byte
	clock      func() time.Time
}

// NewLinkTokenIssuer создает новый генератор токенов.
func NewLinkTokenIssuer(store LinkTokenStore, limiter RateLimiter, ttl time.Duration, hashSecret []byte) *LinkTokenIssuer {
	return &LinkTokenIssuer{
		store:      store,
		limiter:    limiter,
		ttl:        ttl,
		hashSecret: hashSecret,
		clock:      time.Now,
	}
}

// Issue создает новый одноразовый токен и сохраняет его хэш.
func (s *LinkTokenIssuer) Issue(ctx context.Context, userID, phone string) (string, time.Time, error) {
	if s.limiter != nil && !s.limiter.Allow(userID) {
		return "", time.Time{}, ErrRateLimited
	}
	if len(s.hashSecret) == 0 {
		return "", time.Time{}, errors.New("hash secret required")
	}
	token, err := generateToken()
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := s.clock().Add(s.ttl)
	record := LinkToken{
		TokenHash: hashToken(token, s.hashSecret),
		UserID:    userID,
		Phone:     phone,
		ExpiresAt: expiresAt,
	}
	if err := s.store.Save(ctx, record); err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

// LinkTokenRegistrar хранит токен, переданный основным бэкендом.
type LinkTokenRegistrar struct {
	store      LinkTokenStore
	ttl        time.Duration
	hashSecret []byte
	clock      func() time.Time
}

// NewLinkTokenRegistrar создает регистратор токенов.
func NewLinkTokenRegistrar(store LinkTokenStore, ttl time.Duration, hashSecret []byte) *LinkTokenRegistrar {
	return &LinkTokenRegistrar{
		store:      store,
		ttl:        ttl,
		hashSecret: hashSecret,
		clock:      time.Now,
	}
}

// Register сохраняет одноразовый токен для номера телефона.
func (r *LinkTokenRegistrar) Register(ctx context.Context, token, phone string) (time.Time, error) {
	if token == "" || phone == "" {
		return time.Time{}, errors.New("token and phone required")
	}
	if len(r.hashSecret) == 0 {
		return time.Time{}, errors.New("hash secret required")
	}
	expiresAt := r.clock().Add(r.ttl)
	record := LinkToken{
		TokenHash: hashToken(token, r.hashSecret),
		UserID:    phone,
		Phone:     phone,
		ExpiresAt: expiresAt,
	}
	if err := r.store.Save(ctx, record); err != nil {
		return time.Time{}, err
	}
	return expiresAt, nil
}

// TelegramLinker проверяет токены и связывает ID чатов с телефонами.
type TelegramLinker struct {
	store      LinkTokenStore
	linkStore  TelegramLinkStore
	clock      func() time.Time
	maxSkew    time.Duration
	verifyTime bool
	hashSecret []byte
}

// NewTelegramLinker создает обработчик проверки для Telegram.
func NewTelegramLinker(store LinkTokenStore, linkStore TelegramLinkStore, hashSecret []byte) *TelegramLinker {
	return &TelegramLinker{
		store:      store,
		linkStore:  linkStore,
		clock:      time.Now,
		maxSkew:    0,
		verifyTime: true,
		hashSecret: hashSecret,
	}
}

// VerifyAndLink проверяет токен и привязывает ID чата.
func (l *TelegramLinker) VerifyAndLink(ctx context.Context, token string, chatID int64) (string, error) {
	if token == "" {
		return "", telegram.ErrInvalidToken
	}
	stored, err := l.store.Consume(ctx, hashToken(token, l.hashSecret))
	if err != nil {
		if errors.Is(err, ErrLinkTokenNotFound) {
			return "", telegram.ErrInvalidToken
		}
		return "", err
	}
	if l.verifyTime && l.clock().After(stored.ExpiresAt.Add(l.maxSkew)) {
		return "", telegram.ErrInvalidToken
	}

	link := TelegramLink{
		UserID:         stored.UserID,
		Phone:          stored.Phone,
		TelegramChatID: chatID,
		VerifiedAt:     l.clock(),
	}
	if err := l.linkStore.LinkChat(ctx, link); err != nil {
		return "", err
	}

	return stored.Phone, nil
}

func hashToken(token string, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(token))
	return mac.Sum(nil)
}

func generateToken() (string, error) {
	seed := make([]byte, 24)
	if _, err := rand.Read(seed); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(seed), nil
}
