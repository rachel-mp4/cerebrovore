package model

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// watchEvent represents a bump happening in a thread that a user
// watches; it gets sent on threadsocket to anyone connected
type watchEvent struct {
	Topic     *string `json:"topic,omitempty"`
	ID        uint32  `json:"id"`
	BumpLimit *bool   `json:"bumpLimit,omitempty"`
}

// watcher represents an open thread watcher connection, this is
// created for basically every tab that you have open of the site,
// since ATM like all the templates involve the list of threads
type watcher struct {
	conn   *websocket.Conn
	ch     chan watchEvent
	ctx    context.Context
	cancel context.CancelFunc
}

// NotifyWatchers notifies all online watchers of a thread that a
// bump just occurred
func (m *Model) NotifyWatchers(forID uint32) {
	m.tmapmu.Lock()
	tm, ok := m.tmap[forID]
	m.tmapmu.Unlock()
	if !ok {
		return
	}
	tm.watchersmu.Lock()
	for w := range tm.watchers {
		select {
		case w.ch <- watchEvent{tm.topic, forID, nil}:
		case <-w.ctx.Done():
			delete(tm.watchers, w)
		default:
		}
	}
	tm.watchersmu.Unlock()
}

// NotifyBumpLimit notifies all online watchers of a thread that the
// thread just hit the bump limit, and it removes them all from the
// map for good measure
func (m *Model) NotifyBumpLimit(threadID uint32) {
	m.tmapmu.Lock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.Unlock()
	if !ok {
		return
	}
	tm.watchersmu.Lock()
	tm.bumplimit = true
	for w := range tm.watchers {
		select {
		case w.ch <- watchEvent{tm.topic, threadID, &tm.bumplimit}:
			delete(tm.watchers, w)
		default:
			delete(tm.watchers, w)
		}
	}
	tm.watchersmu.Unlock()
}

// GetThreadSocketHandler gets the websocket handler for a given user's
// collection of threadIDs that they are watching
func (m *Model) GetThreadSocketHandler(threadIDs []uint32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		watcher := &watcher{}
		ch := make(chan watchEvent, 10)
		watcher.ch = ch
		watcher.conn = conn
		watcher.ctx, watcher.cancel = context.WithCancel(r.Context())
		m.tmapmu.Lock()
		for _, tid := range threadIDs {
			tm, ok := m.tmap[tid]
			if !ok {
				continue
			}
			tm.watchersmu.Lock()
			if !tm.bumplimit {
				tm.watchers[watcher] = true
			}
			tm.watchersmu.Unlock()
		}
		m.tmapmu.Unlock()
		go watcher.readloop()
		watcher.watch()
	}
}

// watch writes events to a watcher's threadsocket
func (w *watcher) watch() {
	defer w.cancel()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			err := w.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
			if err != nil {
				return
			}
		case we := <-w.ch:
			err := w.conn.WriteJSON(we)
			if err != nil {
				return
			}
		}
	}
}

func (w *watcher) readloop() {
	defer w.cancel()
	w.conn.SetReadLimit(512)
	w.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	w.conn.SetPongHandler(func(string) error {
		w.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		if _, _, err := w.conn.ReadMessage(); err != nil {
			return
		}
	}
}
