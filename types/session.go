package types

import (
	"time"

	"github.com/google/uuid"
)

type SessionRecord struct {
	ID           uuid.UUID
	Path         string
	Msg          []byte
	RequestTime  time.Time
	Response     []byte
	ResponseTime time.Time
	Status       int
	Error        string
	DurationMs   int64
	Created_at   time.Time
}

type RequestParams struct {
	Terminal string `json:"terminalId" validate:"required"`
	Amount   int    `json:"amount" validate:"required"`
	DateTime string `json:"dateTime" validate:"required"`
	Async    *bool  `json:"async" validate:"required"`
	KeyName  string `json:"keyName" validate:"required"`
	Message  string `json:"message" validate:"required"`
}

func NewSessionFromParams(params RequestParams, path string, b []byte) (*SessionRecord, error) {
	return &SessionRecord{
		ID:          uuid.New(),
		Path:        path,
		Msg:         b,
		RequestTime: time.Now(),
		Created_at:  time.Now(),
	}, nil
}
