package model

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/rachel-mp4/cerebrovore/clog"
)

// watchEvent represents a bump happening in a thread that a user
// watches; it gets sent on threadsocket to anyone connected
type watchEvent struct {
	Topic     *string `json:"topic,omitempty"`
	ID        uint32  `json:"id"`
	BumpLimit *bool   `json:"bumpLimit,omitempty"`
	New       *bool   `json:"new,omitempty"`
}

// watcherConn represents an open thread watcher connection, this is
// created for basically every tab that you have open of the site,
// since ATM like all the templates involve the list of threads
type watcherConn struct {
	conn   *websocket.Conn
	ch     chan any
	ctx    context.Context
	cancel context.CancelFunc
}

// userWatcherCtx represents all of a user's currently open connections
// alongside all threads that they are currently watching (so we can remove
// them from all those threads after their final openConn disconnects)
type userWatcherCtx struct {
	threadsWatched map[uint32]bool
	twmu           sync.Mutex
	openConns      map[*watcherConn]bool
	ocmu           sync.RWMutex
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
		for w := range uwctx.openConns {
			select {
			case w.ch <- watchEvent{Topic: tm.topic, ID: threadID, BumpLimit: nil, New: &knew}:
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
			case w.ch <- watchEvent{tm.topic, forID, nil, nil}:
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
				case w.ch <- watchEvent{tm.topic, threadID, &bl, nil}:
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
			case w.ch <- socketEvent{BumpLimit: &tm.full}:
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
func (m *Model) GetWatcherHandler(username string, getWatchedThreads func(string, context.Context) ([]uint32, error)) http.HandlerFunc {
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
		watchr := &watcherConn{}
		ch := make(chan any, 10)
		watchr.ch = ch
		watchr.conn = conn
		watchr.ctx, watchr.cancel = context.WithCancel(r.Context())
		m.watchersmu.Lock()
		uwctx, ok := m.watchers[username]
		if ok {
			uwctx.ocmu.Lock()
			uwctx.openConns[watchr] = true
			uwctx.ocmu.Unlock()
		} else {
			uwctx = &userWatcherCtx{
				threadsWatched: make(map[uint32]bool),
				openConns:      make(map[*watcherConn]bool, 1),
			}
			uwctx.openConns[watchr] = true
			m.watchers[username] = uwctx
		}
		m.watchersmu.Unlock()
		if !ok {
			// i don't think that we need watchersmu anymore for this
			// second half, because since we still have our reference to
			// uwctx and since we are an open conn that can't terminate
			// yet, the reference in the m.watchers map can't be deleted
			// by some other conn connecting and disconnecting
			threadIDs, err := getWatchedThreads(username, r.Context())
			if err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					clog.Warn("db err: %s", err.Error())
					threadIDs = nil
				}
			}
			for _, tid := range threadIDs {
				m.tmapmu.RLock()
				tm, ok := m.tmap[tid]
				m.tmapmu.RUnlock()
				if !ok {
					continue
				}
				tm.watchersmu.Lock()
				if !tm.bumplimit {
					tm.watchers[username] = true
					uwctx.twmu.Lock()
					uwctx.threadsWatched[tid] = true
					uwctx.twmu.Unlock()
				}
				tm.watchersmu.Unlock()
			}
		}

		go watchr.readloop()
		watchr.watch()
		m.watchersmu.Lock()
		uwctx = m.watchers[username]
		if len(uwctx.openConns) == 1 {
			delete(m.watchers, username)
			for tid := range uwctx.threadsWatched {
				m.tmapmu.RLock()
				tm, ok := m.tmap[tid]
				m.tmapmu.RUnlock()
				if !ok {
					continue
				}
				tm.watchersmu.Lock()
				delete(tm.watchers, username)
				tm.watchersmu.Unlock()
			}
			m.watchersmu.Unlock()
		} else {
			m.watchersmu.Unlock()
			uwctx.ocmu.Lock()
			delete(uwctx.openConns, watchr)
			uwctx.ocmu.Unlock()
		}
	}
}

// watch writes events to a watcher's threadsocket
func (w *watcherConn) watch() {
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

func (w *watcherConn) readloop() {
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
