package email

import (
	"crypto/tls"
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
	devMode  bool
	useTLS   bool // true = implicit TLS (port 465); false = STARTTLS (port 587)
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
		devMode:  os.Getenv("DEV_EMAIL_MODE") == "true",
		useTLS:   os.Getenv("SMTP_TLS") == "true",
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
	if m.devMode || m.host == "" {
		log.Printf("📧 EMAIL (dev mode — not sent)\nTo: %s\nSubject: %s\nBody:\n%s", to, subject, body)
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

	if m.useTLS {
		return m.sendImplicitTLS(addr, auth, to, msg)
	}
	return m.sendSTARTTLS(addr, auth, to, msg)
}

// sendSTARTTLS uses the standard smtp.SendMail which negotiates STARTTLS (port 587).
func (m *Mailer) sendSTARTTLS(addr string, auth smtp.Auth, to, msg string) error {
	log.Printf("SMTP: connecting to %s (STARTTLS)…", addr)
	if err := smtp.SendMail(addr, auth, m.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("smtp send (STARTTLS): %w", err)
	}
	log.Printf("SMTP: sent to %s", to)
	return nil
}

// sendImplicitTLS connects with immediate TLS (port 465).
func (m *Mailer) sendImplicitTLS(addr string, auth smtp.Auth, to, msg string) error {
	log.Printf("SMTP: connecting to %s (implicit TLS)…", addr)
	tlsConfig := &tls.Config{ServerName: m.host}
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Quit()

	if auth != nil {
		log.Printf("SMTP: authenticating as %s…", m.username)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(m.from); err != nil {
		return fmt.Errorf("smtp MAIL FROM: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT TO: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}
	if _, err := fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp close writer: %w", err)
	}
	log.Printf("SMTP: sent to %s", to)
	return nil
}
