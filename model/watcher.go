package model

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type watchMessage struct {
	Type string     `json:"type"`
	Data watchEvent `json:"data"`
}

// watchEvent represents a bump happening in a thread that a user
// watches; it gets sent on threadsocket to anyone connected
type watchEvent struct {
	Topic     *string `json:"topic,omitempty"`
	ID        uint32  `json:"id"`
	BumpLimit *bool   `json:"bumpLimit,omitempty"`
	New       *bool   `json:"new,omitempty"`
}

// clientConn represents an open thread watcher connection, this is
// created for basically every tab that you have open of the site,
// since ATM like all the templates involve the list of threads
type clientConn struct {
	conn   *websocket.Conn
	ch     chan any
	ctx    context.Context
	cancel context.CancelFunc
}

// userWatcherCtx represents all of a user's currently open connections
// alongside all threads that they are currently watching (so we can remove
// them from all those threads after their final openConn disconnects)
type userWatcherCtx struct {
	threadsWatched map[uint32]bool // keys are thread ids
	twmu           sync.Mutex

	openConns map[*clientConn]bool // values are if they want new threads
	ocmu      sync.RWMutex

	cleanupTimer *time.Timer
	cleanupmu    sync.Mutex
}

func (m *Model) Watch(username string, threadID uint32) (bumplimit bool) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.RUnlock()
	if !ok {
		return false
	}
	m.watchersmu.RLock()
	watcherCtx, ok := m.watchers[username]
	m.watchersmu.RUnlock()
	if !ok {
		return false
	}
	tm.watchersmu.Lock()
	if tm.bumplimit {
		tm.watchersmu.Unlock()
		return true
	}
	tm.watchers[username] = true
	tm.watchersmu.Unlock()
	watcherCtx.twmu.Lock()
	watcherCtx.threadsWatched[threadID] = true
	watcherCtx.twmu.Unlock()
	return false
}

func (m *Model) Unwatch(username string, threadID uint32) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	m.watchersmu.RLock()
	watcherCtx, ok := m.watchers[username]
	m.watchersmu.RUnlock()
	if !ok {
		return
	}
	tm.watchersmu.Lock()
	delete(tm.watchers, username)
	tm.watchersmu.Unlock()
	watcherCtx.twmu.Lock()
	delete(watcherCtx.threadsWatched, threadID)
	watcherCtx.twmu.Unlock()
}

func (m *Model) NotifyNewThread(threadID uint32) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	knew := true
	m.watchersmu.RLock()
	for _, uwctx := range m.watchers {
		uwctx.ocmu.RLock()
		for w, ok := range uwctx.openConns {
			if !ok {
				continue
			}
			select {
			case w.ch <- watchMessage{"watcher", watchEvent{Topic: tm.topic, ID: threadID, BumpLimit: nil, New: &knew}}:
			default:
			}
		}
		uwctx.ocmu.RUnlock()
	}
	m.watchersmu.RUnlock()
}

// NotifyWatchers notifies all online watchers of a thread that a
// bump just occurred
func (m *Model) NotifyWatchers(forID uint32) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[forID]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	tm.watchersmu.RLock()
	usernames := make([]string, 0, len(tm.watchers))
	for w := range tm.watchers {
		usernames = append(usernames, w)
	}
	tm.watchersmu.RUnlock()
	for _, u := range usernames {
		m.watchersmu.RLock()
		watcherctx, ok := m.watchers[u]
		m.watchersmu.RUnlock()
		if !ok {
			continue
		}
		watcherctx.ocmu.RLock()
		for w := range watcherctx.openConns {
			select {
			case w.ch <- watchMessage{"watcher", watchEvent{tm.topic, forID, nil, nil}}:
			default:
			}
		}
		watcherctx.ocmu.RUnlock()
	}
}

// NotifyBumpLimit notifies all online watchers of a thread that the
// thread just hit the bump limit, and it removes them all from the
// map for good measure
func (m *Model) NotifyBumpLimit(threadID uint32) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	tm.watchersmu.Lock()
	tm.bumplimit = true
	usernames := make([]string, 0, len(tm.watchers))
	for w := range tm.watchers {
		usernames = append(usernames, w)
		delete(tm.watchers, w)
	}
	tm.watchersmu.Unlock()
	go func() {
		bl := true
		for _, u := range usernames {
			m.watchersmu.RLock()
			watcherctx, ok := m.watchers[u]
			m.watchersmu.RUnlock()
			if !ok {
				continue
			}
			watcherctx.ocmu.RLock()
			for w := range watcherctx.openConns {
				select {
				case w.ch <- watchMessage{"watcher", watchEvent{tm.topic, threadID, &bl, nil}}:
				default:
				}
			}
			watcherctx.ocmu.RUnlock()
			watcherctx.twmu.Lock()
			delete(watcherctx.threadsWatched, threadID)
			watcherctx.twmu.Unlock()
		}
	}()
	go func() {
		tm.subsmu.RLock()
		defer tm.subsmu.RUnlock()
		for w := range tm.subs {
			select {
			case w.ch <- socketMessage{"thread", socketEvent{BumpLimit: &tm.bumplimit}}:
			default:
			}
		}
	}()
}

// GetThreadSocketHandler gets the websocket handler for a given user's
// collection of threadIDs that they are watching
// order of things is a bit weird, don't wanna make db call until we know
// that there's no redundant info. there's a potential data race here,
// getWatchedThreads + Unwatch thread -> unwatch deletes from map of threads
// -> getWatchedThreads adds it back into map of threads, but that seems rare
// and is less bad than the other race where a watched thread would never send
// watch notifications (either thing is solved by ending all connections from a
// user or cycling watch after their userWatcherCtx stabilizes)

// watch writes events to a watcher's threadsocket
func (w *clientConn) watch() {
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

func (w *clientConn) readloop() {
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
