package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"profzom/internal/app"
	"profzom/internal/common"
	"profzom/internal/domain/auth"
	"profzom/internal/http/middleware"
	"profzom/internal/http/response"
)

type AuthHandler struct {
	auth        *app.AuthService
	limiter     *middleware.RateLimiter
	internalKey string
}

func NewAuthHandler(auth *app.AuthService, limiter *middleware.RateLimiter, internalKey string) *AuthHandler {
	return &AuthHandler{auth: auth, limiter: limiter, internalKey: internalKey}
}

type registerResponse struct {
	UserID   string `json:"user_id"`
	LinkCode string `json:"link_code"`
}

type verifyOTPRequest struct {
	UserID     string          `json:"user_id"`
	TelegramID int64           `json:"telegram_id,omitempty"`
	Code       string          `json:"code"`
	Role       json.RawMessage `json:"role,omitempty"`
}

type verifyResponse struct {
	Token     string `json:"token"`
	IsNewUser bool   `json:"is_new_user"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type requestOTPByTelegramRequest struct {
	TelegramID int64 `json:"telegram_id"`
}

type requestOTPByTelegramResponse struct {
	Code      string `json:"code"`
	ExpiresAt string `json:"expires_at"`
}

var otpPattern = regexp.MustCompile(`^[0-9]{6}$`)

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	result, err := h.auth.Register(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, registerResponse{UserID: result.UserID.String(), LinkCode: result.LinkCode})
}

func (h *AuthHandler) RequestOTPByTelegram(w http.ResponseWriter, r *http.Request) {
	if !requireInternalAuth(w, r, h.internalKey) {
		return
	}
	var req requestOTPByTelegramRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if req.TelegramID <= 0 {
		response.Error(w, common.NewValidationError("invalid request", map[string]string{"telegram_id": "telegram_id is required"}))
		return
	}
	if h.limiter != nil {
		ipKey := "otp:tg:ip:" + middleware.ClientIP(r)
		if !h.limiter.Allow(ipKey, 5, time.Minute) {
			response.Error(w, common.NewError(common.CodeRateLimited, "otp rate limit exceeded", nil))
			return
		}
		tgKey := fmt.Sprintf("otp:tg:%d", req.TelegramID)
		if !h.limiter.Allow(tgKey, 3, time.Minute) {
			response.Error(w, common.NewError(common.CodeRateLimited, "otp rate limit exceeded", nil))
			return
		}
	}
	result, err := h.auth.RequestOTPByTelegram(r.Context(), req.TelegramID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, requestOTPByTelegramResponse{Code: result.Code, ExpiresAt: result.ExpiresAt.Format(time.RFC3339)})
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	userID := strings.TrimSpace(req.UserID)
	telegramID := req.TelegramID
	code := strings.TrimSpace(req.Code)
	fields := map[string]string{}
	if len(req.Role) > 0 {
		fields["role"] = "role is not allowed"
	}
	if userID != "" && telegramID != 0 {
		fields["user_id"] = "user_id is not allowed when telegram_id is provided"
		fields["telegram_id"] = "telegram_id is not allowed when user_id is provided"
	}
	if userID == "" && telegramID == 0 {
		fields["user_id"] = "user_id is required"
		fields["telegram_id"] = "telegram_id is required"
	}
	if userID != "" {
		if _, err := common.ParseUUID(userID); err != nil {
			fields["user_id"] = "invalid user_id"
		}
	}
	if telegramID < 0 {
		fields["telegram_id"] = "invalid telegram_id"
	}
	if code == "" {
		fields["code"] = "code is required"
	} else if !otpPattern.MatchString(code) {
		fields["code"] = "invalid code format"
	}
	if len(fields) > 0 {
		response.Error(w, common.NewValidationError("invalid request", fields))
		return
	}
	if h.limiter != nil {
		ipKey := "otp-verify:ip:" + middleware.ClientIP(r)
		if !h.limiter.Allow(ipKey, 10, time.Minute) {
			response.Error(w, common.NewError(common.CodeRateLimited, "otp rate limit exceeded", nil))
			return
		}
		if telegramID != 0 {
			tgKey := fmt.Sprintf("otp-verify:tg:%d", telegramID)
			if !h.limiter.Allow(tgKey, 5, time.Minute) {
				response.Error(w, common.NewError(common.CodeRateLimited, "otp rate limit exceeded", nil))
				return
			}
		} else {
			userKey := "otp-verify:user:" + userID
			if !h.limiter.Allow(userKey, 5, time.Minute) {
				response.Error(w, common.NewError(common.CodeRateLimited, "otp rate limit exceeded", nil))
				return
			}
		}
	}
	var (
		pair      *auth.TokenPair
		isNewUser bool
		err       error
	)
	if telegramID != 0 {
		pair, _, isNewUser, err = h.auth.VerifyOTPByTelegram(r.Context(), telegramID, code)
	} else {
		pair, _, isNewUser, err = h.auth.VerifyOTP(r.Context(), userID, code)
	}
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, verifyResponse{Token: pair.AccessToken, IsNewUser: isNewUser})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	pair, err := h.auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"access_token": pair.AccessToken, "refresh_token": pair.RefreshToken, "expires_at": pair.ExpiresAt.Format(time.RFC3339)})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := h.auth.Logout(r.Context(), req.RefreshToken); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}
