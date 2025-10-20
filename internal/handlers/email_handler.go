package handlers

import (
	"net/http"

	"github.com/aarondever/notiflow/internal/models"
	"github.com/aarondever/notiflow/internal/services"
	"github.com/gin-gonic/gin"
)

type EmailHandler struct {
	emailService *services.EmailService
}

func NewEmailHandler(emailService *services.EmailService) *EmailHandler {
	return &EmailHandler{emailService: emailService}
}

func (handler *EmailHandler) SetupRouters(router *gin.Engine) {
	emailV1 := router.Group("/api/v1/email")
	{
		emailV1.POST("/", handler.SendEmail)
	}
}

func (handler *EmailHandler) SendEmail(c *gin.Context) {
	var params models.SendEmailRequest
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email, err := handler.emailService.SendEmail(c.Request.Context(), &params)
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
