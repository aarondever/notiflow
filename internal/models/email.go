package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type EmailStatus string

const (
	StatusPending EmailStatus = "pending"
	StatusSent    EmailStatus = "sent"
	StatusFailed  EmailStatus = "failed"
)

type Email struct {
	ID          bson.ObjectID `json:"id" bson:"_id,omitempty"`
	To          []string      `json:"to" bson:"to"`
	CC          []string      `json:"cc,omitempty" bson:"cc,omitempty"`
	BCC         []string      `json:"bcc,omitempty" bson:"bcc,omitempty"`
	Subject     string        `json:"subject" bson:"subject"`
	Body        string        `json:"body" bson:"body"`
	IsHTML      bool          `json:"is_html" bson:"is_html"`
	Status      EmailStatus   `json:"status" bson:"status"`
	ErrorMsg    string        `json:"error_message,omitempty" bson:"error_message,omitempty"`
	CreatedAt   time.Time     `json:"created_at" bson:"created_at"`
	SentAt      time.Time     `json:"sent_at,omitempty" bson:"sent_at,omitempty"`
	Attachments []Attachment  `json:"attachments,omitempty" bson:"attachments,omitempty"`
}

type Attachment struct {
	Filename    string `json:"filename" bson:"filename"`
	Content     []byte `json:"content" bson:"content"`
	ContentType string `json:"content_type" bson:"content_type"`
}

type SendEmailRequest struct {
	To          []string     `json:"to" bind:"required"`
	CC          []string     `json:"cc,omitempty"`
	BCC         []string     `json:"bcc,omitempty"`
	Subject     string       `json:"subject" bind:"required"`
	Body        string       `json:"body" bind:"required"`
	IsHTML      bool         `json:"is_html"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type EmailResponse struct {
	ID        string      `json:"id"`
	Status    EmailStatus `json:"status"`
	Message   string      `json:"message"`
	CreatedAt time.Time   `json:"created_at"`
}
