package domain

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidMessage = errors.New("invalid message")

type Message struct {
	Body   string `json:"body"`
	Author string `json:"author"`
}

func (m Message) Validate() error {
	var missing []string
	if strings.TrimSpace(m.Author) == "" {
		missing = append(missing, "author")
	}
	if strings.TrimSpace(m.Body) == "" {
		missing = append(missing, "body")
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: missing required fields: %s", ErrInvalidMessage, strings.Join(missing, ", "))
	}
	return nil
}
