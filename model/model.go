package model

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rachel-mp4/cerebrovore/types"
	lrcpb "github.com/rachel-mp4/lrcproto/gen/go"
	"net/http"
	"sync"
	"sync/atomic"
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
	rt := &threadModel{topic: thread.Topic, id: thread.ID}
	m.tmap[thread.ID] = rt
}

// AddThread allocates and returns the id for a new thread,
// which it creates
func (m *Model) AddThread(topic *string) uint32 {
	m.tmapmu.Lock()
	defer m.tmapmu.Unlock()
	threadID := m.getIDAllocator()()
	nt := &threadModel{id: threadID, topic: topic}
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
		case w.ch <- forID:
		default:
			delete(tm.watchers, w)
		}
	}
	tm.watchersmu.Unlock()
}

type watcher struct {
	conn *websocket.Conn
	ch   chan uint32
}

func (m *Model) GetThreadSocketHandler(threadIDs []uint32) http.HandlerFunc {
	watcher := &watcher{}
	ch := make(chan uint32, 10)
	watcher.ch = ch
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
		watcher.conn = conn
	}
}

func (w *watcher) Watch() {

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
