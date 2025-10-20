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
	"go.mongodb.org/mongo-driver/v2/bson"
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

func (service *EmailService) SendEmail(ctx context.Context, params *models.SendEmailRequest) (*models.Email, error) {
	if service.smtpServerCount == 0 {
		return nil, fmt.Errorf("no SMTP servers configured")
	}

	// Create email record
	email := models.Email{
		To:          params.To,
		CC:          params.CC,
		BCC:         params.BCC,
		Subject:     params.Subject,
		Body:        params.Body,
		IsHTML:      params.IsHTML,
		Attachments: params.Attachments,
	}

	// Save to database
	dbEmail, err := service.db.CreateEmail(ctx, email)
	if err != nil {
		slog.Error("Failed to create email", "error", err)
		return nil, err
	}

	// Send email asynchronously
	go service.sendEmailAsync(dbEmail.ID, params)

	service.smtpServerUsageCount++

	return dbEmail, nil
}

func (service *EmailService) sendEmailAsync(emailID bson.ObjectID, params *models.SendEmailRequest) {
	ctx := context.Background()

	smtpServer := service.cfg.SMTPServers[service.smtpServerUsageCount%service.smtpServerCount]
	slog.Info("Using SMTP server sending email", "username", smtpServer.Username, "host")

	// Create message
	message := gomail.NewMessage()
	message.SetHeader("From", smtpServer.FromEmail)
	message.SetHeader("To", params.To...)

	if len(params.CC) > 0 {
		message.SetHeader("Cc", params.CC...)
	}
	if len(params.BCC) > 0 {
		message.SetHeader("Bcc", params.BCC...)
	}

	message.SetHeader("Subject", params.Subject)

	if params.IsHTML {
		message.SetBody("text/html", params.Body)
	} else {
		message.SetBody("text/plain", params.Body)
	}

	// Add attachments
	for _, attachment := range params.Attachments {
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
		_, err = service.db.UpdateEmailFail(ctx, models.Email{
			ID:       emailID,
			ErrorMsg: err.Error(),
		})
		if err != nil {
			slog.Error("Failed to update email", "error", err)
		}

		return
	}

	// Update status to sent
	_, err := service.db.UpdateEmailSent(ctx, models.Email{
		ID:     emailID,
		SentAt: time.Now(),
	})
	if err != nil {
		slog.Error("Failed to update email", "error", err)
	}
}
