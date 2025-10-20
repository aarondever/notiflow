package handlers

import (
	"context"

	"github.com/aarondever/notiflow/internal/models"
	"github.com/aarondever/notiflow/internal/services"
	"github.com/aarondever/notiflow/internal/types"
	pb "github.com/aarondever/notiflow/proto/email"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EmailGRPCHandler struct {
	emailService types.EmailService
	pb.UnimplementedEmailServiceServer
}

func NewEmailGRPCHandler(emailService *services.EmailService) *EmailGRPCHandler {
	return &EmailGRPCHandler{
		emailService: emailService,
	}
}

func (h *EmailGRPCHandler) SendEmail(ctx context.Context, request *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	// Convert proto request to internal model
	attachments := make([]models.Attachment, len(request.Attachments))
	for i, att := range request.Attachments {
		attachments[i] = models.Attachment{
			Filename:    att.Filename,
			Content:     att.Content,
			ContentType: att.ContentType,
		}
	}

	email, err := h.emailService.SendEmail(ctx, &models.Email{
		To:          request.To,
		CC:          request.Cc,
		BCC:         request.Bcc,
		Subject:     request.Subject,
		Body:        request.Body,
		IsHTML:      request.IsHtml,
		Attachments: attachments,
	})
	if err != nil {
		return nil, err
	}

	return &pb.SendEmailResponse{
		Id:        email.ID.Hex(),
		Status:    string(models.StatusPending),
		Message:   "Email queued for sending",
		CreatedAt: timestamppb.New(email.CreatedAt),
	}, nil
}
