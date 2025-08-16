package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aarondever/url-forg/internal/config"
	"github.com/aarondever/url-forg/internal/database"
	"github.com/aarondever/url-forg/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"gopkg.in/gomail.v2"
	"io"
	"log/slog"
	"time"
)

const EmailCollectionName = "emails"

type EmailService struct {
	db  *database.Database
	cfg *config.Config
}

func NewEmailService(db *database.Database, cfg *config.Config) *EmailService {
	return &EmailService{
		db:  db,
		cfg: cfg,
	}
}

func (service *EmailService) SendEmail(ctx context.Context, params *models.SendEmailRequest) (*models.Email, error) {
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
		return nil, fmt.Errorf("failed to save email: %w", err)
	}

	// Send email asynchronously
	go service.sendEmailAsync(dbEmail.ID, params)

	return dbEmail, nil
}

func (service *EmailService) sendEmailAsync(emailID bson.ObjectID, params *models.SendEmailRequest) {
	ctx := context.Background()

	// Create message
	message := gomail.NewMessage()
	message.SetHeader("From", service.cfg.FromEmail)
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
		service.cfg.SMTPHost,
		service.cfg.SMTPPort,
		service.cfg.SMTPUsername,
		service.cfg.SMTPPassword)

	// Send email
	if err := dialer.DialAndSend(message); err != nil {
		// Update status to failed
		_, err = service.db.UpdateEmail(ctx, models.Email{
			ID:       emailID,
			Status:   models.StatusFailed,
			ErrorMsg: err.Error(),
		})
		if err != nil {
			slog.Error("Failed to update email", "error", err)
		}
		return
	}

	// Update status to sent
	_, err := service.db.UpdateEmail(ctx, models.Email{
		ID:       emailID,
		Status:   models.StatusSent,
		ErrorMsg: "",
		SentAt:   time.Now(),
	})
	if err != nil {
		slog.Error("Failed to update email", "error", err)
	}
}
