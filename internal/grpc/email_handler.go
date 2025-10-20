package grpc

import (
	"context"

	"github.com/aarondever/notiflow/internal/models"
	"github.com/aarondever/notiflow/internal/services"
	pb "github.com/aarondever/notiflow/proto/email"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EmailGRPCHandler struct {
	pb.UnimplementedEmailServiceServer
	emailService *services.EmailService
}

func NewEmailGRPCHandler(emailService *services.EmailService) *EmailGRPCHandler {
	return &EmailGRPCHandler{
		emailService: emailService,
	}
}

func (h *EmailGRPCHandler) SendEmail(ctx context.Context, req *pb.SendEmailRequest) (*pb.SendEmailResponse, error) {
	// Convert proto request to internal model
	attachments := make([]models.Attachment, len(req.Attachments))
	for i, att := range req.Attachments {
		attachments[i] = models.Attachment{
			Filename:    att.Filename,
			Content:     att.Content,
			ContentType: att.ContentType,
		}
	}

	params := &models.SendEmailRequest{
		To:          req.To,
		CC:          req.Cc,
		BCC:         req.Bcc,
		Subject:     req.Subject,
		Body:        req.Body,
		IsHTML:      req.IsHtml,
		Attachments: attachments,
	}

	email, err := h.emailService.SendEmail(ctx, params)
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
