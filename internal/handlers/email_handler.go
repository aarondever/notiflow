package handlers

import (
	"github.com/aarondever/url-forg/internal/models"
	"github.com/aarondever/url-forg/internal/services"
	"github.com/aarondever/url-forg/internal/utils"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type EmailHandler struct {
	emailService *services.EmailService
}

func NewEmailHandler(emailService *services.EmailService) *EmailHandler {
	return &EmailHandler{emailService: emailService}
}

func (handler *EmailHandler) RegisterRoutes(router *chi.Mux) {
	router.Route("/api/v1/email", func(router chi.Router) {
	})
}

func (handler *EmailHandler) SendEmail(responseWriter http.ResponseWriter, request *http.Request) {
	var params models.SendEmailRequest
	if err := utils.DecodeRequestBody(request, &params); err != nil {
		utils.RespondWithError(responseWriter, http.StatusBadRequest, err.Error())
		return
	}

	email, err := handler.emailService.SendEmail(request.Context(), &params)
	if err != nil {
		utils.RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(responseWriter, http.StatusCreated, models.EmailResponse{
		ID:        email.ID.Hex(),
		Status:    models.StatusPending,
		Message:   "Email queued for sending",
		CreatedAt: email.CreatedAt,
	})
}
