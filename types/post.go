package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/rachel-mp4/cerebrovore/utils"
)

type Post struct {
	ID             uint32
	ThreadID       uint32
	Username       string
	Anon           bool
	Nick           *string
	Color          *uint32
	PostedAt       time.Time
	TextContent    *TextContent
	ImageContent   *ImageContent
	Backlinks      []uint32
	Deleted        bool
	ViewerIsYou    bool
	CanSeeAnon     bool
	LinkToModerate bool
	ProfileHead    *ProfileHead
	FromEnd        *int
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
	if len(blbl) == 0 {
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

type ForumThreadThumb struct {
	ID               uint32
	Topic            *string
	PostedAt         time.Time
	BumpedAt         time.Time
	ReplyCount       int
	OP               Post
	LP               *Post
	Deleted          bool
	ManuallyArchived bool
}

type Thread struct {
	ID               uint32
	Topic            *string
	PostedAt         time.Time
	BumpedAt         time.Time
	ReplyCount       int
	OP               Post
	Posts            []Post
	Deleted          bool
	Dead             bool
	ManuallyArchived bool
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

func TopicOrIdtoa(t Thread) string {
	if t.Topic != nil && *t.Topic != "" {
		return *t.Topic
	}
	return fmt.Sprintf("#%s", utils.IDToA(t.ID))
}

func ForumTopicOrIdtoa(ft ForumThreadThumb) string {
	if ft.Topic != nil && *ft.Topic != "" {
		return *ft.Topic
	}
	return fmt.Sprintf("#%s", utils.IDToA(ft.ID))
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

func Archived(t Thread) bool {
	return t.ManuallyArchived || utils.MaxReplies(t.ReplyCount)
}
