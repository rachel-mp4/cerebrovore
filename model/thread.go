package model

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rachel-mp4/cerebrovore/utils"
	"github.com/rachel-mp4/lrcd"
)

// threadModel is the model for a thread. if the server is non nil, then
// it should be started
// i grouped the fields based on which mutex relates to what
type threadModel struct {
	id     uint32
	topic  *string
	server *lrcd.Server
	full   bool
	dead   bool
	mu     sync.Mutex

	subs   map[*clientConn]bool // these are connections who are currently in the thread
	subsmu sync.RWMutex         // fka thread socket

	watchers   map[string]bool // keys are usernames
	bumplimit  bool
	watchersmu sync.RWMutex

	wormwatchdata  *wormwatchdata
	wormwatchers   map[*clientConn]bool // values mean nothing
	wormwatchersmu sync.Mutex
}

func (m *Model) GetMessageData(tid uint32, pid uint32) (username *string, curState *string, err error) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[tid]
	m.tmapmu.RUnlock()
	if !ok {
		err = ErrThreadDNE
		return
	}
	if tm.dead {
		err = ErrThreadDead
		return
	}
	tm.mu.Lock()
	if tm.server != nil {
		username, curState, err = tm.server.GetExternIDAndBodyFrom(pid)
	} else {
		err = ErrServerDNE
	}
	tm.mu.Unlock()
	return
}

// newThreadModel creates a new thread model. it does not create or start
// an lrc server
func newThreadModel(id uint32, topic *string) *threadModel {
	return &threadModel{
		id:           id,
		topic:        topic,
		subs:         make(map[*clientConn]bool),
		watchers:     make(map[string]bool),
		wormwatchers: make(map[*clientConn]bool),
		wormwatchdata: &wormwatchdata{
			index:    0,
			watchers: make(map[string]int),
		},
	}
}

func newDeadThreadModel(id uint32, topic *string) *threadModel {
	return &threadModel{
		id:       id,
		dead:     true,
		topic:    topic,
		subs:     make(map[*clientConn]bool),
		watchers: make(map[string]bool),
	}
}

// GetWSHandler returns the wshandler for an lrc server, if it exists
func (tm *threadModel) getWSHandler(username string) (http.HandlerFunc, error) {
	if tm.dead {
		return nil, ErrThreadDead
	}
	if tm.server == nil {
		return nil, ErrServerDNE
	}
	return tm.server.WSHandlerExternal(username), nil
}

// recreateServer recreates & starts the server for a given threadModel, according
// to the threadModel's specs + the given idAllocator
func (tm *threadModel) recreateServer(idAllocator func() uint32) error {
	// perhaps 2 clients requested thread at same time, the first got lock and
	// recreated server while the second was waiting and then the second got
	// lock and the server now exists
	if tm.dead {
		return ErrThreadDead
	}
	if tm.server != nil {
		return nil
	}
	opts := []lrcd.Option{
		lrcd.WithIDAllocator(idAllocator),
		lrcd.WithConsumerSetExternalId(),
		lrcd.WithSlowCleanup(5 * time.Minute),
		lrcd.WithServerURIAndSecret(utils.IDToA(tm.id), os.Getenv("LRCD_SECRET")),
	}
	if tm.topic != nil {
		opts = append(opts, lrcd.WithWelcome(*tm.topic))
	}
	s, err := lrcd.NewServer(
		opts...,
	)
	if err != nil {
		return fmt.Errorf("newserver: %w", err)
	}
	tm.server = s
	err = tm.server.Start()
	if err != nil {
		tm.server = nil
		return fmt.Errorf("server start: %w", err)
	}
	return nil
}

// destroyServer stops a server, returning an error
// if there was some issue stopping it (likely it was already stopped)
func (tm *threadModel) destroyServer() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.server == nil {
		return nil
	}
	_, err := tm.server.Stop()
	if err != nil {
		return fmt.Errorf("server stop: %w", err)
	}
	tm.server = nil
	return nil
}

// kawaiiDestroyServer only kills the server if it's empty uwu
func (tm *threadModel) kawaiiDestroyServer() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if tm.server == nil {
		return
	}
	stopped := tm.server.StopIfEmpty()
	if stopped {
		tm.server = nil
	}
}
