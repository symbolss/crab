package middleware

import (
	"context"
	"net/http"

	"github.com/family/crab-server/internal/services"
	"github.com/family/crab-server/pkg/response"
	"github.com/google/uuid"
)

// AuthMiddleware validates JWT tokens and sets user info in context
type AuthMiddleware struct {
	svc *services.FamilyService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(svc *services.FamilyService) *AuthMiddleware {
	return &AuthMiddleware{svc: svc}
}

// RequireAuth is a middleware that requires valid authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header required")
			return
		}

		// Extract Bearer token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid authorization header format")
			return
		}

		tokenString := authHeader[7:]

		// Validate token
		claims, err := m.svc.ValidateToken(tokenString)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
			return
		}

		// Extract user ID and role
		sub, ok := claims["sub"].(string)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid token claims")
			return
		}

		userID, err := uuid.Parse(sub)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid user ID in token")
			return
		}

		role, ok := claims["role"].(string)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid role in token")
			return
		}

		// Add user info to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userId", userID)
		ctx = context.WithValue(ctx, "role", role)

		// For parent tokens, the userId is the familyId
		if role == "parent" {
			ctx = context.WithValue(ctx, "familyId", userID)
		} else if role == "child" {
			// For child tokens, look up the familyId from the database
			child, err := m.svc.GetChildByID(ctx, userID)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "child not found")
				return
			}
			ctx = context.WithValue(ctx, "familyId", child.FamilyID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireParent is a middleware that requires parent role
func (m *AuthMiddleware) RequireParent(next http.Handler) http.Handler {
	return m.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value("role").(string)
		if !ok || role != "parent" {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "Parent access required")
			return
		}
		next.ServeHTTP(w, r)
	}))
}
