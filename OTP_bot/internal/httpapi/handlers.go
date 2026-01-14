package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/profzoom/otp_bot/internal/linking"
	"github.com/profzoom/otp_bot/internal/observability"
	"github.com/profzoom/otp_bot/internal/phone"
	"github.com/profzoom/otp_bot/internal/ratelimit"
	"github.com/profzoom/otp_bot/internal/telegram"
)

const otpMessagePrefix = "ProfZoom login code: "

// API предоставляет HTTP-обработчики для привязки Telegram.
type API struct {
	linkRegistrar       *linking.LinkTokenRegistrar
	linkStore           linking.TelegramLinkStore
	internalKey         string
	linkTokenIPLimiter  ratelimit.Limiter
	linkTokenBotLimiter ratelimit.Limiter
	logger              *slog.Logger
}

// NewAPI собирает набор обработчиков API.
func NewAPI(registrar *linking.LinkTokenRegistrar, linkStore linking.TelegramLinkStore, internalKey string, linkTokenIPLimiter ratelimit.Limiter, linkTokenBotLimiter ratelimit.Limiter, logger *slog.Logger) *API {
	if logger == nil {
		logger = slog.Default()
	}
	if linkTokenIPLimiter == nil {
		linkTokenIPLimiter = ratelimit.NoopLimiter{}
	}
	if linkTokenBotLimiter == nil {
		linkTokenBotLimiter = ratelimit.NoopLimiter{}
	}
	return &API{
		linkRegistrar:       registrar,
		linkStore:           linkStore,
		internalKey:         strings.TrimSpace(internalKey),
		linkTokenIPLimiter:  linkTokenIPLimiter,
		linkTokenBotLimiter: linkTokenBotLimiter,
		logger:              logger,
	}
}

// HandleLinkToken регистрирует токен привязки, переданный основным бэкендом.
func (a *API) HandleLinkToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !requireInternalAuth(w, r, a.internalKey, a.logger) {
		return
	}

	var payload struct {
		Phone string `json:"phone"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_payload")
		return
	}

	normalizedPhone := phone.Normalize(payload.Phone)
	token := strings.TrimSpace(payload.Token)
	if normalizedPhone == "" || token == "" {
		writeError(w, http.StatusBadRequest, "invalid_payload")
		return
	}

	if !a.allowLinkToken(clientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "rate_limited")
		return
	}
	if a.linkRegistrar == nil {
		writeError(w, http.StatusInternalServerError, "link_token_failed")
		return
	}

	if _, err := a.linkRegistrar.Register(r.Context(), token, normalizedPhone); err != nil {
		writeError(w, http.StatusInternalServerError, "link_token_failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

// OTPHandler принимает аутентифицированные запросы на доставку OTP.
type OTPHandler struct {
	sender         telegram.Service
	linkStore      linking.TelegramLinkStore
	internalKey    string
	maxBodyBytes   int64
	perChatLimiter ratelimit.Limiter
	perIPLimiter   ratelimit.Limiter
	botLimiter     ratelimit.Limiter
	logger         *slog.Logger
}

// NewOTPHandler создает обработчик доставки OTP.
func NewOTPHandler(sender telegram.Service, internalKey string, linkStore linking.TelegramLinkStore, perChatLimiter ratelimit.Limiter, perIPLimiter ratelimit.Limiter, botLimiter ratelimit.Limiter, logger *slog.Logger) *OTPHandler {
	if logger == nil {
		logger = slog.Default()
	}
	if perChatLimiter == nil {
		perChatLimiter = ratelimit.NoopLimiter{}
	}
	if perIPLimiter == nil {
		perIPLimiter = ratelimit.NoopLimiter{}
	}
	if botLimiter == nil {
		botLimiter = ratelimit.NoopLimiter{}
	}
	return &OTPHandler{
		sender:         sender,
		linkStore:      linkStore,
		internalKey:    strings.TrimSpace(internalKey),
		maxBodyBytes:   1 << 20,
		perChatLimiter: perChatLimiter,
		perIPLimiter:   perIPLimiter,
		botLimiter:     botLimiter,
		logger:         logger,
	}
}

// ServeHTTP обрабатывает запросы на отправку OTP.
func (h *OTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if !requireInternalAuth(w, r, h.internalKey, h.logger) {
		return
	}

	body := http.MaxBytesReader(w, r.Body, h.maxBodyBytes)
	defer body.Close()

	var payload struct {
		Phone string    `json:"phone"`
		Code  CodeValue `json:"code"`
	}
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		h.logger.Warn("invalid otp payload", slog.String("request_id", observability.RequestIDFromContext(r.Context())))
		writeError(w, http.StatusBadRequest, "invalid_payload")
		return
	}

	normalizedPhone := phone.Normalize(payload.Phone)
	code := strings.TrimSpace(string(payload.Code))
	if normalizedPhone == "" || code == "" {
		writeError(w, http.StatusBadRequest, "invalid_payload")
		return
	}

	h.logger.Info("otp send attempt", slog.String("request_id", observability.RequestIDFromContext(r.Context())), slog.String("phone", normalizedPhone), slog.String("result", "attempt"))

	if h.linkStore == nil {
		h.logger.Error("otp link store missing", slog.String("request_id", observability.RequestIDFromContext(r.Context())), slog.String("phone", normalizedPhone))
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	link, err := h.linkStore.GetByPhone(r.Context(), normalizedPhone)
	if err != nil {
		if errors.Is(err, linking.ErrTelegramLinkNotFound) {
			h.logger.Warn("otp phone not linked", slog.String("request_id", observability.RequestIDFromContext(r.Context())), slog.String("phone", normalizedPhone), slog.String("result", "not_linked"))
			writeError(w, http.StatusBadRequest, "phone_not_linked")
			return
		}
		h.logger.Error("otp link lookup failed", slog.String("request_id", observability.RequestIDFromContext(r.Context())), slog.String("phone", normalizedPhone), slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal_error")
		return
	}

	if !h.allowSend(link.TelegramChatID, clientIP(r)) {
		h.logger.Warn("otp rate limit exceeded", slog.String("request_id", observability.RequestIDFromContext(r.Context())), slog.Int64("chat_id", link.TelegramChatID), slog.String("result", "rate_limited"))
		writeError(w, http.StatusTooManyRequests, "rate_limited")
		return
	}

	message := otpMessagePrefix + code
	if err := h.sender.SendMessage(r.Context(), link.TelegramChatID, message); err != nil {
		h.logger.Error("failed to send otp", slog.String("request_id", observability.RequestIDFromContext(r.Context())), slog.Int64("chat_id", link.TelegramChatID), slog.String("result", "failed"), slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "telegram_failed")
		return
	}

	h.logger.Info("otp sent", slog.String("request_id", observability.RequestIDFromContext(r.Context())), slog.Int64("chat_id", link.TelegramChatID), slog.String("result", "sent"))
	writeJSON(w, http.StatusOK, map[string]any{"sent": true})
}

func (h *OTPHandler) allowSend(chatID int64, ip string) bool {
	if !h.perChatLimiter.Allow(fmt.Sprintf("chat:%d", chatID)) {
		return false
	}
	if ip != "" && !h.perIPLimiter.Allow("ip:"+ip) {
		return false
	}
	return h.botLimiter.Allow("bot")
}

// HandleStatus возвращает статус привязки для номера телефона.
func (a *API) HandleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !requireInternalAuth(w, r, a.internalKey, a.logger) {
		return
	}

	normalizedPhone := phone.Normalize(r.URL.Query().Get("phone"))
	if normalizedPhone == "" {
		var payload struct {
			Phone string `json:"phone"`
		}
		body := http.MaxBytesReader(w, r.Body, 1<<20)
		defer body.Close()
		if err := json.NewDecoder(body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "invalid_payload")
			return
		}
		normalizedPhone = phone.Normalize(payload.Phone)
		if normalizedPhone == "" {
			writeError(w, http.StatusBadRequest, "missing_phone")
			return
		}
	}

	_, err := a.linkStore.GetByPhone(r.Context(), normalizedPhone)
	if err != nil {
		if errors.Is(err, linking.ErrTelegramLinkNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"linked": false})
			return
		}
		writeError(w, http.StatusInternalServerError, "status_failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"linked": true})
}

func (a *API) allowLinkToken(ip string) bool {
	if ip != "" && !a.linkTokenIPLimiter.Allow("ip:"+ip) {
		return false
	}
	return a.linkTokenBotLimiter.Allow("bot")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]any{"error": code})
}

// CodeValue принимает JSON строку или число.
type CodeValue string

// UnmarshalJSON допускает числовые или строковые OTP-коды.
func (c *CodeValue) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return errors.New("empty code")
	}
	if data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*c = CodeValue(value)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*c = CodeValue(num.String())
	return nil
}
