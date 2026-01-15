package middleware

import (
	"context"
	"net/http"
	"strings"

	"profzom/internal/common"
	"profzom/internal/domain/user"
	"profzom/internal/http/response"
	"profzom/internal/security"
)

type contextKey string

const (
	ContextUserIDKey contextKey = "user_id"
	ContextRolesKey  contextKey = "roles"
)

type AuthMiddleware struct {
	jwt *security.JWTProvider
}

func NewAuthMiddleware(jwt *security.JWTProvider) *AuthMiddleware {
	return &AuthMiddleware{jwt: jwt}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, common.NewError(common.CodeUnauthorized, "missing authorization header", nil))
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Error(w, common.NewError(common.CodeUnauthorized, "invalid authorization header", nil))
			return
		}
		claims, err := m.jwt.Parse(parts[1])
		if err != nil {
			response.Error(w, common.NewError(common.CodeUnauthorized, "invalid token", err))
			return
		}
		userID, err := common.ParseUUID(claims.UserID)
		if err != nil {
			response.Error(w, common.NewError(common.CodeUnauthorized, "invalid user id", err))
			return
		}
		roles := make([]user.Role, 0, len(claims.Roles))
		for _, role := range claims.Roles {
			roles = append(roles, user.Role(role))
		}
		ctx := context.WithValue(r.Context(), ContextUserIDKey, userID)
		ctx = context.WithValue(ctx, ContextRolesKey, roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(role user.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, ok := r.Context().Value(ContextRolesKey).([]user.Role)
			if !ok {
				response.Error(w, common.NewError(common.CodeForbidden, "roles not found", nil))
				return
			}
			for _, userRole := range roles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			response.Error(w, common.NewError(common.CodeForbidden, "insufficient role", nil))
		})
	}
}

func UserIDFromContext(ctx context.Context) (common.UUID, bool) {
	id, ok := ctx.Value(ContextUserIDKey).(common.UUID)
	return id, ok
}
