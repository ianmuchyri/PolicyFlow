package handlers

import (
	"database/sql"
	"errors"
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

// List returns all users. SuperAdmin sees all; DeptAdmin sees own department only.
// GET /api/users
func (h *User) List(c echo.Context) error {
	role := c.Get(mw.CtxUserRole).(string)
	deptID := c.Get(mw.CtxDeptID) // *string or nil

	var users []*database.User
	var err error

	if role == mw.RoleSuperAdmin || deptID == nil {
		users, err = h.db.ListUsers()
	} else {
		users, err = h.db.ListUsersByDepartment(*deptID.(*string))
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if users == nil {
		users = []*database.User{}
	}
	return c.JSON(http.StatusOK, users)
}

// Create creates a new user and sends them a magic-link welcome email.
// POST /api/users
func (h *User) Create(c echo.Context) error {
	var body struct {
		Email        string  `json:"email"`
		Name         string  `json:"name"`
		Role         string  `json:"role"`
		DepartmentID *string `json:"department_id"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if body.Email == "" || body.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and name are required")
	}
	if body.Role == "" {
		body.Role = mw.RoleStaff
	}
	validRoles := map[string]bool{
		mw.RoleSuperAdmin: true,
		mw.RoleDeptAdmin:  true,
		mw.RoleStaff:      true,
	}
	if !validRoles[body.Role] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid role")
	}

	// DeptAdmin can only create users in their own department.
	callerRole := c.Get(mw.CtxUserRole).(string)
	if callerRole == mw.RoleDeptAdmin {
		deptID := c.Get(mw.CtxDeptID)
		if deptID == nil {
			return echo.NewHTTPError(http.StatusForbidden, "department admin must belong to a department")
		}
		body.DepartmentID = deptID.(*string)
		// DeptAdmin cannot create SuperAdmin users.
		if body.Role == mw.RoleSuperAdmin {
			return echo.NewHTTPError(http.StatusForbidden, "cannot create super admin")
		}
	}

	creatorID := c.Get(mw.CtxUserID).(string)
	user, err := h.db.CreateUser(body.Email, body.Name, body.Role, &creatorID, body.DepartmentID)
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "user already exists or database error")
	}

	// Send welcome email with magic link.
	magicToken, err := h.auth.BuildMagicTokenForUser(user.Email)
	if err == nil {
		magicURL := fmt.Sprintf("%s/api/magic-login?token=%s", h.auth.BaseURL(), magicToken)
		_ = h.mailer.SendNewUserWelcome(user.Email, user.Name, magicURL)
	}

	return c.JSON(http.StatusCreated, user)
}

// Update updates an existing user's name, email, role, and department.
// PUT /api/users/:id  (SuperAdmin only)
func (h *User) Update(c echo.Context) error {
	targetID := c.Param("id")
	target, err := h.db.GetUserByID(targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	var body struct {
		Name         string  `json:"name"`
		Email        string  `json:"email"`
		Role         string  `json:"role"`
		DepartmentID *string `json:"department_id"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	// Apply defaults from existing data.
	if body.Name == "" {
		body.Name = target.Name
	}
	if body.Email == "" {
		body.Email = target.Email
	}
	if body.Role == "" {
		body.Role = target.Role
	}

	validRoles := map[string]bool{
		mw.RoleSuperAdmin: true,
		mw.RoleDeptAdmin:  true,
		mw.RoleStaff:      true,
	}
	if !validRoles[body.Role] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid role")
	}

	// Prevent downgrading the last SuperAdmin.
	if target.Role == mw.RoleSuperAdmin && body.Role != mw.RoleSuperAdmin {
		count, err := h.db.CountSuperAdmins()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "database error")
		}
		if count <= 1 {
			return echo.NewHTTPError(http.StatusConflict, "cannot downgrade the last super admin")
		}
	}

	if err := h.db.UpdateUser(targetID, body.Name, body.Email, body.Role, body.DepartmentID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	updated, _ := h.db.GetUserByID(targetID)
	return c.JSON(http.StatusOK, updated)
}

// Delete removes a user.
// DELETE /api/users/:id  (SuperAdmin only)
func (h *User) Delete(c echo.Context) error {
	targetID := c.Param("id")
	callerID := c.Get(mw.CtxUserID).(string)

	if targetID == callerID {
		return echo.NewHTTPError(http.StatusConflict, "cannot delete yourself")
	}

	target, err := h.db.GetUserByID(targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	// Prevent deleting the last SuperAdmin.
	if target.Role == mw.RoleSuperAdmin {
		count, err := h.db.CountSuperAdmins()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "database error")
		}
		if count <= 1 {
			return echo.NewHTTPError(http.StatusConflict, "cannot delete the last super admin")
		}
	}

	if err := h.db.DeleteUser(targetID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.NoContent(http.StatusNoContent)
}
