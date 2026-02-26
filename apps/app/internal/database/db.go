package database

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// DB wraps the SQL database and provides all query methods.
type DB struct {
	conn *sql.DB
}

func New(conn *sql.DB) *DB {
	return &DB{conn: conn}
}

// Init creates base tables and configures SQLite pragmas.
func (db *DB) Init() error {
	pragmas := `
PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;
PRAGMA busy_timeout = 5000;
`
	if _, err := db.conn.Exec(pragmas); err != nil {
		return fmt.Errorf("pragmas: %w", err)
	}

	schema := `
CREATE TABLE IF NOT EXISTS users (
	id         TEXT PRIMARY KEY,
	email      TEXT UNIQUE NOT NULL,
	name       TEXT NOT NULL,
	role       TEXT NOT NULL DEFAULT 'Staff',
	created_by TEXT,
	created_at TEXT NOT NULL,
	FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS policies (
	id                 TEXT PRIMARY KEY,
	title              TEXT NOT NULL,
	current_version_id TEXT,
	status             TEXT NOT NULL DEFAULT 'Draft',
	department         TEXT NOT NULL DEFAULT '',
	created_at         TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS policy_versions (
	id             TEXT PRIMARY KEY,
	policy_id      TEXT NOT NULL,
	content        TEXT NOT NULL,
	version_string TEXT NOT NULL,
	changelog      TEXT NOT NULL DEFAULT '',
	created_at     TEXT NOT NULL,
	FOREIGN KEY (policy_id) REFERENCES policies(id)
);

CREATE TABLE IF NOT EXISTS acknowledgements (
	id                TEXT PRIMARY KEY,
	user_id           TEXT NOT NULL,
	policy_version_id TEXT NOT NULL,
	timestamp         TEXT NOT NULL,
	signature_hash    TEXT NOT NULL,
	UNIQUE(user_id, policy_version_id),
	FOREIGN KEY (user_id) REFERENCES users(id),
	FOREIGN KEY (policy_version_id) REFERENCES policy_versions(id)
);
`
	if _, err := db.conn.Exec(schema); err != nil {
		return fmt.Errorf("schema: %w", err)
	}
	return nil
}

// parseTime tries multiple formats to robustly parse timestamps stored in SQLite.
func parseTime(s string) time.Time {
	for _, format := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z"} {
		if t, err := time.Parse(format, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

// now returns the current UTC time formatted as RFC3339 for consistent SQLite storage.
func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// ─── Models ────────────────────────────────────────────────────────────────

type Department struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type User struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	CreatedBy      *string   `json:"created_by,omitempty"`
	DepartmentID   *string   `json:"department_id"`
	DepartmentName *string   `json:"department_name"`
	CreatedAt      time.Time `json:"created_at"`
}

type Policy struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	CurrentVersionID *string   `json:"current_version_id,omitempty"`
	Status           string    `json:"status"`
	Department       string    `json:"department"` // legacy text field
	DepartmentID     *string   `json:"department_id"`
	DepartmentName   *string   `json:"department_name"`
	VisibilityType   string    `json:"visibility_type"`
	CreatedAt        time.Time `json:"created_at"`
}

type PolicyVersion struct {
	ID            string    `json:"id"`
	PolicyID      string    `json:"policy_id"`
	Content       string    `json:"content"`
	VersionString string    `json:"version_string"`
	Changelog     string    `json:"changelog"`
	CreatedAt     time.Time `json:"created_at"`
}

type Acknowledgement struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	PolicyVersionID string    `json:"policy_version_id"`
	Timestamp       time.Time `json:"timestamp"`
	SignatureHash   string    `json:"signature_hash"`
}

// ─── scanner helper ────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

// ─── Department queries ────────────────────────────────────────────────────

func (db *DB) CreateDepartment(name, description string) (*Department, error) {
	d := &Department{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
	}
	ts := now()
	_, err := db.conn.Exec(
		`INSERT INTO departments (id, name, description, created_at, updated_at) VALUES (?,?,?,?,?)`,
		d.ID, d.Name, d.Description, ts, ts,
	)
	if err != nil {
		return nil, err
	}
	d.CreatedAt = parseTime(ts)
	d.UpdatedAt = parseTime(ts)
	return d, nil
}

func (db *DB) GetDepartment(id string) (*Department, error) {
	return db.scanDepartment(db.conn.QueryRow(
		`SELECT id, name, description, created_at, updated_at FROM departments WHERE id = ?`, id,
	))
}

func (db *DB) GetDepartmentByName(name string) (*Department, error) {
	return db.scanDepartment(db.conn.QueryRow(
		`SELECT id, name, description, created_at, updated_at FROM departments WHERE name = ?`, name,
	))
}

func (db *DB) ListDepartments() ([]*Department, error) {
	rows, err := db.conn.Query(
		`SELECT id, name, description, created_at, updated_at FROM departments ORDER BY name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var depts []*Department
	for rows.Next() {
		d, err := db.scanDepartment(rows)
		if err != nil {
			return nil, err
		}
		depts = append(depts, d)
	}
	return depts, rows.Err()
}

func (db *DB) UpdateDepartment(id, name, description string) (*Department, error) {
	ts := now()
	_, err := db.conn.Exec(
		`UPDATE departments SET name=?, description=?, updated_at=? WHERE id=?`,
		name, description, ts, id,
	)
	if err != nil {
		return nil, err
	}
	return db.GetDepartment(id)
}

func (db *DB) DeleteDepartment(id string) error {
	_, err := db.conn.Exec(`DELETE FROM departments WHERE id=?`, id)
	return err
}

func (db *DB) DepartmentHasPolicies(id string) (bool, error) {
	var count int
	err := db.conn.QueryRow(
		`SELECT COUNT(*) FROM policies WHERE department_id=?`, id,
	).Scan(&count)
	return count > 0, err
}

func (db *DB) scanDepartment(row scanner) (*Department, error) {
	d := &Department{}
	var createdAt, updatedAt string
	if err := row.Scan(&d.ID, &d.Name, &d.Description, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	d.CreatedAt = parseTime(createdAt)
	d.UpdatedAt = parseTime(updatedAt)
	return d, nil
}

// ─── User queries ──────────────────────────────────────────────────────────

func (db *DB) CreateUser(email, name, role string, createdBy *string, departmentID *string) (*User, error) {
	u := &User{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         name,
		Role:         role,
		CreatedBy:    createdBy,
		DepartmentID: departmentID,
	}
	ts := now()
	_, err := db.conn.Exec(
		`INSERT INTO users (id, email, name, role, created_by, department_id, created_at) VALUES (?,?,?,?,?,?,?)`,
		u.ID, u.Email, u.Name, u.Role, u.CreatedBy, u.DepartmentID, ts,
	)
	if err != nil {
		return nil, err
	}
	u.CreatedAt = parseTime(ts)
	return u, nil
}

func (db *DB) UpdateUser(id, name, email, role string, departmentID *string) error {
	_, err := db.conn.Exec(
		`UPDATE users SET name=?, email=?, role=?, department_id=? WHERE id=?`,
		name, email, role, departmentID, id,
	)
	return err
}

func (db *DB) DeleteUser(id string) error {
	_, err := db.conn.Exec(`DELETE FROM users WHERE id=?`, id)
	return err
}

func (db *DB) CountSuperAdmins() (int, error) {
	var count int
	err := db.conn.QueryRow(`SELECT COUNT(*) FROM users WHERE role='SuperAdmin'`).Scan(&count)
	return count, err
}

func (db *DB) GetUserByID(id string) (*User, error) {
	return db.scanUser(db.conn.QueryRow(
		`SELECT u.id, u.email, u.name, u.role, u.created_by, u.department_id, d.name, u.created_at
		 FROM users u LEFT JOIN departments d ON u.department_id = d.id WHERE u.id = ?`, id,
	))
}

func (db *DB) GetUserByEmail(email string) (*User, error) {
	return db.scanUser(db.conn.QueryRow(
		`SELECT u.id, u.email, u.name, u.role, u.created_by, u.department_id, d.name, u.created_at
		 FROM users u LEFT JOIN departments d ON u.department_id = d.id WHERE u.email = ?`, email,
	))
}

func (db *DB) ListUsers() ([]*User, error) {
	rows, err := db.conn.Query(
		`SELECT u.id, u.email, u.name, u.role, u.created_by, u.department_id, d.name, u.created_at
		 FROM users u LEFT JOIN departments d ON u.department_id = d.id ORDER BY u.created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := db.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (db *DB) ListUsersByDepartment(deptID string) ([]*User, error) {
	rows, err := db.conn.Query(
		`SELECT u.id, u.email, u.name, u.role, u.created_by, u.department_id, d.name, u.created_at
		 FROM users u LEFT JOIN departments d ON u.department_id = d.id
		 WHERE u.department_id = ? ORDER BY u.created_at ASC`, deptID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := db.scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (db *DB) scanUser(row scanner) (*User, error) {
	u := &User{}
	var createdBy, deptID, deptName sql.NullString
	var createdAt string
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &createdBy, &deptID, &deptName, &createdAt)
	if err != nil {
		return nil, err
	}
	if createdBy.Valid {
		u.CreatedBy = &createdBy.String
	}
	if deptID.Valid {
		u.DepartmentID = &deptID.String
	}
	if deptName.Valid {
		u.DepartmentName = &deptName.String
	}
	u.CreatedAt = parseTime(createdAt)
	return u, nil
}

// ─── Policy queries ────────────────────────────────────────────────────────

func (db *DB) CreatePolicy(title, department string, departmentID *string, visibilityType string) (*Policy, error) {
	p := &Policy{
		ID:             uuid.New().String(),
		Title:          title,
		Department:     department,
		DepartmentID:   departmentID,
		VisibilityType: visibilityType,
		Status:         "Draft",
	}
	ts := now()
	_, err := db.conn.Exec(
		`INSERT INTO policies (id, title, department, department_id, visibility_type, status, created_at) VALUES (?,?,?,?,?,?,?)`,
		p.ID, p.Title, p.Department, p.DepartmentID, p.VisibilityType, p.Status, ts,
	)
	if err != nil {
		return nil, err
	}
	p.CreatedAt = parseTime(ts)
	return p, nil
}

func (db *DB) GetPolicy(id string) (*Policy, error) {
	return db.scanPolicy(db.conn.QueryRow(
		`SELECT p.id, p.title, p.current_version_id, p.status, p.department, p.department_id, d.name, p.visibility_type, p.created_at
		 FROM policies p LEFT JOIN departments d ON p.department_id = d.id WHERE p.id = ?`, id,
	))
}

// ListPoliciesForUser returns policies visible to the given role/department.
// SuperAdmin sees all. Others see org-wide + their own department's policies.
func (db *DB) ListPoliciesForUser(role string, deptID *string) ([]*Policy, error) {
	var (
		rows *sql.Rows
		err  error
	)
	base := `SELECT p.id, p.title, p.current_version_id, p.status, p.department,
	                p.department_id, d.name, p.visibility_type, p.created_at
	         FROM policies p LEFT JOIN departments d ON p.department_id = d.id`

	if role == "SuperAdmin" {
		rows, err = db.conn.Query(base + ` ORDER BY p.created_at DESC`)
	} else if deptID != nil {
		rows, err = db.conn.Query(
			base+` WHERE p.visibility_type = 'organization'
			            OR (p.visibility_type = 'department' AND p.department_id = ?)
			       ORDER BY p.created_at DESC`,
			*deptID,
		)
	} else {
		// No department — only org-wide policies.
		rows, err = db.conn.Query(base + ` WHERE p.visibility_type = 'organization' ORDER BY p.created_at DESC`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*Policy
	for rows.Next() {
		p, err := db.scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

// ListPolicies returns all policies (admin use — no visibility filter).
func (db *DB) ListPolicies() ([]*Policy, error) {
	rows, err := db.conn.Query(
		`SELECT p.id, p.title, p.current_version_id, p.status, p.department,
		        p.department_id, d.name, p.visibility_type, p.created_at
		 FROM policies p LEFT JOIN departments d ON p.department_id = d.id ORDER BY p.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []*Policy
	for rows.Next() {
		p, err := db.scanPolicy(rows)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

func (db *DB) UpdatePolicy(id, title, status, department string, departmentID *string, visibilityType string) error {
	_, err := db.conn.Exec(
		`UPDATE policies SET title=?, status=?, department=?, department_id=?, visibility_type=? WHERE id=?`,
		title, status, department, departmentID, visibilityType, id,
	)
	return err
}

func (db *DB) SetPolicyCurrentVersion(policyID, versionID string) error {
	_, err := db.conn.Exec(
		`UPDATE policies SET current_version_id=? WHERE id=?`, versionID, policyID,
	)
	return err
}

func (db *DB) scanPolicy(row scanner) (*Policy, error) {
	p := &Policy{}
	var cvID, deptID, deptName sql.NullString
	var createdAt string
	err := row.Scan(&p.ID, &p.Title, &cvID, &p.Status, &p.Department, &deptID, &deptName, &p.VisibilityType, &createdAt)
	if err != nil {
		return nil, err
	}
	if cvID.Valid {
		p.CurrentVersionID = &cvID.String
	}
	if deptID.Valid {
		p.DepartmentID = &deptID.String
	}
	if deptName.Valid {
		p.DepartmentName = &deptName.String
	}
	p.CreatedAt = parseTime(createdAt)
	return p, nil
}

// ─── Policy version queries ────────────────────────────────────────────────

func (db *DB) CreatePolicyVersion(policyID, content, versionString, changelog string) (*PolicyVersion, error) {
	v := &PolicyVersion{
		ID:            uuid.New().String(),
		PolicyID:      policyID,
		Content:       content,
		VersionString: versionString,
		Changelog:     changelog,
	}
	ts := now()
	_, err := db.conn.Exec(
		`INSERT INTO policy_versions (id, policy_id, content, version_string, changelog, created_at) VALUES (?,?,?,?,?,?)`,
		v.ID, v.PolicyID, v.Content, v.VersionString, v.Changelog, ts,
	)
	if err != nil {
		return nil, err
	}
	v.CreatedAt = parseTime(ts)
	return v, nil
}

func (db *DB) GetPolicyVersion(id string) (*PolicyVersion, error) {
	return db.scanVersion(db.conn.QueryRow(
		`SELECT id, policy_id, content, version_string, changelog, created_at FROM policy_versions WHERE id = ?`, id,
	))
}

func (db *DB) ListPolicyVersions(policyID string) ([]*PolicyVersion, error) {
	rows, err := db.conn.Query(
		`SELECT id, policy_id, content, version_string, changelog, created_at FROM policy_versions WHERE policy_id=? ORDER BY created_at DESC`,
		policyID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []*PolicyVersion
	for rows.Next() {
		v, err := db.scanVersion(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (db *DB) scanVersion(row scanner) (*PolicyVersion, error) {
	v := &PolicyVersion{}
	var createdAt string
	err := row.Scan(&v.ID, &v.PolicyID, &v.Content, &v.VersionString, &v.Changelog, &createdAt)
	if err != nil {
		return nil, err
	}
	v.CreatedAt = parseTime(createdAt)
	return v, nil
}

// ─── Acknowledgement queries ───────────────────────────────────────────────

func (db *DB) CreateAcknowledgement(userID, policyVersionID string) (*Acknowledgement, error) {
	ts := time.Now().UTC()
	sig := fmt.Sprintf("%x", sha256.Sum256([]byte(userID+policyVersionID+ts.String())))
	a := &Acknowledgement{
		ID:              uuid.New().String(),
		UserID:          userID,
		PolicyVersionID: policyVersionID,
		Timestamp:       ts,
		SignatureHash:   sig,
	}
	_, err := db.conn.Exec(
		`INSERT INTO acknowledgements (id, user_id, policy_version_id, timestamp, signature_hash) VALUES (?,?,?,?,?)`,
		a.ID, a.UserID, a.PolicyVersionID, ts.Format(time.RFC3339), a.SignatureHash,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (db *DB) HasAcknowledged(userID, policyVersionID string) (bool, error) {
	var count int
	err := db.conn.QueryRow(
		`SELECT COUNT(*) FROM acknowledgements WHERE user_id=? AND policy_version_id=?`,
		userID, policyVersionID,
	).Scan(&count)
	return count > 0, err
}

func (db *DB) ListAcknowledgements(policyVersionID string) ([]*Acknowledgement, error) {
	rows, err := db.conn.Query(
		`SELECT id, user_id, policy_version_id, timestamp, signature_hash FROM acknowledgements WHERE policy_version_id=? ORDER BY timestamp DESC`,
		policyVersionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var acks []*Acknowledgement
	for rows.Next() {
		a := &Acknowledgement{}
		var ts string
		if err := rows.Scan(&a.ID, &a.UserID, &a.PolicyVersionID, &ts, &a.SignatureHash); err != nil {
			return nil, err
		}
		a.Timestamp = parseTime(ts)
		acks = append(acks, a)
	}
	return acks, rows.Err()
}

func (db *DB) ListUserAcknowledgements(userID string) ([]*Acknowledgement, error) {
	rows, err := db.conn.Query(
		`SELECT id, user_id, policy_version_id, timestamp, signature_hash FROM acknowledgements WHERE user_id=? ORDER BY timestamp DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var acks []*Acknowledgement
	for rows.Next() {
		a := &Acknowledgement{}
		var ts string
		if err := rows.Scan(&a.ID, &a.UserID, &a.PolicyVersionID, &ts, &a.SignatureHash); err != nil {
			return nil, err
		}
		a.Timestamp = parseTime(ts)
		acks = append(acks, a)
	}
	return acks, rows.Err()
}

// ─── Admin stats ───────────────────────────────────────────────────────────

type Stats struct {
	TotalUsers     int `json:"total_users"`
	TotalPolicies  int `json:"total_policies"`
	PublishedCount int `json:"published_count"`
	DraftCount     int `json:"draft_count"`
	ReviewCount    int `json:"review_count"`
	ArchivedCount  int `json:"archived_count"`
	TotalAckCount  int `json:"total_acknowledgements"`
}

func (db *DB) GetStats() (*Stats, error) {
	s := &Stats{}
	db.conn.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&s.TotalUsers)
	db.conn.QueryRow(`SELECT COUNT(*) FROM policies`).Scan(&s.TotalPolicies)
	db.conn.QueryRow(`SELECT COUNT(*) FROM policies WHERE status='Published'`).Scan(&s.PublishedCount)
	db.conn.QueryRow(`SELECT COUNT(*) FROM policies WHERE status='Draft'`).Scan(&s.DraftCount)
	db.conn.QueryRow(`SELECT COUNT(*) FROM policies WHERE status='Review'`).Scan(&s.ReviewCount)
	db.conn.QueryRow(`SELECT COUNT(*) FROM policies WHERE status='Archived'`).Scan(&s.ArchivedCount)
	db.conn.QueryRow(`SELECT COUNT(*) FROM acknowledgements`).Scan(&s.TotalAckCount)
	return s, nil
}

// AckStatusForUser returns a map of policy_version_id → bool for all acknowledgements by a user.
func (db *DB) AckStatusForUser(userID string) (map[string]bool, error) {
	rows, err := db.conn.Query(
		`SELECT policy_version_id FROM acknowledgements WHERE user_id=?`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]bool{}
	for rows.Next() {
		var vid string
		if err := rows.Scan(&vid); err != nil {
			return nil, err
		}
		result[vid] = true
	}
	return result, rows.Err()
}
