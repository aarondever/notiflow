package utils

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
)

// RespondWithJSON marshals the given payload to JSON and writes it to the response writer
func RespondWithJSON(responseWriter http.ResponseWriter, statusCode int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal JSON response", "error", err, "payload", payload)
		RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	responseWriter.Header().Add("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	responseWriter.Write(data)
}

// RespondWithError sends a standardized error response in JSON format
func RespondWithError(responseWriter http.ResponseWriter, statusCode int, err string) {
	slog.Error("Responding with error", "error", err, "status_code", statusCode)

	type errorResponse struct {
		Error string `json:"error"`
	}

	RespondWithJSON(responseWriter, statusCode, errorResponse{err})
}

func DecodeRequestBody(request *http.Request, params any) error {
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(params); err != nil {
		slog.Error("Error decoding request body", "error", err)
		return err
	}

	var validate = validator.New()
	if err := validate.Struct(params); err != nil {
		slog.Error("Validation error", "error", err)
		return err
	}

	return nil
}
