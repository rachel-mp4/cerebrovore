package model

import (
	"github.com/rachel-mp4/cerebrovore/utils"
)

type socketMessage struct {
	Type string      `json:"type"`
	Data socketEvent `json:"data"`
}

// watchEvent represents a bump happening in a thread that a user
// watches; it gets sent on threadsocket to anyone connected
type socketEvent struct {
	SystemMessage *string `json:"systemMessage,omitempty"`
	ID            *uint32 `json:"id,omitempty"`
	Username      *string `json:"username,omitempty"`
	Remaining     *string `json:"remaining,omitempty"`
	BumpLimit     *bool   `json:"bumpLimit,omitempty"`
	ReplyLimit    *bool   `json:"replyLimit,omitempty"`
}

func (m *Model) NotifyReply(tid uint32, id uint32, username *string, replyCount int) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[tid]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	rem := utils.PercentRemaining(&replyCount)
	e := socketMessage{"thread", socketEvent{ID: &id, Username: username, Remaining: &rem}}
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
