package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/smtp"
	"path/filepath"

	"exchange/config"
)

// EmailService defines the interface for sending transactional emails.
type EmailService interface {
	SendEmail(ctx context.Context, to, subject, templateName string, templateData interface{}) error
}

type smtpEmailService struct {
	cfg       *config.Config
	templates *template.Template
}

// NewEmailService creates a new EmailService using SMTP.
func NewEmailService(cfg *config.Config) (EmailService, error) {
	// Parse all HTML templates in the templates directory
	tmpl, err := template.ParseGlob(filepath.Join("internal", "services", "templates", "*.html"))
	if err != nil {
		// Log the error but don't fail completely if templates aren't there yet during dev,
		// though usually we'd want to fail fast. For safety:
		return nil, fmt.Errorf("failed to parse email templates: %w", err)
	}

	return &smtpEmailService{
		cfg:       cfg,
		templates: tmpl,
	}, nil
}

// SendEmail sends an email to the specified recipient using an HTML template.
func (s *smtpEmailService) SendEmail(ctx context.Context, to, subject, templateName string, templateData interface{}) error {
	if s.cfg.SMTPHost == "" || s.cfg.SMTPPort == 0 || s.cfg.SMTPUsername == "" {
		// SMTP not configured, just log or return (useful for local dev without email)
		fmt.Printf("MOCK EMAIL SENT TO: %s, SUBJECT: %s, TEMPLATE: %s\n", to, subject, templateName)
		return nil
	}

	// 1. Execute the template
	var body bytes.Buffer
	if err := s.templates.ExecuteTemplate(&body, templateName, templateData); err != nil {
		return fmt.Errorf("executing email template '%s': %w", templateName, err)
	}

	// 2. Build the email headers and body
	// MIME headers for HTML email
	headers := make(map[string]string)
	headers["From"] = s.cfg.SMTPFrom
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body.String()

	// 3. Connect to SMTP and send
	auth := smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	// Note: In production, consider using a goroutine or a job queue to send emails asynchronously
	// so it doesn't block the HTTP request. We will run this synchronously here, but the caller
	// can wrap this in a goroutine.
	err := smtp.SendMail(addr, auth, s.cfg.SMTPFrom, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("sending email to %s: %w", to, err)
	}

	return nil
}
