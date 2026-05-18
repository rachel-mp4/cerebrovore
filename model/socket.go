package model

import (
	"github.com/rachel-mp4/cerebrovore/utils"
)

type socketMessage struct {
	Type string      `json:"type"`
	Data socketEvent `json:"data"`
}

// socketEvent represents a reply happening in a thread that a user
// watches; it gets sent on threadsocket to anyone connected
type socketEvent struct {
	SystemMessage *string `json:"systemMessage,omitempty"`
	ID            *uint32 `json:"id,omitempty"`
	Color         *uint32 `json:"color,omitempty"`
	Deleted       *bool   `json:"deleted,omitempty"`
	Username      *string `json:"username,omitempty"`
	Remaining     *string `json:"remaining,omitempty"`
	BumpLimit     *bool   `json:"bumpLimit,omitempty"`
	ReplyLimit    *bool   `json:"replyLimit,omitempty"`
	RenderedHTML  *string `json:"renderedHTML,omitempty"`
}

func (m *Model) NotifyReply(tid uint32, id uint32, username *string, color *uint32, replyCount int, renderedHTML *string) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[tid]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	rem := utils.PercentRemaining(&replyCount)
	e := socketMessage{"thread", socketEvent{ID: &id, Username: username, Color: color, Remaining: &rem, RenderedHTML: renderedHTML}}
	go func() {
		tm.subsmu.RLock()
		defer tm.subsmu.RUnlock()
		for w := range tm.subs {
			select {
			case w.ch <- e:
			default:
			}
		}
	}()
}

func (m *Model) NotifyDelete(tid, id uint32) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[tid]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	yes := true
	e := socketMessage{"thread", socketEvent{ID: &id, Deleted: &yes}}
	go func() {
		tm.subsmu.RLock()
		defer tm.subsmu.RUnlock()
		for w := range tm.subs {
			select {
			case w.ch <- e:
			default:
			}
		}
	}()
}

func (m *Model) SystemMessage(msg string) {
	e := socketMessage{"thread", socketEvent{SystemMessage: &msg}}
	m.tmapmu.RLock()
	for _, tm := range m.tmap {
		tm.subsmu.RLock()
		for w := range tm.subs {
			select {
			case w.ch <- e:
			default:
			}
		}
		tm.subsmu.RUnlock()
	}
	m.tmapmu.RUnlock()
}
