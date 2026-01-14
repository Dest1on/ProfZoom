package handlers

import (
    "net/http"
    "strings"

    "github.com/Dest1on/ProfZoom-backend/internal/app"
    "github.com/Dest1on/ProfZoom-backend/internal/common"
    "github.com/Dest1on/ProfZoom-backend/internal/domain/user"
    "github.com/Dest1on/ProfZoom-backend/internal/http/middleware"
    "github.com/Dest1on/ProfZoom-backend/internal/http/response"
)

type UserHandler struct {
    users *app.UserService
}

func NewUserHandler(users *app.UserService) *UserHandler {
    return &UserHandler{users: users}
}

type setRoleRequest struct {
    Role string `json:"role"`
}

func (h *UserHandler) SetRole(w http.ResponseWriter, r *http.Request) {
    userID, ok := middleware.UserIDFromContext(r.Context())
    if !ok {
        response.Error(w, errUnauthorized())
        return
    }
    var req setRoleRequest
    if err := decodeJSON(r, &req); err != nil {
        response.Error(w, err)
        return
    }
	role := strings.TrimSpace(req.Role)
	if role == "" {
		response.Error(w, common.NewValidationError("invalid request", map[string]string{"role": "role is required"}))
		return
	}
	normalized := strings.ToLower(role)
	if err := h.users.SetRole(r.Context(), userID, user.Role(normalized)); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"role": normalized})
}
