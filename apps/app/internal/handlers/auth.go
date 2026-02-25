package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"policyflow/internal/database"
	"policyflow/internal/email"
	mw "policyflow/internal/middleware"
)

// Auth handles magic-link authentication.
type Auth struct {
	db        *database.DB
	mailer    *email.Mailer
	jwtSecret []byte
	baseURL   string
}

func NewAuth(db *database.DB, mailer *email.Mailer, jwtSecret string) *Auth {
	base := os.Getenv("BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	return &Auth{
		db:        db,
		mailer:    mailer,
		jwtSecret: []byte(jwtSecret),
		baseURL:   base,
	}
}

// RequestMagicLink sends a login link to the given email address.
// POST /api/magic-link
func (h *Auth) RequestMagicLink(c echo.Context) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&body); err != nil || body.Email == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email required")
	}

	user, err := h.db.GetUserByEmail(body.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Don't reveal whether the email exists
			return c.JSON(http.StatusOK, map[string]string{"message": "if that email is registered, a link has been sent"})
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	magicToken, err := h.buildMagicToken(user.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "token error")
	}

	magicURL := fmt.Sprintf("%s/api/magic-login?token=%s", h.baseURL, magicToken)
	if err := h.mailer.SendMagicLink(user.Email, user.Name, magicURL); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "email error")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "if that email is registered, a link has been sent"})
}

// MagicLogin validates a magic-link token and returns a session JWT.
// GET /api/magic-login?token=JWT
func (h *Auth) MagicLogin(c echo.Context) error {
	tokenStr := c.QueryParam("token")
	if tokenStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token required")
	}

	email, err := h.parseMagicToken(tokenStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired link")
	}

	user, err := h.db.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusUnauthorized, "user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	sessionToken, err := h.buildSessionToken(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "session error")
	}

	// Redirect to the frontend with the session token embedded as a query param.
	// The frontend stores it and redirects to /policies.
	redirectURL := fmt.Sprintf("%s/auth-callback?token=%s", h.baseURL, sessionToken)
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// Me returns the currently authenticated user.
// GET /api/me
func (h *Auth) Me(c echo.Context) error {
	userID := c.Get(mw.CtxUserID).(string)
	user, err := h.db.GetUserByID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, user)
}

// ─── Token helpers ─────────────────────────────────────────────────────────

func (h *Auth) buildMagicToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  email,
		"type": "magic",
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

func (h *Auth) parseMagicToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return h.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "magic" {
		return "", fmt.Errorf("wrong token type")
	}
	email, ok := claims["sub"].(string)
	if !ok || email == "" {
		return "", fmt.Errorf("missing sub")
	}
	return email, nil
}

func (h *Auth) buildSessionToken(user *database.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"type":  "session",
		"exp":   time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}

// BuildMagicTokenForUser is exposed for use by the user creation handler.
func (h *Auth) BuildMagicTokenForUser(email string) (string, error) {
	return h.buildMagicToken(email)
}

func (h *Auth) BaseURL() string {
	return h.baseURL
}
