package seed

import (
	"database/sql"
	"errors"
	"log"

	"policyflow/internal/database"
)

// Run seeds the database with initial data if it hasn't been seeded yet.
// adminEmail and adminName configure the bootstrap admin account; they fall
// back to "admin@policyflow.local" / "Policy Admin" when empty.
// It is safe to call on every startup — it is idempotent.
func Run(db *database.DB, adminEmail, adminName string) error {
	if adminEmail == "" {
		adminEmail = "admin@policyflow.local"
	}
	if adminName == "" {
		adminName = "Policy Admin"
	}

	// Check if admin user already exists.
	_, err := db.GetUserByEmail(adminEmail)
	if err == nil {
		return nil // already seeded
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	log.Println("Seeding database with initial data…")

	// Create sample departments.
	hr, err := db.CreateDepartment("Human Resources", "HR policies and employee relations")
	if err != nil {
		return err
	}
	log.Printf("  Created department: %s (id=%s)", hr.Name, hr.ID)

	eng, err := db.CreateDepartment("Engineering", "Technical standards and engineering practices")
	if err != nil {
		return err
	}
	log.Printf("  Created department: %s (id=%s)", eng.Name, eng.ID)

	// Create admin user (SuperAdmin, no department).
	admin, err := db.CreateUser(adminEmail, adminName, "SuperAdmin", nil, nil)
	if err != nil {
		return err
	}
	log.Printf("  Created admin user: %s (id=%s)", admin.Email, admin.ID)

	// Create a staff test user in HR.
	staff, err := db.CreateUser("staff@policyflow.local", "Test Staff", "Staff", &admin.ID, &hr.ID)
	if err != nil {
		return err
	}
	log.Printf("  Created staff user: %s (id=%s)", staff.Email, staff.ID)

	// Create a sample org-wide policy.
	policy, err := db.CreatePolicy("Employee Code of Conduct", "Human Resources", nil, "organization")
	if err != nil {
		return err
	}
	log.Printf("  Created policy: %s (id=%s)", policy.Title, policy.ID)

	content := `# Employee Code of Conduct

## 1. Purpose

This Code of Conduct establishes the standards of professional behavior expected of all employees. It applies to all staff members regardless of their position or department.

## 2. Core Principles

- **Integrity**: Act honestly and ethically in all interactions
- **Respect**: Treat every colleague, customer, and partner with dignity
- **Accountability**: Take responsibility for your actions and decisions
- **Confidentiality**: Protect sensitive business and personal information

## 3. Professional Conduct

Employees are expected to:

- Arrive on time and fulfill their job responsibilities
- Communicate professionally in all forms of correspondence
- Avoid conflicts of interest and disclose potential conflicts to management
- Comply with all applicable laws and company policies

## 4. Workplace Respect

We are committed to a work environment free from:
- Harassment, discrimination, or bullying of any kind
- Retaliation against those who report concerns in good faith

## 5. Reporting Violations

If you observe or experience a violation of this policy, report it immediately to your manager, HR, or through the anonymous ethics hotline.

## 6. Acknowledgement

By acknowledging this policy, you confirm that you have read, understood, and agree to comply with its terms.
`
	version, err := db.CreatePolicyVersion(policy.ID, content, "v1.0.0", "Initial release")
	if err != nil {
		return err
	}
	if err := db.SetPolicyCurrentVersion(policy.ID, version.ID); err != nil {
		return err
	}
	if err := db.UpdatePolicy(policy.ID, policy.Title, "Published", policy.Department, nil, "organization"); err != nil {
		return err
	}
	log.Printf("  Created policy version %s (id=%s)", version.VersionString, version.ID)

	// Create a sample department-scoped policy for Engineering.
	engPolicy, err := db.CreatePolicy("Engineering Security Standards", "Engineering", &eng.ID, "department")
	if err != nil {
		return err
	}
	engVersion, err := db.CreatePolicyVersion(engPolicy.ID,
		"# Engineering Security Standards\n\nAll engineers must follow secure coding practices and review guidelines.",
		"v1.0.0", "Initial release")
	if err != nil {
		return err
	}
	if err := db.SetPolicyCurrentVersion(engPolicy.ID, engVersion.ID); err != nil {
		return err
	}
	if err := db.UpdatePolicy(engPolicy.ID, engPolicy.Title, "Published", engPolicy.Department, &eng.ID, "department"); err != nil {
		return err
	}
	log.Printf("  Created department policy: %s (id=%s)", engPolicy.Title, engPolicy.ID)

	log.Println("Seeding complete.")
	return nil
}
