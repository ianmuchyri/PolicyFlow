package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	_ "modernc.org/sqlite"

	"policyflow/internal/database"
	mw "policyflow/internal/middleware"
)

// makeTestDB opens an in-memory SQLite DB, runs Init + Migrate, and returns it.
func makeTestDB(t *testing.T) *database.DB {
	t.Helper()
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { conn.Close() })

	db := database.New(conn)
	if err := db.Init(); err != nil {
		t.Fatalf("db.Init: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("db.Migrate: %v", err)
	}
	return db
}

// makeCtx builds an echo context with role/deptID set — bypassing JWT middleware.
func makeCtx(e *echo.Echo, method, body string, policyID, role string, deptID *string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(policyID)
	c.Set(mw.CtxUserRole, role)
	c.Set(mw.CtxUserID, "test-user-id")
	if deptID != nil {
		c.Set(mw.CtxDeptID, deptID)
	}
	return c, rec
}

func strPtr(s string) *string { return &s }

// ─── Policy.Update() tests ──────────────────────────────────────────────────

// TestDeptAdmin_Update_CannotEscalateVisibility verifies that a DeptAdmin sending
// {"visibility_type":"organization"} on their own dept-scoped policy gets a 200
// response but the policy remains department-scoped.
func TestDeptAdmin_Update_CannotEscalateVisibility(t *testing.T) {
	db := makeTestDB(t)
	dept, _ := db.CreateDepartment("Engineering", "")
	policy, _ := db.CreatePolicy("Test Policy", "", strPtr(dept.ID), "department")

	e := echo.New()
	h := NewPolicy(db)

	body := `{"visibility_type":"organization"}`
	c, rec := makeCtx(e, http.MethodPut, body, policy.ID, mw.RoleDeptAdmin, strPtr(dept.ID))

	if err := h.Update(c); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got database.Policy
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.VisibilityType != "department" {
		t.Errorf("visibility_type = %q; want %q", got.VisibilityType, "department")
	}
}

// TestDeptAdmin_Update_CannotReassignDepartment verifies that a DeptAdmin sending
// a different department_id cannot move a policy to another department.
func TestDeptAdmin_Update_CannotReassignDepartment(t *testing.T) {
	db := makeTestDB(t)
	deptA, _ := db.CreateDepartment("Engineering", "")
	deptB, _ := db.CreateDepartment("HR", "")
	policy, _ := db.CreatePolicy("Test Policy", "", strPtr(deptA.ID), "department")

	e := echo.New()
	h := NewPolicy(db)

	body := `{"department_id":"` + deptB.ID + `"}`
	c, rec := makeCtx(e, http.MethodPut, body, policy.ID, mw.RoleDeptAdmin, strPtr(deptA.ID))

	if err := h.Update(c); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got database.Policy
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.DepartmentID == nil || *got.DepartmentID != deptA.ID {
		t.Errorf("department_id = %v; want %q", got.DepartmentID, deptA.ID)
	}
}

// TestSuperAdmin_Update_CanChangeVisibility verifies that a SuperAdmin CAN change
// visibility_type and department_id freely.
func TestSuperAdmin_Update_CanChangeVisibility(t *testing.T) {
	db := makeTestDB(t)
	deptA, _ := db.CreateDepartment("Engineering", "")
	policy, _ := db.CreatePolicy("Test Policy", "", strPtr(deptA.ID), "department")

	e := echo.New()
	h := NewPolicy(db)

	body := `{"visibility_type":"organization"}`
	c, rec := makeCtx(e, http.MethodPut, body, policy.ID, mw.RoleSuperAdmin, nil)

	if err := h.Update(c); err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got database.Policy
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.VisibilityType != "organization" {
		t.Errorf("visibility_type = %q; want %q", got.VisibilityType, "organization")
	}
}

// ─── Policy.CreateVersion() tests ───────────────────────────────────────────

// TestDeptAdmin_CreateVersion_BlockedOnOrgWidePolicy verifies that a DeptAdmin
// gets a 403 when trying to add a version to an org-wide policy.
func TestDeptAdmin_CreateVersion_BlockedOnOrgWidePolicy(t *testing.T) {
	db := makeTestDB(t)
	dept, _ := db.CreateDepartment("Engineering", "")
	orgPolicy, _ := db.CreatePolicy("Org Policy", "", nil, "organization")

	e := echo.New()
	h := NewPolicy(db)

	body := `{"content":"# Content","version_string":"v1.0.0","changelog":"init"}`
	c, _ := makeCtx(e, http.MethodPost, body, orgPolicy.ID, mw.RoleDeptAdmin, strPtr(dept.ID))

	err := h.CreateVersion(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusForbidden {
		t.Errorf("expected 403 HTTPError, got %v", err)
	}
}

// TestDeptAdmin_CreateVersion_BlockedOnOtherDeptPolicy verifies that a DeptAdmin
// gets a 403 when trying to add a version to another department's policy.
func TestDeptAdmin_CreateVersion_BlockedOnOtherDeptPolicy(t *testing.T) {
	db := makeTestDB(t)
	deptA, _ := db.CreateDepartment("Engineering", "")
	deptB, _ := db.CreateDepartment("HR", "")
	deptBPolicy, _ := db.CreatePolicy("HR Policy", "", strPtr(deptB.ID), "department")

	e := echo.New()
	h := NewPolicy(db)

	body := `{"content":"# Content","version_string":"v1.0.0","changelog":"init"}`
	c, _ := makeCtx(e, http.MethodPost, body, deptBPolicy.ID, mw.RoleDeptAdmin, strPtr(deptA.ID))

	err := h.CreateVersion(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok || he.Code != http.StatusForbidden {
		t.Errorf("expected 403 HTTPError, got %v", err)
	}
}

// TestDeptAdmin_CreateVersion_AllowedOnOwnPolicy verifies that a DeptAdmin CAN
// add a version to their own department's dept-scoped policy.
func TestDeptAdmin_CreateVersion_AllowedOnOwnPolicy(t *testing.T) {
	db := makeTestDB(t)
	dept, _ := db.CreateDepartment("Engineering", "")
	ownPolicy, _ := db.CreatePolicy("Own Policy", "", strPtr(dept.ID), "department")

	e := echo.New()
	h := NewPolicy(db)

	body := `{"content":"# Content","version_string":"v1.0.0","changelog":"init"}`
	c, rec := makeCtx(e, http.MethodPost, body, ownPolicy.ID, mw.RoleDeptAdmin, strPtr(dept.ID))

	if err := h.CreateVersion(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

// TestSuperAdmin_CreateVersion_AllowedOnOrgWidePolicy verifies that a SuperAdmin
// CAN add a version to an org-wide policy.
func TestSuperAdmin_CreateVersion_AllowedOnOrgWidePolicy(t *testing.T) {
	db := makeTestDB(t)
	orgPolicy, _ := db.CreatePolicy("Org Policy", "", nil, "organization")

	e := echo.New()
	h := NewPolicy(db)

	body := `{"content":"# Content","version_string":"v1.0.0","changelog":"init"}`
	c, rec := makeCtx(e, http.MethodPost, body, orgPolicy.ID, mw.RoleSuperAdmin, nil)

	if err := h.CreateVersion(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}
