package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/rachel-mp4/cerebrovore/utils"
)

type Post struct {
	ID           uint32
	ThreadID     uint32
	Username     string
	Anon         bool
	Nick         *string
	Color        *uint32
	PostedAt     time.Time
	TextContent  *TextContent
	ImageContent *ImageContent
	Backlinks    []uint32
}

func (p *Post) String() string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf(`
		ID: %d (%s)
		ThreadID: %d (%s)
		Username: %s
		Anon: %t
		Nick: %s
		Color: %s
		PostedAt: %s
		Backlinks: %s
		`,
		p.ID, utils.IDToA(p.ID),
		p.ThreadID, utils.IDToA(p.ThreadID),
		p.Username,
		p.Anon,
		safeprint(p.Nick),
		utils.ColorToAp(p.Color),
		p.PostedAt.String(),
		blprint(p.Backlinks))
}

func blprint(blbl []uint32) string {
	if blbl == nil || len(blbl) == 0 {
		return ""
	}
	f := strings.Repeat("%d (%s), ", len(blbl))
	a := make([]any, 2*len(blbl))
	for i, v := range blbl {
		a[2*i] = v
		a[2*i+1] = utils.IDToA(v)
	}
	return fmt.Sprintf(f, a...)
}

type Thread struct {
	ID         uint32
	Topic      *string
	PostedAt   time.Time
	BumpedAt   time.Time
	ReplyCount int
	OP         Post
	Posts      []Post
}

func (t *Thread) String() string {
	if t == nil {
		return "nil"
	}
	return fmt.Sprintf(`
		ID: %d (%s)
		Topic: %s
		PostedAt: %s
		BumpedAt: %s
		OP: %s
		len(Posts): %d`, t.ID, utils.IDToA(t.ID), safeprint(t.Topic), t.PostedAt.String(), t.BumpedAt.String(), t.OP.String(), len(t.Posts))
}

func safeprint(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

type TextContent struct {
	Body string
}

type ImageContent struct {
	CID string
	Alt *string
}
