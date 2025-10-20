package handlers

import (
	"net/http"

	"github.com/aarondever/notiflow/internal/models"
	"github.com/aarondever/notiflow/internal/services"
	"github.com/aarondever/notiflow/internal/types"
	"github.com/gin-gonic/gin"
)

type EmailHandler struct {
	emailService types.EmailService
}

func NewEmailHandler(emailService *services.EmailService) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

func (h *EmailHandler) RegisterRouter(router *gin.Engine) {
	emailV1 := router.Group("/api/v1/email")
	{
		emailV1.POST("/", h.SendEmail)
	}
}

func (h *EmailHandler) SendEmail(c *gin.Context) {
	var params models.SendEmailRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email, err := h.emailService.SendEmail(c.Request.Context(), &models.Email{
		To:          params.To,
		CC:          params.CC,
		BCC:         params.BCC,
		Subject:     params.Subject,
		Body:        params.Body,
		IsHTML:      params.IsHTML,
		Attachments: params.Attachments,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.EmailResponse{
		ID:        email.ID.Hex(),
		Status:    models.StatusPending,
		Message:   "Email queued for sending",
		CreatedAt: email.CreatedAt,
	})
}
