package memdb

import (
	"strconv"
	"time"
)

type QuestionOption struct {
	Data  string `json:"d"`
	Label string `json:"l,omitempty"`
}

type StoredMessage struct {
	MessageID int   `json:"m"`
	ChatID    int64 `json:"c"`
}

func (s StoredMessage) MessageSig() (string, int64) {
	return strconv.Itoa(s.MessageID), s.ChatID
}

type QuestionData struct { // should be stored short-term only, the place for inactive questions is the audit log
	AnsweredAt *time.Time `json:"t"`
	AnswererID *int64     `json:"u"`
	AnswerData *string    `json:"a"`

	RelatedMessages []StoredMessage `json:"m"` // so they can all be deleted at once

	SourceTokenID uint             `json:"s"`
	Options       []QuestionOption `json:"o"`

	Ready bool `json:"r"`
}

func (q QuestionData) IsAnswered() bool {
	return q.AnswerData != nil && q.AnsweredAt != nil && q.AnswererID != nil && q.Ready
}
