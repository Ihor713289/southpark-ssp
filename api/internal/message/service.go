package message

import (
	"api/domain"
	"context"
)

type MessageRepository interface {
	Store(ctx context.Context, message domain.Message) (string, error)
}

type DefaultMessageService struct {
	repository MessageRepository
}

func NewDefaultMessageService(repository MessageRepository) *DefaultMessageService {
	return &DefaultMessageService{repository}
}

func (m DefaultMessageService) Store(ctx context.Context, message domain.Message) (string, error) {
	if err := message.Validate(); err != nil {
		return "", err
	}
	return m.repository.Store(ctx, message)
}
