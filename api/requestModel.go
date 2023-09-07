package api

import (
	"github.com/marcsello/marcsellocorp-bot/db"
	"time"
)

type NotifyRequest struct {
	Text    string `json:"text"`
	Channel string `json:"channel"`
}

type NotifyResponse struct {
	DeliveredToAnyone bool `json:"delivered_to_anyone"`
}

type QuestionOption struct {
	Data  string `json:"data"`  // Data will be in the answer
	Label string `json:"label"` // Label will be on the button
}

type QuestionRequest struct {
	Text    string `json:"text"`
	Channel string `json:"channel"`

	Options []QuestionOption `json:"options"`
}

type QuestionAnswer struct {
	Data       string    `json:"data"`
	AnsweredAt time.Time `json:"at"`
	AnsweredBy db.User   `json:"by"`
}

type QuestionResponse struct {
	ID string `json:"id"` // RandomID stored in the db

	Answer *QuestionAnswer `json:"answer"`
}
