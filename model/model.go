package model

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
	lrcpb "github.com/rachel-mp4/lrcproto/gen/go"
)

type Model struct {
	tmapmu sync.Mutex
	tmap   map[uint32]*threadModel

	id  uint32
	pid uint32 // "prestige" id // lol what drugs was i onnnnnn
}

// NewModel creates and initializes a new Model, returning it
func NewModel(threads []types.Thread, maxid uint32) *Model {
	m := &Model{id: maxid}
	m.tmap = make(map[uint32]*threadModel, len(threads))
	for _, thread := range threads {
		m.recreateThread(thread)
	}
	go cleaner(m)
	return m
}

// cleaner cleans up any empty servers every 10 minutes, i just
// picked the time scale at random. i prefer for it to not be
// immediate just because that way you can't constantly create
// and destroy servers by constantly refreshing some empty
// thread. cleaner is a bit of a costly operation, maybe there's
// a better way of doing this, but it's fine for now. if thread
// has hit max reply count & there are no longer any users
// connected to the lrc server (thus kawaiiDestroyServer
// succesfully destroyed it & it's set to nil), then we can
// permanently remove the threadModel from our map, since we
// don't need any real-time functions out of it any more: it is
// succesfully archived
func cleaner(m *Model) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			m.tmapmu.Lock()
			for id, tm := range m.tmap {
				tm.kawaiiDestroyServer()
				if tm.full && tm.server == nil {
					log.Println("threadModelKilled")
					delete(m.tmap, id)
				}
			}
			m.tmapmu.Unlock()
		}
	}
}

// recreateThread is an internal function to recreateThreads
// that already existed, for use in model initialization,
// reading from database
func (m *Model) recreateThread(thread types.Thread) {
	if utils.MaxReplies(thread.ReplyCount) {
		log.Println("tc >= reply limit = archived thread = no threadmodel")
		return
	}
	rt := newThreadModel(thread.ID, thread.Topic)
	m.tmap[thread.ID] = rt
}

// AddThread allocates and returns the id for a new thread,
// which it creates
func (m *Model) AddThread(topic *string) uint32 {
	m.tmapmu.Lock()
	defer m.tmapmu.Unlock()
	threadID := m.getIDAllocator()()
	nt := newThreadModel(threadID, topic)
	m.tmap[threadID] = nt
	return threadID
}

// DeleteThread deletes a thread after destroying it's server
func (m *Model) DeleteThread(threadID uint32) error {
	m.tmapmu.Lock()
	defer m.tmapmu.Unlock()
	thread, ok := m.tmap[threadID]
	if !ok {
		return ErrThreadDNE
	}
	err := thread.destroyServer()
	if err != nil {
		return fmt.Errorf("destroy server: %w", err)
	}
	delete(m.tmap, threadID)
	return nil
}

// GetThreadWSHandler gets the ws handler for a thread's lrc server
func (m *Model) GetThreadWSHandler(threadID uint32) (http.HandlerFunc, error) {
	m.tmapmu.Lock()
	defer m.tmapmu.Unlock()
	thread, ok := m.tmap[threadID]
	if !ok {
		return nil, ErrThreadDNE
	}
	if thread.full {
		return nil, ErrThreadFull
	}
	if thread.server == nil {
		err := thread.recreateServer(m.getIDAllocator())
		if err != nil {
			return nil, fmt.Errorf("recreateserver: %w", err)
		}
	}
	handler, err := thread.GetWSHandler()
	if err != nil {
		return nil, fmt.Errorf("getwshandler: %w", err)
	}
	return handler, nil
}

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
		// maybe not necessary to delete a connection in this case,
		// however i think that things get written quickly, and the channel
		// is buffered, so i'm assuming they disconnected but the context
		// isn't done for some reason
		default:
			delete(tm.watchers, w)
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
	bl := true
	tm.watchersmu.Lock()
	for w := range tm.watchers {
		select {
		case w.ch <- watchEvent{tm.topic, threadID, &bl}:
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
			tm.watchers[watcher] = true
			tm.watchersmu.Unlock()
		}
		m.tmapmu.Unlock()
		watcher.watch()
	}
}

// watch writes events to a watcher's threadsocket
func (w *watcher) watch() {
	defer w.cancel()
	ticker := time.NewTicker(15 * time.Second)
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

// AddBacklinks tells the lrc server to send a batch of replies to all lrc connections
func (m *Model) AddBacklinks(threadID uint32, batch lrcpb.Event_Replybatch) {
	tm, ok := m.tmap[threadID]
	if !ok {
		return
	}
	tm.server.SendReplyBatch(&batch)
}

// getIDAllocator produces an IDAllocator function that returns an
// id (uint32) with coordination between all other IDAllocator functions
func (m *Model) getIDAllocator() func() uint32 {
	return func() uint32 {
		next := atomic.AddUint32(&m.id, 1)
		return next
	}
}

func (m *Model) ReplyLimit(threadID uint32) {
	m.tmapmu.Lock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.Unlock()
	if !ok {
		return
	}
	tm.full = true
}
