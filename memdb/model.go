package memdb

import (
	"time"
)

type QuestionData struct { // should be stored short-term only, the place for inactive questions is the audit log
	AnsweredAt *time.Time `json:"t"`
	AnswererID *int64     `json:"u"`
	AnswerData *string    `json:"a"`

	RelatedMessages []int `json:"m"` // so they can all be deleted at once

	SourceToken uint `json:"s"`

	Ready bool `json:"r"`
}
