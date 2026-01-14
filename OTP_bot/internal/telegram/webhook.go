package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/profzoom/otp_bot/internal/observability"
	"github.com/profzoom/otp_bot/internal/phone"
)

const (
	telegramSecretHeader = "X-Telegram-Bot-Api-Secret-Token"
	verifyPrefix         = "verify_"
)

// Update представляет payload обновления Telegram для доставки через webhook.
type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

type Message struct {
	MessageID int64    `json:"message_id"`
	Chat      Chat     `json:"chat"`
	Text      string   `json:"text"`
	From      User     `json:"from"`
	Contact   *Contact `json:"contact,omitempty"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type Contact struct {
	PhoneNumber string `json:"phone_number"`
	UserID      int64  `json:"user_id,omitempty"`
}

// ErrInvalidToken сообщает, что токен верификации неверный или истек.
var ErrInvalidToken = errors.New("verification token invalid or expired")

// ErrLinkNotFound сообщает об отсутствии привязки Telegram.
var ErrLinkNotFound = errors.New("telegram link not found")

// VerificationService проверяет токен верификации и привязывает чат Telegram.
type VerificationService interface {
	VerifyAndLink(ctx context.Context, token string, chatID int64) (string, error)
}

// LinkInfo описывает связанную пару телефон-чат.
type LinkInfo struct {
	Phone  string
	ChatID int64
}

// LinkStore управляет связями телефон-чат для бота.
type LinkStore interface {
	GetByPhone(ctx context.Context, phone string) (LinkInfo, error)
	GetByChatID(ctx context.Context, chatID int64) (LinkInfo, error)
	LinkChat(ctx context.Context, phone string, chatID int64) error
}

// Bot обрабатывает входящие обновления Telegram.
type Bot struct {
	sender    Service
	verifier  VerificationService
	linkStore LinkStore
	logger    *slog.Logger
}

// NewBot создает обработчик бота Telegram.
func NewBot(sender Service, verifier VerificationService, linkStore LinkStore, logger *slog.Logger) *Bot {
	if logger == nil {
		logger = slog.Default()
	}
	return &Bot{
		sender:    sender,
		verifier:  verifier,
		linkStore: linkStore,
		logger:    logger,
	}
}

// HandleUpdate маршрутизирует поддерживаемые команды Telegram.
func (b *Bot) HandleUpdate(ctx context.Context, update Update) error {
	if update.Message == nil {
		return nil
	}

	msg := update.Message
	if msg.Chat.ID <= 0 {
		return nil
	}
	if msg.Chat.Type != "" && msg.Chat.Type != "private" {
		return nil
	}

	if msg.Contact != nil {
		return b.handleContact(ctx, msg)
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return nil
	}

	command, arg := parseCommand(text)
	switch command {
	case "/start":
		return b.handleStart(ctx, msg, arg)
	case "/help":
		return b.sendHelp(ctx, msg.Chat.ID)
	case "/status":
		return b.handleStatus(ctx, msg.Chat.ID)
	default:
		return b.sendMessage(ctx, msg.Chat.ID, "I did not understand. Use /start or /help.", nil)
	}
}

func parseCommand(text string) (string, string) {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "", ""
	}
	command := fields[0]
	if idx := strings.Index(command, "@"); idx != -1 {
		command = command[:idx]
	}
	arg := ""
	if len(fields) > 1 {
		arg = fields[1]
	}
	return command, arg
}

func (b *Bot) handleStart(ctx context.Context, message *Message, arg string) error {
	token := strings.TrimSpace(arg)
	if token == "" {
		return b.sendStart(ctx, message.Chat.ID)
	}
	if strings.HasPrefix(token, verifyPrefix) {
		token = strings.TrimPrefix(token, verifyPrefix)
	}
	if token == "" {
		return b.sendStart(ctx, message.Chat.ID)
	}

	phoneNumber, err := b.verifier.VerifyAndLink(ctx, token, message.Chat.ID)
	if err != nil {
		if errors.Is(err, ErrInvalidToken) {
			b.logger.Info("invalid verification token", slog.Int64("chat_id", message.Chat.ID))
			return b.sendMessage(ctx, message.Chat.ID, "Invalid or expired link. Please request a new one in the app.", b.contactKeyboard())
		}
		b.logger.Error("verification failed", slog.Int64("chat_id", message.Chat.ID), slog.String("error", err.Error()))
		return fmt.Errorf("telegram verification failed: %w", err)
	}

	b.logger.Info("telegram linked", slog.Int64("chat_id", message.Chat.ID))
	msg := fmt.Sprintf("Phone %s linked. You can now receive codes here.", phoneNumber)
	return b.sendMessage(ctx, message.Chat.ID, msg, nil)
}

func (b *Bot) handleStatus(ctx context.Context, chatID int64) error {
	if b.linkStore == nil {
		return b.sendMessage(ctx, chatID, "Link status is unavailable right now.", nil)
	}

	link, err := b.linkStore.GetByChatID(ctx, chatID)
	if err != nil {
		if errors.Is(err, ErrLinkNotFound) {
			return b.sendMessage(ctx, chatID, "Your Telegram is not linked yet. Use /start to link your phone.", b.contactKeyboard())
		}
		return err
	}

	msg := fmt.Sprintf("Your Telegram is linked to %s.", link.Phone)
	return b.sendMessage(ctx, chatID, msg, nil)
}

func (b *Bot) handleContact(ctx context.Context, message *Message) error {
	if b.linkStore == nil {
		return nil
	}

	contact := message.Contact
	if contact == nil {
		return nil
	}
	if contact.UserID != 0 && contact.UserID != message.From.ID {
		return b.sendMessage(ctx, message.Chat.ID, "Please share your own phone number.", nil)
	}

	normalizedPhone := phone.Normalize(contact.PhoneNumber)
	if normalizedPhone == "" {
		return b.sendMessage(ctx, message.Chat.ID, "Unable to read the phone number. Please try again.", nil)
	}

	if link, err := b.linkStore.GetByPhone(ctx, normalizedPhone); err == nil {
		if link.ChatID == message.Chat.ID {
			msg := fmt.Sprintf("This phone is already linked to your Telegram account (%s).", normalizedPhone)
			return b.sendMessage(ctx, message.Chat.ID, msg, nil)
		}
		return b.sendMessage(ctx, message.Chat.ID, "This phone is already linked to a different Telegram account. Please request a new link in the app.", nil)
	} else if !errors.Is(err, ErrLinkNotFound) {
		return err
	}

	if existing, err := b.linkStore.GetByChatID(ctx, message.Chat.ID); err == nil {
		if existing.Phone != normalizedPhone {
			msg := fmt.Sprintf("This Telegram account is already linked to %s. If you need to relink, request a new link in the app.", existing.Phone)
			return b.sendMessage(ctx, message.Chat.ID, msg, nil)
		}
	} else if !errors.Is(err, ErrLinkNotFound) {
		return err
	}

	if err := b.linkStore.LinkChat(ctx, normalizedPhone, message.Chat.ID); err != nil {
		return err
	}

	msg := fmt.Sprintf("Phone %s linked. You can now receive codes here.", normalizedPhone)
	return b.sendMessage(ctx, message.Chat.ID, msg, nil)
}

func (b *Bot) sendStart(ctx context.Context, chatID int64) error {
	text := "Hello! I deliver one-time codes for ProfZoom.\n" +
		"To link your account, open the app and request a Telegram link.\n" +
		"You can also share your phone number here using the button below."
	return b.sendMessage(ctx, chatID, text, b.contactKeyboard())
}

func (b *Bot) sendHelp(ctx context.Context, chatID int64) error {
	text := "I send ProfZoom login codes. Use /start to link your phone, or share your phone number here."
	return b.sendMessage(ctx, chatID, text, b.contactKeyboard())
}

func (b *Bot) contactKeyboard() *ReplyKeyboardMarkup {
	return &ReplyKeyboardMarkup{
		Keyboard: [][]KeyboardButton{{{
			Text:           "Share my phone number",
			RequestContact: true,
		}}},
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}
}

func (b *Bot) sendMessage(ctx context.Context, chatID int64, text string, replyMarkup any) error {
	if replyMarkup != nil {
		if sender, ok := b.sender.(MarkupSender); ok {
			return sender.SendMessageWithMarkup(ctx, chatID, text, replyMarkup)
		}
	}
	return b.sender.SendMessage(ctx, chatID, text)
}

// WebhookHandler проверяет запросы Telegram webhook и передает обновления.
type WebhookHandler struct {
	bot          *Bot
	secretToken  string
	maxBodyBytes int64
	logger       *slog.Logger
}

// NewWebhookHandler создает обработчик webhook, который проверяет секретный токен Telegram.
func NewWebhookHandler(bot *Bot, secretToken string, logger *slog.Logger) *WebhookHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &WebhookHandler{
		bot:          bot,
		secretToken:  secretToken,
		maxBodyBytes: 1 << 20,
		logger:       logger,
	}
}

// ServeHTTP реализует http.Handler для коллбэков Telegram webhook.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.secretToken != "" && r.Header.Get(telegramSecretHeader) != h.secretToken {
		h.logger.Warn("unauthorized webhook request", slog.String("request_id", observability.RequestIDFromContext(r.Context())))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body := http.MaxBytesReader(w, r.Body, h.maxBodyBytes)
	defer body.Close()

	payload, err := io.ReadAll(body)
	if err != nil {
		h.logger.Warn("failed to read webhook body", slog.String("request_id", observability.RequestIDFromContext(r.Context())))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var update Update
	if err := json.Unmarshal(payload, &update); err != nil {
		h.logger.Warn("invalid webhook payload", slog.String("request_id", observability.RequestIDFromContext(r.Context())))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.bot.HandleUpdate(r.Context(), update); err != nil {
		h.logger.Error("failed to handle telegram update", slog.String("request_id", observability.RequestIDFromContext(r.Context())), slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
