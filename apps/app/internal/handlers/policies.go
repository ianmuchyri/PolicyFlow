package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"policyflow/internal/database"
	mw "policyflow/internal/middleware"
)

// Policy handles policy management and acknowledgement endpoints.
type Policy struct {
	db *database.DB
}

func NewPolicy(db *database.DB) *Policy {
	return &Policy{db: db}
}

// List returns policies visible to the current user based on role and department.
// GET /api/policies
func (h *Policy) List(c echo.Context) error {
	role := c.Get(mw.CtxUserRole).(string)
	deptID, _ := c.Get(mw.CtxDeptID).(*string)

	policies, err := h.db.ListPoliciesForUser(role, deptID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if policies == nil {
		policies = []*database.Policy{}
	}

	// Attach acknowledgement status for the current user.
	userID := c.Get(mw.CtxUserID).(string)
	ackMap, _ := h.db.AckStatusForUser(userID)

	type policyWithAck struct {
		*database.Policy
		Acknowledged bool `json:"acknowledged"`
	}
	result := make([]policyWithAck, len(policies))
	for i, p := range policies {
		acked := false
		if p.CurrentVersionID != nil {
			acked = ackMap[*p.CurrentVersionID]
		}
		result[i] = policyWithAck{Policy: p, Acknowledged: acked}
	}

	return c.JSON(http.StatusOK, result)
}

// Get returns a single policy with its current version content.
// Enforces visibility: non-SuperAdmin users cannot access dept-scoped policies outside their dept.
// GET /api/policies/:id
func (h *Policy) Get(c echo.Context) error {
	policy, err := h.db.GetPolicy(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "policy not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	// Enforce visibility for non-SuperAdmin.
	role := c.Get(mw.CtxUserRole).(string)
	if role != mw.RoleSuperAdmin && policy.VisibilityType == "department" {
		deptID, _ := c.Get(mw.CtxDeptID).(*string)
		if deptID == nil || policy.DepartmentID == nil || *deptID != *policy.DepartmentID {
			return echo.NewHTTPError(http.StatusNotFound, "policy not found")
		}
	}

	var currentVersion *database.PolicyVersion
	if policy.CurrentVersionID != nil {
		currentVersion, _ = h.db.GetPolicyVersion(*policy.CurrentVersionID)
	}

	userID := c.Get(mw.CtxUserID).(string)
	acknowledged := false
	if currentVersion != nil {
		acknowledged, _ = h.db.HasAcknowledged(userID, currentVersion.ID)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"policy":          policy,
		"current_version": currentVersion,
		"acknowledged":    acknowledged,
	})
}

// Versions returns all versions for a policy.
// GET /api/policies/:id/versions
func (h *Policy) Versions(c echo.Context) error {
	versions, err := h.db.ListPolicyVersions(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if versions == nil {
		versions = []*database.PolicyVersion{}
	}
	return c.JSON(http.StatusOK, versions)
}

// Acknowledge records a user's acknowledgement of the current policy version.
// POST /api/policies/:id/acknowledge
func (h *Policy) Acknowledge(c echo.Context) error {
	policy, err := h.db.GetPolicy(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "policy not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	if policy.Status != "Published" {
		return echo.NewHTTPError(http.StatusBadRequest, "can only acknowledge published policies")
	}
	if policy.CurrentVersionID == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "policy has no current version")
	}

	userID := c.Get(mw.CtxUserID).(string)
	already, err := h.db.HasAcknowledged(userID, *policy.CurrentVersionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if already {
		return echo.NewHTTPError(http.StatusConflict, "already acknowledged")
	}

	ack, err := h.db.CreateAcknowledgement(userID, *policy.CurrentVersionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusCreated, ack)
}

// Create creates a new policy.
// POST /api/policies
func (h *Policy) Create(c echo.Context) error {
	var body struct {
		Title          string  `json:"title"`
		Department     string  `json:"department"`
		DepartmentID   *string `json:"department_id"`
		VisibilityType string  `json:"visibility_type"`
	}
	if err := c.Bind(&body); err != nil || body.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	if body.VisibilityType == "" {
		body.VisibilityType = "organization"
	}
	validVis := map[string]bool{"organization": true, "department": true}
	if !validVis[body.VisibilityType] {
		return echo.NewHTTPError(http.StatusBadRequest, "visibility_type must be organization or department")
	}

	// DeptAdmin can only create dept-scoped policies for their own department.
	role := c.Get(mw.CtxUserRole).(string)
	if role == mw.RoleDeptAdmin {
		deptID, _ := c.Get(mw.CtxDeptID).(*string)
		if deptID == nil {
			return echo.NewHTTPError(http.StatusForbidden, "department admin must belong to a department")
		}
		body.VisibilityType = "department"
		body.DepartmentID = deptID
	}

	policy, err := h.db.CreatePolicy(body.Title, body.Department, body.DepartmentID, body.VisibilityType)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusCreated, policy)
}

// Update updates policy metadata and status.
// PUT /api/policies/:id
func (h *Policy) Update(c echo.Context) error {
	policy, err := h.db.GetPolicy(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "policy not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	// DeptAdmin can only update their own department's policies.
	role := c.Get(mw.CtxUserRole).(string)
	var callerDeptID *string
	if role == mw.RoleDeptAdmin {
		callerDeptID, _ = c.Get(mw.CtxDeptID).(*string)
		if callerDeptID == nil || policy.DepartmentID == nil || *callerDeptID != *policy.DepartmentID {
			return echo.NewHTTPError(http.StatusForbidden, "cannot edit policies outside your department")
		}
	}

	var body struct {
		Title          string  `json:"title"`
		Status         string  `json:"status"`
		Department     string  `json:"department"`
		DepartmentID   *string `json:"department_id"`
		VisibilityType string  `json:"visibility_type"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	// Apply defaults from existing data.
	if body.Title == "" {
		body.Title = policy.Title
	}
	if body.Status == "" {
		body.Status = policy.Status
	}
	if body.Department == "" {
		body.Department = policy.Department
	}
	if body.VisibilityType == "" {
		body.VisibilityType = policy.VisibilityType
	}
	if body.DepartmentID == nil {
		body.DepartmentID = policy.DepartmentID
	}

	// DeptAdmin cannot escalate visibility or reassign to another department.
	if role == mw.RoleDeptAdmin {
		body.VisibilityType = "department"
		body.DepartmentID = callerDeptID
	}

	validStatuses := map[string]bool{"Draft": true, "Review": true, "Published": true, "Archived": true}
	if !validStatuses[body.Status] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
	}

	if err := h.db.UpdatePolicy(policy.ID, body.Title, body.Status, body.Department, body.DepartmentID, body.VisibilityType); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	updated, _ := h.db.GetPolicy(policy.ID)
	return c.JSON(http.StatusOK, updated)
}

// CreateVersion adds a new version to a policy and sets it as current.
// POST /api/policies/:id/versions
func (h *Policy) CreateVersion(c echo.Context) error {
	policy, err := h.db.GetPolicy(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "policy not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	// DeptAdmin can only add versions to their own department's dept-scoped policies.
	role := c.Get(mw.CtxUserRole).(string)
	if role == mw.RoleDeptAdmin {
		deptID, _ := c.Get(mw.CtxDeptID).(*string)
		if policy.VisibilityType != "department" ||
			deptID == nil || policy.DepartmentID == nil || *deptID != *policy.DepartmentID {
			return echo.NewHTTPError(http.StatusForbidden, "cannot add versions to policies outside your department")
		}
	}

	var body struct {
		Content       string `json:"content"`
		VersionString string `json:"version_string"`
		Changelog     string `json:"changelog"`
	}
	if err := c.Bind(&body); err != nil || body.Content == "" || body.VersionString == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "content and version_string are required")
	}

	version, err := h.db.CreatePolicyVersion(policy.ID, body.Content, body.VersionString, body.Changelog)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	if err := h.db.SetPolicyCurrentVersion(policy.ID, version.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	return c.JSON(http.StatusCreated, version)
}

// AdminStats returns aggregate statistics.
// GET /api/admin/stats
func (h *Policy) AdminStats(c echo.Context) error {
	stats, err := h.db.GetStats()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	policies, _ := h.db.ListPolicies()
	type policyAckCount struct {
		PolicyID string `json:"policy_id"`
		Title    string `json:"title"`
		AckCount int    `json:"ack_count"`
	}
	var ackCounts []policyAckCount
	for _, p := range policies {
		if p.CurrentVersionID != nil && p.Status == "Published" {
			acks, _ := h.db.ListAcknowledgements(*p.CurrentVersionID)
			ackCounts = append(ackCounts, policyAckCount{
				PolicyID: p.ID,
				Title:    p.Title,
				AckCount: len(acks),
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"stats":      stats,
		"ack_counts": ackCounts,
	})
}
