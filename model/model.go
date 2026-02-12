package model

import (
	"fmt"
	"net/http"
	"sync"
)

type Model struct {
	tmapmu sync.Mutex
	tmap   map[uint32]*threadModel

	idMu sync.Mutex
	id   uint32
	pid  uint32 // "prestige" id
}

// NewModel creates and initializes a new Model, returning it
func NewModel() *Model {
	m := &Model{}
	m.tmap = make(map[uint32]*threadModel)
	return m
}

// recreateThread is an internal function to recreateThreads
// that already existed, for use in model initialization,
// reading from database
func (m *Model) recreateThread(threadID uint32) {
	rt := &threadModel{}
	m.tmap[threadID] = rt
}

// AddThread allocates and returns the id for a new thread,
// which it creates
func (m *Model) AddThread() uint32 {
	m.tmapmu.Lock()
	defer m.tmapmu.Unlock()
	allocate := m.getIDAllocator()
	threadID := allocate()
	nt := &threadModel{}
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

func (m *Model) GetThreads() []uint32 {
	m.tmapmu.Lock()
	defer m.tmapmu.Unlock()
	res := make([]uint32, len(m.tmap))
	i := 0
	for k := range m.tmap {
		res[i] = k
		i = i + 1
	}
	return res
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

// getIDAllocator produces an IDAllocator function that returns an
// id (uint32) with coordination between all other IDAllocator functions
func (m *Model) getIDAllocator() func() uint32 {
	return func() uint32 {
		m.idMu.Lock()
		defer m.idMu.Unlock()
		next := m.id + 1
		// lrc uses uint32 for message id numbers,
		// which is big, but figured i might as well
		// code a system to allow shared ids across
		// a whole site like imageboards, so prestige
		// id lets us store it in db as uint64 so we
		// don't have to worry about running out of
		// id numbers
		if next < m.id {
			next = next + 1
			m.pid += 1
		}
		m.id = next
		return next
	}
}
