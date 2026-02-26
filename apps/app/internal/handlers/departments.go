package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"policyflow/internal/database"
)

// Departments handles department management endpoints.
type Departments struct {
	db *database.DB
}

func NewDepartments(db *database.DB) *Departments {
	return &Departments{db: db}
}

// List returns all departments. Available to all authenticated users.
// GET /api/departments
func (h *Departments) List(c echo.Context) error {
	depts, err := h.db.ListDepartments()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if depts == nil {
		depts = []*database.Department{}
	}
	return c.JSON(http.StatusOK, depts)
}

// Create creates a new department.
// POST /api/departments  (SuperAdmin only)
func (h *Departments) Create(c echo.Context) error {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.Bind(&body); err != nil || body.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}

	dept, err := h.db.CreateDepartment(body.Name, body.Description)
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "department already exists or database error")
	}
	return c.JSON(http.StatusCreated, dept)
}

// Update updates a department's name and description.
// PUT /api/departments/:id  (SuperAdmin only)
func (h *Departments) Update(c echo.Context) error {
	id := c.Param("id")
	existing, err := h.db.GetDepartment(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "department not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if body.Name == "" {
		body.Name = existing.Name
	}
	if body.Description == "" {
		body.Description = existing.Description
	}

	dept, err := h.db.UpdateDepartment(id, body.Name, body.Description)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusOK, dept)
}

// Delete removes a department. Returns 409 if policies are still assigned to it.
// DELETE /api/departments/:id  (SuperAdmin only)
func (h *Departments) Delete(c echo.Context) error {
	id := c.Param("id")
	if _, err := h.db.GetDepartment(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "department not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	hasPolicies, err := h.db.DepartmentHasPolicies(id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if hasPolicies {
		return echo.NewHTTPError(http.StatusConflict, "department has assigned policies; reassign them first")
	}

	if err := h.db.DeleteDepartment(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.NoContent(http.StatusNoContent)
}
