package utils

import (
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
)

func RespondWithJSON(responseWriter http.ResponseWriter, statusCode int, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON response: %v", payload)
		RespondWithError(responseWriter, http.StatusInternalServerError, err.Error())
		return
	}

	responseWriter.Header().Add("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)
	responseWriter.Write(data)
}

func RespondWithError(responseWriter http.ResponseWriter, statusCode int, err string) {
	log.Printf("Responding with %d error: %s", statusCode, err)

	type errorResponse struct {
		Error string `json:"error"`
	}

	RespondWithJSON(responseWriter, statusCode, errorResponse{err})
}

func DecodeRequestBody(request *http.Request, params any) error {
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(params); err != nil {
		log.Printf("Error decoding request body: %v", err)
		return err
	}

	var validate = validator.New()
	if err := validate.Struct(params); err != nil {
		log.Printf("Validation error: %v", err)
		return err
	}

	return nil
}
