package model

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rachel-mp4/cerebrovore/types"
	lrcpb "github.com/rachel-mp4/lrcproto/gen/go"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type Model struct {
	tmapmu sync.Mutex
	tmap   map[uint32]*threadModel

	id  uint32
	pid uint32 // "prestige" id
}

// NewModel creates and initializes a new Model, returning it
func NewModel(threads []types.Thread, mid uint32) *Model {
	m := &Model{id: mid}
	m.tmap = make(map[uint32]*threadModel, len(threads))
	for _, thread := range threads {
		m.recreateThread(thread)
	}
	return m
}

// recreateThread is an internal function to recreateThreads
// that already existed, for use in model initialization,
// reading from database
func (m *Model) recreateThread(thread types.Thread) {
	rt := &threadModel{topic: thread.Topic, id: thread.ID, watchers: make(map[*watcher]bool)}
	m.tmap[thread.ID] = rt
}

// AddThread allocates and returns the id for a new thread,
// which it creates
func (m *Model) AddThread(topic *string) uint32 {
	m.tmapmu.Lock()
	defer m.tmapmu.Unlock()
	threadID := m.getIDAllocator()()
	nt := &threadModel{id: threadID, topic: topic, watchers: make(map[*watcher]bool)}
	m.tmap[threadID] = nt
	return threadID
}

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

func (m *Model) GetThread(threadID uint32) error {
	m.tmapmu.Lock()
	defer m.tmapmu.Unlock()
	_, ok := m.tmap[threadID]
	if !ok {
		return ErrThreadDNE
	}
	return nil
}

func (m *Model) GetThreadWSHandler(threadID uint32) (http.HandlerFunc, error) {
	m.tmapmu.Lock()
	defer m.tmapmu.Unlock()
	thread, ok := m.tmap[threadID]
	if !ok {
		return nil, ErrThreadDNE
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

type watchEvent struct {
	Topic *string `json:"topic,omitempty"`
	ID    uint32  `json:"id"`
}

type watcher struct {
	conn   *websocket.Conn
	ch     chan watchEvent
	ctx    context.Context
	cancel context.CancelFunc
}

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
		case w.ch <- watchEvent{tm.topic, forID}:
			log.Println("send")
		case <-w.ctx.Done():
			delete(tm.watchers, w)
			log.Println("delete from ctx")
		default:
			delete(tm.watchers, w)
			log.Println("delete")
		}
	}
	tm.watchersmu.Unlock()
}

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
			log.Println("attach!")
		}
		m.tmapmu.Unlock()
		watcher.watch()
	}
}

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
