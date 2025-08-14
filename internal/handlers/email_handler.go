package handlers

import (
	"github.com/aarondever/url-forg/internal/services"
	"github.com/go-chi/chi/v5"
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
