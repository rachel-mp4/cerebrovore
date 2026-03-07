package model

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/rachel-mp4/cerebrovore/utils"
	"github.com/rachel-mp4/lrcd"
)

// threadModel is the model for a thread. if the server is non nil, then
// it should be started
type threadModel struct {
	mu     sync.Mutex
	id     uint32
	topic  *string
	server *lrcd.Server
	full   bool

	watchers   map[*watcher]bool
	watchersmu sync.Mutex
}

// newThreadModel creates a new thread model. it does not create or start
// an lrc server
func newThreadModel(id uint32, topic *string) *threadModel {
	return &threadModel{
		id:       id,
		topic:    topic,
		watchers: make(map[*watcher]bool),
	}
}

// GetWSHandler returns the wshandler for an lrc server, if it exists
func (tm *threadModel) GetWSHandler() (http.HandlerFunc, error) {
	if tm.server == nil {
		return nil, ErrServerDNE
	}
	return tm.server.WSHandler(), nil
}

// recreateServer recreates & starts the server for a given threadModel, according
// to the threadModel's specs + the given idAllocator
func (tm *threadModel) recreateServer(idAllocator func() uint32) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	// perhaps 2 clients requested thread at same time, the first got lock and
	// recreated server while the second was waiting and then the second got
	// lock and the server now exists
	if tm.server != nil {
		return nil
	}
	opts := []lrcd.Option{
		lrcd.WithIDAllocator(idAllocator),
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
