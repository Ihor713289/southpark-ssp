package message

import (
	"api/domain"
	"api/internal/common"
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type MessageHandler struct {
	MessageService MessageService
}

type MessageService interface {
	Store(ctx context.Context, message domain.Message) (string, error)
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

	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		common.WriteJSON(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	store, err := handler.MessageService.Store(r.Context(), message)
	if err != nil {
		common.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.WriteJSON(w, http.StatusCreated, store)
}
