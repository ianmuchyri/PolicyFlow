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

// Init creates tables and configures SQLite pragmas.
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
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS policies (
	id                 TEXT PRIMARY KEY,
	title              TEXT NOT NULL,
	current_version_id TEXT,
	status             TEXT NOT NULL DEFAULT 'Draft',
	department         TEXT NOT NULL DEFAULT '',
	created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS policy_versions (
	id             TEXT PRIMARY KEY,
	policy_id      TEXT NOT NULL,
	content        TEXT NOT NULL,
	version_string TEXT NOT NULL,
	changelog      TEXT NOT NULL DEFAULT '',
	created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (policy_id) REFERENCES policies(id)
);

CREATE TABLE IF NOT EXISTS acknowledgements (
	id                TEXT PRIMARY KEY,
	user_id           TEXT NOT NULL,
	policy_version_id TEXT NOT NULL,
	timestamp         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
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

// ─── Models ────────────────────────────────────────────────────────────────

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedBy *string   `json:"created_by,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Policy struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	CurrentVersionID *string   `json:"current_version_id,omitempty"`
	Status           string    `json:"status"`
	Department       string    `json:"department"`
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

// ─── User queries ──────────────────────────────────────────────────────────

func (db *DB) CreateUser(email, name, role string, createdBy *string) (*User, error) {
	u := &User{
		ID:        uuid.New().String(),
		Email:     email,
		Name:      name,
		Role:      role,
		CreatedBy: createdBy,
		CreatedAt: time.Now().UTC(),
	}
	_, err := db.conn.Exec(
		`INSERT INTO users (id, email, name, role, created_by, created_at) VALUES (?,?,?,?,?,?)`,
		u.ID, u.Email, u.Name, u.Role, u.CreatedBy, u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (db *DB) GetUserByID(id string) (*User, error) {
	return db.scanUser(db.conn.QueryRow(
		`SELECT id, email, name, role, created_by, created_at FROM users WHERE id = ?`, id,
	))
}

func (db *DB) GetUserByEmail(email string) (*User, error) {
	return db.scanUser(db.conn.QueryRow(
		`SELECT id, email, name, role, created_by, created_at FROM users WHERE email = ?`, email,
	))
}

func (db *DB) ListUsers() ([]*User, error) {
	rows, err := db.conn.Query(
		`SELECT id, email, name, role, created_by, created_at FROM users ORDER BY created_at ASC`,
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

type scanner interface {
	Scan(dest ...any) error
}

func (db *DB) scanUser(row scanner) (*User, error) {
	u := &User{}
	var createdBy sql.NullString
	var createdAt string
	err := row.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &createdBy, &createdAt)
	if err != nil {
		return nil, err
	}
	if createdBy.Valid {
		u.CreatedBy = &createdBy.String
	}
	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return u, nil
}

// ─── Policy queries ────────────────────────────────────────────────────────

func (db *DB) CreatePolicy(title, department string) (*Policy, error) {
	p := &Policy{
		ID:         uuid.New().String(),
		Title:      title,
		Department: department,
		Status:     "Draft",
		CreatedAt:  time.Now().UTC(),
	}
	_, err := db.conn.Exec(
		`INSERT INTO policies (id, title, department, status, created_at) VALUES (?,?,?,?,?)`,
		p.ID, p.Title, p.Department, p.Status, p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (db *DB) GetPolicy(id string) (*Policy, error) {
	return db.scanPolicy(db.conn.QueryRow(
		`SELECT id, title, current_version_id, status, department, created_at FROM policies WHERE id = ?`, id,
	))
}

func (db *DB) ListPolicies() ([]*Policy, error) {
	rows, err := db.conn.Query(
		`SELECT id, title, current_version_id, status, department, created_at FROM policies ORDER BY created_at DESC`,
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

func (db *DB) UpdatePolicy(id, title, status, department string) error {
	_, err := db.conn.Exec(
		`UPDATE policies SET title=?, status=?, department=? WHERE id=?`,
		title, status, department, id,
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
	var cvID sql.NullString
	var createdAt string
	err := row.Scan(&p.ID, &p.Title, &cvID, &p.Status, &p.Department, &createdAt)
	if err != nil {
		return nil, err
	}
	if cvID.Valid {
		p.CurrentVersionID = &cvID.String
	}
	p.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
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
		CreatedAt:     time.Now().UTC(),
	}
	_, err := db.conn.Exec(
		`INSERT INTO policy_versions (id, policy_id, content, version_string, changelog, created_at) VALUES (?,?,?,?,?,?)`,
		v.ID, v.PolicyID, v.Content, v.VersionString, v.Changelog, v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
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
	v.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
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
		a.ID, a.UserID, a.PolicyVersionID, a.Timestamp, a.SignatureHash,
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
		a.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
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
		a.Timestamp, _ = time.Parse("2006-01-02 15:04:05", ts)
		acks = append(acks, a)
	}
	return acks, rows.Err()
}

// ─── Admin stats ───────────────────────────────────────────────────────────

type Stats struct {
	TotalUsers      int `json:"total_users"`
	TotalPolicies   int `json:"total_policies"`
	PublishedCount  int `json:"published_count"`
	DraftCount      int `json:"draft_count"`
	ReviewCount     int `json:"review_count"`
	ArchivedCount   int `json:"archived_count"`
	TotalAckCount   int `json:"total_acknowledgements"`
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
