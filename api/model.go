package api

import (
	"github.com/marcsello/marcsellocorp-bot/db"
	"time"
)

// Notify

type NotifyRequest struct {
	Text    string `json:"text"`
	Channel string `json:"channel"`
}

type NotifyResponse struct {
	DeliveredToAnyone bool `json:"delivered_to_anyone"`
}

// Question

type UserRepr struct {
	ID        int64  `json:"id"` // This must be a signed int, because telegram assign negative id to groups
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoUrl  string `json:"photo_url"`
}

func UserToUserRepr(u db.User) UserRepr {
	return UserRepr{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Username:  u.Username,
		PhotoUrl:  u.PhotoUrl,
	}
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

type QuestionAnswer struct { // part of QuestionResponse
	Data       string    `json:"data"`
	AnsweredAt time.Time `json:"at"`
	AnsweredBy UserRepr  `json:"by"`
}

type QuestionResponse struct {
	ID string `json:"id"` // RandomID stored in the db

	Answer *QuestionAnswer `json:"answer"`
}
