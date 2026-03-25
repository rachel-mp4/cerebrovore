package model

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/rachel-mp4/cerebrovore/utils"
)

// watchEvent represents a bump happening in a thread that a user
// watches; it gets sent on threadsocket to anyone connected
type socketEvent struct {
	SystemMessage *string `json:"systemMessage,omitempty"`
	ID            *uint32 `json:"id,omitempty"`
	Username      *string `json:"username,omitempty"`
	Remaining     *string `json:"remaining,omitempty"`
	BumpLimit     *bool   `json:"new,omitempty"`
	ReplyLimit    *bool   `json:"replyLimit,omitempty"`
}

func (m *Model) GetThreadSocket(tid uint32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.tmapmu.RLock()
		tm, ok := m.tmap[tid]
		m.tmapmu.RUnlock()
		if !ok {
			http.Error(w, "thread does not exist", http.StatusNotFound)
			return
		}

		upgrader := &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		watchr := &watcherConn{}
		ch := make(chan any, 10)
		watchr.ch = ch
		watchr.conn = conn
		watchr.ctx, watchr.cancel = context.WithCancel(r.Context())
		tm.mu.Lock()
		tm.subsmu.Lock()
		if tm.full {
			tm.subsmu.Unlock()
			tm.mu.Unlock()
			return
		}
		tm.subs[watchr] = true
		tm.subsmu.Unlock()
		tm.mu.Unlock()
		go watchr.readloop()
		watchr.watch()
		tm.subsmu.Lock()
		delete(tm.subs, watchr)
		tm.subsmu.Unlock()
	}
}

func (m *Model) NotifyReply(tid uint32, id uint32, username *string, replyCount int) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[tid]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	rem := utils.PercentRemaining(&replyCount)
	e := socketEvent{ID: &id, Username: username, Remaining: &rem}
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
	e := socketEvent{SystemMessage: &msg}
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
