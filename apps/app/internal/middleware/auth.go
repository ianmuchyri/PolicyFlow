package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"policyflow/internal/database"
)

// Claims holds the JWT payload for session tokens.
type Claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Role  string `json:"role"`
	Type  string `json:"type"`
}

// Role constants.
const (
	RoleSuperAdmin = "SuperAdmin"
	RoleDeptAdmin  = "DeptAdmin"
	RoleStaff      = "Staff"
)

// Context keys.
const (
	CtxUserID    = "user_id"
	CtxUserEmail = "user_email"
	CtxUserRole  = "user_role"
	CtxDeptID    = "user_dept_id" // *string, may be nil
)

// Auth provides JWT-based authentication middleware.
type Auth struct {
	secret []byte
	db     *database.DB
}

func NewAuth(secret string, db *database.DB) *Auth {
	return &Auth{secret: []byte(secret), db: db}
}

// Require validates the Bearer token, stores claims in the Echo context,
// and fetches the user's department_id from the DB.
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

		// Fetch department_id from DB so handlers can enforce scoping.
		user, err := a.db.GetUserByID(claims.Subject)
		if err == nil {
			c.Set(CtxDeptID, user.DepartmentID) // *string, may be nil
		}

		return next(c)
	}
}

// RequireSuperAdmin enforces the SuperAdmin role. Must follow Require.
func (a *Auth) RequireSuperAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Get(CtxUserRole) != RoleSuperAdmin {
			return echo.NewHTTPError(http.StatusForbidden, "super admin only")
		}
		return next(c)
	}
}

// RequireDeptAdmin enforces SuperAdmin or DeptAdmin role. Must follow Require.
func (a *Auth) RequireDeptAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get(CtxUserRole)
		if role != RoleSuperAdmin && role != RoleDeptAdmin {
			return echo.NewHTTPError(http.StatusForbidden, "admin only")
		}
		return next(c)
	}
}

// RequireAdmin is an alias for RequireDeptAdmin kept for backward compatibility.
func (a *Auth) RequireAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return a.RequireDeptAdmin(next)
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
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return r.URL.Query().Get("token")
}
