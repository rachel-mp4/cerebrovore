package types

import (
	"time"
)

type Appeal struct {
	Ban
	Post
}

type Ban struct {
	Id        int
	Username  string
	PostId    *uint32
	Reason    *string
	Comment   *string
	Response  *string
	BannedAt  time.Time
	Until     time.Time
	Moderator string
	Repealed  *bool
}
