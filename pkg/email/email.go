package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// Sender defines the interface for an email sender.
type Sender interface {
	Send(to, subject, body string) error
}

// LogSender is an implementation of Sender that logs emails to the console.
type LogSender struct{}

// NewLogSender creates a new LogSender.
func NewLogSender() *LogSender {
	return &LogSender{}
}

// Send logs the email details to the console instead of sending a real email.
func (s *LogSender) Send(to, subject, body string) error {
	logger.Info("Sending email",
		zap.String("to", to),
		zap.String("subject", subject),
		zap.String("body", body),
	)
	return nil
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

// SMTPSender is an implementation of Sender that sends emails via SMTP.
type SMTPSender struct {
	config SMTPConfig
}

// NewSMTPSender creates a new SMTPSender.
func NewSMTPSender(config SMTPConfig) *SMTPSender {
	return &SMTPSender{config: config}
}

// Send sends an email via SMTP.
func (s *SMTPSender) Send(to, subject, body string) error {
	// Create the message
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.config.From, to, subject, body)

	// Connect to the server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var auth smtp.Auth
	if s.config.Username != "" && s.config.Password != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	if s.config.UseTLS {
		// Use TLS connection
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         s.config.Host,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			logger.Error("Failed to connect to SMTP server", zap.Error(err))
			return err
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.config.Host)
		if err != nil {
			logger.Error("Failed to create SMTP client", zap.Error(err))
			return err
		}
		defer func() {
			if err := client.Quit(); err != nil {
				logger.Error("Failed to quit SMTP client", zap.Error(err))
			}
		}()

		if auth != nil {
			if err := client.Auth(auth); err != nil {
				logger.Error("Failed to authenticate with SMTP server", zap.Error(err))
				return err
			}
		}

		if err := client.Mail(s.config.From); err != nil {
			logger.Error("Failed to set sender", zap.Error(err))
			return err
		}

		if err := client.Rcpt(to); err != nil {
			logger.Error("Failed to set recipient", zap.Error(err))
			return err
		}

		writer, err := client.Data()
		if err != nil {
			logger.Error("Failed to get data writer", zap.Error(err))
			return err
		}

		_, err = writer.Write([]byte(msg))
		if err != nil {
			logger.Error("Failed to write message", zap.Error(err))
			return err
		}

		err = writer.Close()
		if err != nil {
			logger.Error("Failed to close writer", zap.Error(err))
			return err
		}
	} else {
		// Use plain SMTP connection
		err := smtp.SendMail(addr, auth, s.config.From, []string{to}, []byte(msg))
		if err != nil {
			logger.Error("Failed to send email", zap.Error(err))
			return err
		}
	}

	logger.Info("Email sent successfully",
		zap.String("to", to),
		zap.String("subject", subject),
	)
	return nil
}

// SendToMultiple sends an email to multiple recipients.
func (s *SMTPSender) SendToMultiple(to []string, subject, body string) error {
	for _, recipient := range to {
		if err := s.Send(recipient, subject, body); err != nil {
			logger.Error("Failed to send email to recipient",
				zap.String("recipient", recipient),
				zap.Error(err))
			// Continue sending to other recipients even if one fails
		}
	}
	return nil
}
