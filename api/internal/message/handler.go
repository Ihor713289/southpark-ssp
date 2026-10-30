package message

import (
	"api/domain"
	"api/internal/common"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const maxBodyBytes = 1 << 20

type MessageHandler struct {
	MessageService MessageService
}

type MessageService interface {
	Store(ctx context.Context, message domain.Message) (string, error)
}

type errorResponse struct {
	Error string `json:"error"`
}

type statusResponse struct {
	Status string `json:"status"`
}

func NewMessageHandler(messageService MessageService) *MessageHandler {
	return &MessageHandler{messageService}
}

func AddMessageHandlers(router *mux.Router, messageService MessageService) {
	var handler = NewMessageHandler(messageService)

	router.HandleFunc("/messages", handler.storeMessage).
		Methods("POST").
		Name("storeMessage")
}

func (handler *MessageHandler) storeMessage(w http.ResponseWriter, r *http.Request) {
	var message domain.Message

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&message); err != nil {
		common.WriteJSON(w, http.StatusBadRequest, errorResponse{"invalid JSON: " + err.Error()})
		return
	}

	store, err := handler.MessageService.Store(r.Context(), message)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidMessage) {
			common.WriteJSON(w, http.StatusUnprocessableEntity, errorResponse{err.Error()})
			return
		}
		log.Printf("failed to store message: %v", err)
		common.WriteJSON(w, http.StatusInternalServerError, errorResponse{"failed to publish message"})
		return
	}
	common.WriteJSON(w, http.StatusCreated, statusResponse{store})
}
