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

// List returns all policies.
// GET /api/policies
func (h *Policy) List(c echo.Context) error {
	policies, err := h.db.ListPolicies()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	if policies == nil {
		policies = []*database.Policy{}
	}

	// Attach acknowledgement status for the current user
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
// GET /api/policies/:id
func (h *Policy) Get(c echo.Context) error {
	policy, err := h.db.GetPolicy(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "policy not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
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
		"policy":         policy,
		"current_version": currentVersion,
		"acknowledged":   acknowledged,
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
// POST /api/policies  (Admin only)
func (h *Policy) Create(c echo.Context) error {
	var body struct {
		Title      string `json:"title"`
		Department string `json:"department"`
	}
	if err := c.Bind(&body); err != nil || body.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}

	policy, err := h.db.CreatePolicy(body.Title, body.Department)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}
	return c.JSON(http.StatusCreated, policy)
}

// Update updates policy metadata and status.
// PUT /api/policies/:id  (Admin only)
func (h *Policy) Update(c echo.Context) error {
	policy, err := h.db.GetPolicy(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "policy not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	var body struct {
		Title      string `json:"title"`
		Status     string `json:"status"`
		Department string `json:"department"`
	}
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	// Apply defaults from existing data
	if body.Title == "" {
		body.Title = policy.Title
	}
	if body.Status == "" {
		body.Status = policy.Status
	}
	if body.Department == "" {
		body.Department = policy.Department
	}

	validStatuses := map[string]bool{"Draft": true, "Review": true, "Published": true, "Archived": true}
	if !validStatuses[body.Status] {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid status")
	}

	if err := h.db.UpdatePolicy(policy.ID, body.Title, body.Status, body.Department); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	updated, _ := h.db.GetPolicy(policy.ID)
	return c.JSON(http.StatusOK, updated)
}

// CreateVersion adds a new version to a policy and sets it as current.
// POST /api/policies/:id/versions  (Admin only)
func (h *Policy) CreateVersion(c echo.Context) error {
	policy, err := h.db.GetPolicy(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound, "policy not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
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
// GET /api/admin/stats  (Admin only)
func (h *Policy) AdminStats(c echo.Context) error {
	stats, err := h.db.GetStats()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "database error")
	}

	// Also return recent acknowledgements per policy
	policies, _ := h.db.ListPolicies()
	type policyAckCount struct {
		PolicyID  string `json:"policy_id"`
		Title     string `json:"title"`
		AckCount  int    `json:"ack_count"`
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
