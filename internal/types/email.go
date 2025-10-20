package types

import (
	"context"

	"github.com/aarondever/notiflow/internal/models"
)

type EmailService interface {
	SendEmail(ctx context.Context, email *models.Email) (*models.Email, error)
}
