package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

// Mailer sends emails via SMTP or logs them if SMTP is not configured.
type Mailer struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func New() *Mailer {
	port := 587
	if p := os.Getenv("SMTP_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	from := os.Getenv("SMTP_USER")
	if f := os.Getenv("SMTP_FROM"); f != "" {
		from = f
	}
	if from == "" {
		from = "policyflow@localhost"
	}
	return &Mailer{
		host:     os.Getenv("SMTP_HOST"),
		port:     port,
		username: os.Getenv("SMTP_USER"),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     from,
	}
}

func (m *Mailer) SendMagicLink(toEmail, toName, magicURL string) error {
	subject := "PolicyFlow — Your login link"
	body := fmt.Sprintf(`Hi %s,

Click the link below to log in to PolicyFlow. This link is valid for 24 hours.

%s

If you did not request this, you can safely ignore this email.

— The PolicyFlow Team
`, toName, magicURL)

	return m.send(toEmail, subject, body)
}

func (m *Mailer) SendNewUserWelcome(toEmail, toName, magicURL string) error {
	subject := "Welcome to PolicyFlow"
	body := fmt.Sprintf(`Hi %s,

An account has been created for you on PolicyFlow, your company's policy management system.

Click the link below to log in for the first time. This link is valid for 24 hours.

%s

After logging in, you can view and acknowledge company policies.

— The PolicyFlow Team
`, toName, magicURL)

	return m.send(toEmail, subject, body)
}

func (m *Mailer) send(to, subject, body string) error {
	if m.host == "" {
		// Development fallback: log the email
		log.Printf("📧 EMAIL (no SMTP configured)\nTo: %s\nSubject: %s\nBody:\n%s", to, subject, body)
		return nil
	}

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	msg := strings.Join([]string{
		fmt.Sprintf("From: PolicyFlow <%s>", m.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if m.username != "" && m.password != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}

	if err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
