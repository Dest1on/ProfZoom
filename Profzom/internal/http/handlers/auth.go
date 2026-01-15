package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"profzom/internal/app"
	"profzom/internal/common"
	"profzom/internal/http/middleware"
	"profzom/internal/http/response"
)

type AuthHandler struct {
	auth    *app.AuthService
	limiter *middleware.RateLimiter
	telegramBotUsername string
}

func NewAuthHandler(auth *app.AuthService, limiter *middleware.RateLimiter, telegramBotUsername string) *AuthHandler {
	return &AuthHandler{auth: auth, limiter: limiter, telegramBotUsername: telegramBotUsername}
}

type requestOTPRequest struct {
	Phone string `json:"phone"`
}

type verifyOTPRequest struct {
	Phone string          `json:"phone"`
	Code  string          `json:"code"`
	Role  json.RawMessage `json:"role,omitempty"`
}

type verifyResponse struct {
	Token     string `json:"token"`
	IsNewUser bool   `json:"is_new_user"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type requestOTPResponse struct {
	Success        bool   `json:"success"`
	NeedLink       bool   `json:"need_link,omitempty"`
	TelegramToken  string `json:"telegram_token,omitempty"`
	TelegramLink   string `json:"telegram_link,omitempty"`
}

var (
	phonePattern = regexp.MustCompile(`^\+[0-9]{7,15}$`)
	otpPattern   = regexp.MustCompile(`^[0-9]{6}$`)
)

func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
	var req requestOTPRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	phone := strings.TrimSpace(req.Phone)
	fields := map[string]string{}
	if phone == "" {
		fields["phone"] = "phone is required"
	} else if !phonePattern.MatchString(phone) {
		fields["phone"] = "invalid phone format"
	}
	if len(fields) > 0 {
		response.Error(w, common.NewValidationError("invalid request", fields))
		return
	}
	if h.limiter != nil {
		ipKey := "otp:ip:" + middleware.ClientIP(r)
		if !h.limiter.Allow(ipKey, 5, time.Minute) {
			response.Error(w, common.NewError(common.CodeRateLimited, "otp rate limit exceeded", nil))
			return
		}
		phoneKey := "otp:phone:" + phone
		if !h.limiter.Allow(phoneKey, 3, time.Minute) {
			response.Error(w, common.NewError(common.CodeRateLimited, "otp rate limit exceeded", nil))
			return
		}
	}
	result, err := h.auth.RequestOTP(r.Context(), phone)
	if err != nil {
		response.Error(w, err)
		return
	}
	if result != nil && result.NeedLink {
		resp := requestOTPResponse{
			Success:       false,
			NeedLink:      true,
			TelegramToken: result.TelegramToken,
		}
		if link := buildTelegramLink(h.telegramBotUsername, result.TelegramToken); link != "" {
			resp.TelegramLink = link
		}
		response.JSON(w, http.StatusOK, resp)
		return
	}
	response.JSON(w, http.StatusOK, requestOTPResponse{Success: true})
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := decodeJSON(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	phone := strings.TrimSpace(req.Phone)
	code := strings.TrimSpace(req.Code)
	fields := map[string]string{}
	if len(req.Role) > 0 {
		fields["role"] = "role is not allowed"
	}
	if phone == "" {
		fields["phone"] = "phone is required"
	} else if !phonePattern.MatchString(phone) {
		fields["phone"] = "invalid phone format"
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
		phoneKey := "otp-verify:phone:" + phone
		if !h.limiter.Allow(phoneKey, 5, time.Minute) {
			response.Error(w, common.NewError(common.CodeRateLimited, "otp rate limit exceeded", nil))
			return
		}
	}
	pair, _, isNewUser, err := h.auth.VerifyOTP(r.Context(), phone, code)
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

func buildTelegramLink(username, token string) string {
	if username == "" || token == "" {
		return ""
	}
	trimmed := strings.TrimPrefix(strings.TrimSpace(username), "@")
	if trimmed == "" {
		return ""
	}
	return "https://t.me/" + trimmed + "?start=" + token
}
