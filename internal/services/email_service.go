package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/aarondever/notiflow/internal/config"
	"github.com/aarondever/notiflow/internal/database"
	"github.com/aarondever/notiflow/internal/models"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	db                   *database.Database
	cfg                  *config.Config
	smtpServerCount      int
	smtpServerUsageCount int
}

func NewEmailService(db *database.Database, cfg *config.Config) *EmailService {
	return &EmailService{
		db:                   db,
		cfg:                  cfg,
		smtpServerCount:      len(cfg.SMTPServers),
		smtpServerUsageCount: 0,
	}
}

func (s *EmailService) SendEmail(ctx context.Context, email *models.Email) (*models.Email, error) {
	if s.smtpServerCount == 0 {
		return nil, fmt.Errorf("no SMTP servers configured")
	}

	// Save to database
	dbEmail, err := s.db.CreateEmail(ctx, email)
	if err != nil {
		slog.Error("Failed to create email", "error", err)
		return nil, err
	}

	// Send email asynchronously
	go s.sendEmailAsync(dbEmail)

	s.smtpServerUsageCount++

	return dbEmail, nil
}

func (s *EmailService) sendEmailAsync(email *models.Email) {
	ctx := context.Background()

	smtpServer := s.cfg.SMTPServers[s.smtpServerUsageCount%s.smtpServerCount]
	slog.Info("Using SMTP server sending email", "username", smtpServer.Username, "host")

	// Create message
	message := gomail.NewMessage()
	message.SetHeader("From", smtpServer.FromEmail)
	message.SetHeader("To", email.To...)

	if len(email.CC) > 0 {
		message.SetHeader("Cc", email.CC...)
	}
	if len(email.BCC) > 0 {
		message.SetHeader("Bcc", email.BCC...)
	}

	message.SetHeader("Subject", email.Subject)

	if email.IsHTML {
		message.SetBody("text/html", email.Body)
	} else {
		message.SetBody("text/plain", email.Body)
	}

	// Add attachments
	for _, attachment := range email.Attachments {
		reader := bytes.NewReader(attachment.Content)
		message.Attach(attachment.Filename, gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := io.Copy(w, reader)
			return err
		}))
	}

	// Configure SMTP
	dialer := gomail.NewDialer(
		smtpServer.Host,
		smtpServer.Port,
		smtpServer.Username,
		smtpServer.Password,
	)

	// Send email
	if err := dialer.DialAndSend(message); err != nil {
		slog.Error("Failed to send email", "error", err)

		// Update status to failed
		_, err = s.db.UpdateEmailFail(ctx, &models.Email{
			ID:       email.ID,
			ErrorMsg: err.Error(),
		})
		if err != nil {
			slog.Error("Failed to update email", "error", err)
		}

		return
	}

	// Update status to sent
	_, err := s.db.UpdateEmailSent(ctx, &models.Email{
		ID:     email.ID,
		SentAt: time.Now(),
	})
	if err != nil {
		slog.Error("Failed to update email", "error", err)
	}
}
