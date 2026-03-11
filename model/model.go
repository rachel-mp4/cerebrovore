package model

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

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

// recreateThread is an internal function to recreateThreads
// that already existed, for use in model initialization,
// reading from database
func (m *Model) recreateThread(thread types.Thread) {
	if utils.MaxReplies(thread.ReplyCount) {
		return
	}
	rt := newThreadModel(thread.ID, thread.Topic)
	if utils.MaxBumps(thread.ReplyCount) {
		rt.bumplimit = true
	}
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
	handler, err := thread.getWSHandler()
	if err != nil {
		return nil, fmt.Errorf("getwshandler: %w", err)
	}
	return handler, nil
}

// AddBacklinks tells the lrc server to send a batch of replies to all lrc connections
func (m *Model) AddBacklinks(threadID uint32, batch lrcpb.Event_Replybatch) {
	tm, ok := m.tmap[threadID]
	if !ok {
		return
	}
	tm.server.SendReplyBatch(&batch)
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

// getIDAllocator produces an IDAllocator function that returns an
// id (uint32) with coordination between all other IDAllocator functions
func (m *Model) getIDAllocator() func() uint32 {
	return func() uint32 {
		next := atomic.AddUint32(&m.id, 1)
		return next
	}
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
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.tmapmu.Lock()
			for id, tm := range m.tmap {
				tm.kawaiiDestroyServer()
				if tm.full && tm.server == nil && len(tm.wormwatchers) == 0 {
					log.Println("threadModelKilled")
					delete(m.tmap, id)
				}
			}
			m.tmapmu.Unlock()
		}
	}
}
