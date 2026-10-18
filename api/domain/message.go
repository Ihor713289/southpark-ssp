package domain

type Message struct {
	Body   string `json:"body"`
	Author string `json:"author"`
}
