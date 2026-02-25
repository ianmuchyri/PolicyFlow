package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Claims holds the JWT payload for session tokens.
type Claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Role  string `json:"role"`
	Type  string `json:"type"`
}

const (
	CtxUserID    = "user_id"
	CtxUserEmail = "user_email"
	CtxUserRole  = "user_role"
)

// Auth provides JWT-based authentication middleware.
type Auth struct {
	secret []byte
}

func NewAuth(secret string) *Auth {
	return &Auth{secret: []byte(secret)}
}

// Require validates the Bearer token and stores claims in the Echo context.
func (a *Auth) Require(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := extractToken(c.Request())
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
		}

		claims, err := a.parseSession(token)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		c.Set(CtxUserID, claims.Subject)
		c.Set(CtxUserEmail, claims.Email)
		c.Set(CtxUserRole, claims.Role)
		return next(c)
	}
}

// RequireAdmin enforces the Admin role. Must follow Require.
func (a *Auth) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Get(CtxUserRole) != "Admin" {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		return next(c)
	}
}

func (a *Auth) parseSession(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, echo.ErrUnauthorized
		}
		return a.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims.Type != "session" {
		return nil, echo.ErrUnauthorized
	}
	return claims, nil
}

func extractToken(r *http.Request) string {
	// Authorization: Bearer <token>
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	// Fallback: query param (for magic link redirects only — not used for session)
	return r.URL.Query().Get("token")
}
