package eventsourcing

import (
	"fmt"
	"gopkg.in/guregu/null.v4"
)

type AuthEvent struct {
	UserId   int64       `json:"user_id"`
	Verified bool        `json:"verified"`
	Guest    bool        `json:"guest"`
	Username null.String `json:"username"`
	Email    null.String `json:"email"`
	BaseChangeEvent
}

func (l AuthEvent) GetPublishKey() string {
	return fmt.Sprintf("%v", l.UserId)
}
