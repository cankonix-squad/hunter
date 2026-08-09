package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/cankonix/cankonix-oi/services/api/internal/identity"
)

type principalContextKey struct{}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var body credentials
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	tokens, err := s.identity.Login(r.Context(), body.Email, body.Password)
	if errors.Is(err, identity.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "Email or password is invalid")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Login is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	var body refreshRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	tokens, err := s.identity.Refresh(r.Context(), body.RefreshToken)
	if errors.Is(err, identity.ErrInvalidRefreshToken) {
		writeError(w, http.StatusUnauthorized, "invalid_refresh_token", "Refresh token is invalid or expired")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Token refresh is temporarily unavailable")
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	var body refreshRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return
	}
	if err := s.identity.Revoke(r.Context(), body.RefreshToken); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Logout is temporarily unavailable")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, http.StatusUnauthorized, "authentication_required", "A bearer access token is required")
			return
		}
		p, err := s.identity.ParseAccess(parts[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid_access_token", "Access token is invalid or expired")
			return
		}
		next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
	})
}

func requireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := principalFrom(r)
			if !ok {
				writeError(w, http.StatusUnauthorized, "authentication_required", "Authentication is required")
				return
			}
			for _, role := range p.Roles {
				if _, ok := allowed[role]; ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeError(w, http.StatusForbidden, "forbidden", "Your role cannot perform this action")
		})
	}
}
