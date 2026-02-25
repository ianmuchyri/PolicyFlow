package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"policyflow/internal/database"
	"policyflow/internal/email"
	mw "policyflow/internal/middleware"
)

// User handles user management endpoints (admin-only).
type User struct {
	db     *database.DB
	mailer *email.Mailer
	auth   *Auth
}

func NewUser(db *database.DB, mailer *email.Mailer, jwtSecret string) *User {
	return &User{
		db:     db,
		mailer: mailer,
		auth:   NewAuth(db, mailer, jwtSecret),
	}
}

// Create creates a new user and sends them a magic-link welcome email.
// POST /api/users  (Admin only)
func (h *User) Create(c echo.Context) error {
	var body struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Role  string `json:"role"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if body.Email == "" || body.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and name are required")
	}
	if body.Role == "" {
		body.Role = "Staff"
	}
	if body.Role != "Admin" && body.Role != "Staff" {
		return echo.NewHTTPError(http.StatusBadRequest, "role must be Admin or Staff")
	}

	creatorID := c.Get(mw.CtxUserID).(string)
	user, err := h.db.CreateUser(body.Email, body.Name, body.Role, &creatorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "user already exists or database error")
	}

	// Send welcome email with magic link
	magicToken, err := h.auth.BuildMagicTokenForUser(user.Email)
	if err == nil {
		magicURL := fmt.Sprintf("%s/api/magic-login?token=%s", h.auth.BaseURL(), magicToken)
		_ = h.mailer.SendNewUserWelcome(user.Email, user.Name, magicURL)
	}

	return c.JSON(http.StatusCreated, user)
}

// List returns all users.
// GET /api/users  (Admin only)
func (h *User) List(c echo.Context) error {
	users, err := h.db.ListUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if users == nil {
		users = []*database.User{}
	}
	return c.JSON(http.StatusOK, users)
}
