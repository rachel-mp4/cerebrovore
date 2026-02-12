package model

import (
	"fmt"
	"github.com/rachel-mp4/lrcd"
	"net/http"
	"sync"
)

// threadModel is the model for a thread. if the server is non nil, then
// it should be started
type threadModel struct {
	mu     sync.Mutex
	server *lrcd.Server
}

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
	s, err := lrcd.NewServer(
		lrcd.WithIDAllocator(idAllocator),
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
