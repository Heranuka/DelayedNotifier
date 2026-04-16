// Package mailer provides implementations for sending notifications via email
package mailer

import (
	"context"
	"crypto/tls"
	"delay/internal/config"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/wb-go/wbf/logger"
	"go.uber.org/zap"
)

const htmlBody = `
<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Notification</title>
</head>
<body style="margin:0; padding:0; font-family: Arial, sans-serif; background-color:#f6f6f6; color:#222;">
  <div style="max-width:600px; margin:0 auto; padding:24px;">
    <div style="background:#ffffff; border-radius:8px; padding:24px; border:1px solid #e5e5e5;">
      <h2 style="margin:0 0 16px; font-size:20px;">Новое уведомление</h2>
      <p style="margin:0 0 12px; line-height:1.5;">{{.Message}}</p>
      <p style="margin:0; line-height:1.5; color:#666;">Кому: {{.Destination}}</p>
    </div>
  </div>
</body>
</html>`

// SMTPMailer sends emails via an SMTP server.
type SMTPMailer struct {
	logger *logger.ZerologAdapter
	cfg    *config.SMTP
}

// NewSMTPMailer constructs an SMTPMailer using the provided SMTP config.
func NewSMTPMailer(cfg *config.SMTP, logger *logger.ZerologAdapter) *SMTPMailer {
	return &SMTPMailer{logger: logger, cfg: cfg}
}

// Send sends an HTML email to the recipient address.
//
// It applies a hard 30s network deadline by setting net.Conn deadlines, so
// stuck network I/O will time out instead of blocking indefinitely.
// The context is used for dialing and for early cancellation.
func (s *SMTPMailer) Send(ctx context.Context, message, destination string) error {
	const opTimeout = 30 * time.Second

	// Derive a deadline from context, but cap it to opTimeout.
	deadline := time.Now().Add(opTimeout)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	dialer := &net.Dialer{Timeout: time.Until(deadline)}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		s.logger.Error("smtp dial failed",
			zap.String("to_masked", maskEmail(destination)),
			zap.String("smtp_host", s.cfg.Host),
			zap.Int("smtp_port", s.cfg.Port),
			zap.Error(err),
		)
		return err
	}
	defer func() { _ = conn.Close() }()

	// Hard deadline for the whole SMTP exchange (read+write).
	_ = conn.SetDeadline(deadline)

	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		s.logger.Error("smtp client init failed",
			zap.String("to_masked", maskEmail(destination)),
			zap.Error(err),
		)
		return err
	}
	defer func() { _ = c.Quit() }()

	if s.cfg.UseTLS {
		tlsCfg := &tls.Config{
			ServerName: s.cfg.Host,
			MinVersion: tls.VersionTLS12,
		}
		if err := c.StartTLS(tlsCfg); err != nil {
			s.logger.Error("smtp starttls failed",
				zap.String("to_masked", maskEmail(destination)),
				zap.Error(err),
			)
			return err
		}
	}

	// Auth (only if server supports it; PlainAuth requires the correct host).
	if s.cfg.Password != "" {
		auth := smtp.PlainAuth("", s.cfg.Email, s.cfg.Password, s.cfg.Host)
		if err := c.Auth(auth); err != nil {
			s.logger.Error("smtp auth failed",
				zap.String("to_masked", maskEmail(destination)),
				zap.Error(err),
			)
			return err
		}
	}

	if err := c.Mail(s.cfg.Email); err != nil {
		s.logger.Error("smtp MAIL FROM failed",
			zap.String("to_masked", maskEmail(destination)),
			zap.Error(err),
		)
		return err
	}

	if err := c.Rcpt(destination); err != nil {
		s.logger.Error("smtp RCPT destination failed",
			zap.String("to_masked", maskEmail(destination)),
			zap.Error(err),
		)
		return err
	}

	wc, err := c.Data()
	if err != nil {
		s.logger.Error("smtp DATA failed",
			zap.String("to_masked", maskEmail(destination)),
			zap.Error(err),
		)
		return err
	}

	msg := buildHTMLMessage(s.cfg.Email, destination, message, htmlBody)

	if _, err := wc.Write(msg); err != nil {
		_ = wc.Close()
		s.logger.Error("smtp message write failed",
			zap.String("to_masked", maskEmail(destination)),
			zap.Error(err),
		)
		return err
	}

	if err := wc.Close(); err != nil {
		s.logger.Error("smtp DATA close failed",
			zap.String("to_masked", maskEmail(destination)),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("smtp email sent",
		zap.String("to_masked", maskEmail(destination)),
	)
	return nil
}

// MaskEmail returns a partially masked version of the email address for logs.
func maskEmail(email string) string {
	at := strings.IndexByte(email, '@')
	if at <= 0 {
		if len(email) < 3 {
			return "***"
		}
		return email[:3] + "***"
	}

	local := email[:at]
	domain := email[at+1:]

	if len(local) < 2 {
		return "***@" + domain
	}

	return local[:2] + "***@" + domain
}

// buildHTMLMessage builds a minimal RFC822 message with an HTML body.
// If you need non-ASCII subjects, implement RFC 2047 encoded-word.
func buildHTMLMessage(from, to, subject, htmlTemplate string) []byte {
	body := strings.ReplaceAll(htmlTemplate, "{{.Message}}", subject)
	body = strings.ReplaceAll(body, "{{.Destination}}", to)

	return []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n%s",
		from, to, subject, body,
	))
}
