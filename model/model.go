package model

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
	lrcpb "github.com/rachel-mp4/lrcproto/gen/go"
)

type Model struct {
	tmapmu sync.RWMutex
	tmap   map[uint32]*threadModel //keys are thread ids

	watchersmu sync.RWMutex
	watchers   map[string]*userWatcherCtx //keys are usernames

	id  uint32
	pid uint32 // "prestige" id // lol what drugs was i onnnnnn
}

// NewModel creates and initializes a new Model, returning it
func NewModel(threads []types.Thread, maxid uint32) *Model {
	m := &Model{id: maxid}
	m.tmap = make(map[uint32]*threadModel, len(threads))
	m.watchers = make(map[string]*userWatcherCtx)
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
	m.tmapmu.RLock()
	defer m.tmapmu.RUnlock()
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
func (m *Model) GetThreadWSHandler(threadID uint32, username string) (http.HandlerFunc, error) {
	clog.Dbug("acquiring tmapmulock wsh")
	m.tmapmu.RLock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.RUnlock()
	clog.Dbug("tmapmulock acquired")
	if !ok {
		return nil, ErrThreadDNE
	}
	clog.Dbug("acquiring lock")
	tm.mu.Lock()
	clog.Dbug("lock acquired")
	if tm.full {
		tm.mu.Unlock()
		return nil, ErrThreadFull
	}
	if tm.server == nil {
		clog.Dbug("recreating server")
		err := tm.recreateServer(m.getIDAllocator())
		if err != nil {
			tm.mu.Unlock()
			return nil, fmt.Errorf("recreateserver: %w", err)
		}
	}
	handler, err := tm.getWSHandler(username)
	if err != nil {
		tm.mu.Unlock()
		return nil, fmt.Errorf("getwshandler: %w", err)
	}
	tm.mu.Unlock()
	clog.Dbug("returning handler")
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
	m.tmapmu.RLock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	tm.mu.Lock()
	tm.full = true
	tm.mu.Unlock()
	go func() {
		tm.subsmu.RLock()
		defer tm.subsmu.RUnlock()
		for w := range tm.subs {
			select {
			case w.ch <- socketMessage{"thread", socketEvent{ReplyLimit: &tm.full}}:
			default:
			}
		}
	}()
}

// it would be nice to try and collate all the non-lrc websockets
// into one websocket per page so that i can better kick banned users.
// it's not a big deal that if they modify their client, they can
// theoretically continue to recieve bumps and wormwatch events, but
// this is still not desired behavior, and a signal that we should
// refactor as appropriate
func (m *Model) BanUser(name string) {
	m.watchersmu.RLock()
	uwctx, ok := m.watchers[name]
	m.watchersmu.RUnlock()
	if ok && uwctx != nil {
		uwctx.ocmu.RLock()
		for w := range uwctx.openConns {
			w.cancel()
		}
		uwctx.ocmu.RUnlock()
	}
	m.tmapmu.RLock()
	for _, tm := range m.tmap {
		go func() {
			tm.mu.Lock() // maybe excessive haha, no null chaining operator in go
			if tm.server != nil {
				tm.server.KickExternalId(name)
			}
			tm.mu.Unlock()
		}()
	}
	m.tmapmu.RUnlock()
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
				if tm.full && tm.server == nil && len(tm.wormwatchers) == 0 && len(tm.subs) == 0 {
					clog.Dbug("thread model killed: %d", id)
					delete(m.tmap, id)
				}
			}
			m.tmapmu.Unlock()
		}
	}
}

type Option = func(options *options) error

type options = struct {
	thread            *uint32
	threadsocket      bool
	wormwatch         bool
	watchedthreads    bool
	getwatchedthreads func(string, context.Context) ([]uint32, error)
	newthreads        bool
}

func WithThreadSocket(id uint32) Option {
	return func(options *options) error {
		if options.thread != nil && id != *options.thread {
			return errors.New("can only watch one thread")
		}
		options.thread = &id
		options.threadsocket = true
		return nil
	}
}

func WithWatchedThreads(getwatchedthreads func(string, context.Context) ([]uint32, error), andNewThreads bool) Option {
	return func(options *options) error {
		options.getwatchedthreads = getwatchedthreads
		options.watchedthreads = true
		options.newthreads = andNewThreads
		return nil
	}
}

func WithWormwatch(id uint32) Option {
	return func(options *options) error {
		if options.thread != nil && id != *options.thread {
			return errors.New("can only watch one thread")
		}
		options.thread = &id
		options.wormwatch = true
		return nil
	}
}

func (m *Model) GetWebSockets(username string, opts ...Option) (http.HandlerFunc, error) {
	myoptions := options{}
	for _, opt := range opts {
		err := opt(&myoptions)
		if err != nil {
			return nil, err
		}
	}
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
		client := &clientConn{}
		client.conn = conn
		client.ch = make(chan any, 10)
		client.ctx, client.cancel = context.WithCancel(context.Background())
		var tm *threadModel
		var ok bool

		if myoptions.thread != nil {
			m.tmapmu.RLock()
			tm, ok = m.tmap[*myoptions.thread]
			m.tmapmu.RUnlock()
		}

		if ok && !tm.full && myoptions.threadsocket {
			go func() {
				tm.subsmu.Lock()
				tm.subs[client] = true
				tm.subsmu.Unlock()
			}()
			defer func() {
				tm.subsmu.Lock()
				delete(tm.subs, client)
				tm.subsmu.Unlock()
			}()
		}

		if ok && !tm.full && myoptions.wormwatch {
			go func() {
				tm.wormwatchersmu.Lock()
				defer tm.wormwatchersmu.Unlock()
				tm.wormwatchers[client] = true
				wwd := tm.wormwatchdata
				wwd.watchers[username] = wwd.watchers[username] + 1
				tnowms := time.Now().UnixMilli()
				client.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypeTimeS, Timestamp: &tnowms}}
				if wwd.queue != nil {
					client.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypeQueue, Entries: wwd.queue}}
				}
				if wwd.start != nil {
					if wwd.pausedAt == nil {
						tstartms := wwd.start.UnixMilli()
						client.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypeStart, Timestamp: &tstartms, Index: &wwd.index}}
					} else {
						tpausems := wwd.pausedAt.Milliseconds()
						client.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypePause, Timestamp: &tpausems, Index: &wwd.index}}
					}
				}
			}()
			defer func() {
				tm.wormwatchersmu.Lock()
				defer tm.wormwatchersmu.Unlock()
				delete(tm.wormwatchers, client)
				wwd := tm.wormwatchdata
				count := wwd.watchers[username] - 1
				if count <= 0 {
					delete(wwd.watchers, username)
					// should recheck skip / pause condition here if i want to be really cool
					// but that's a pain since locks
				} else {
					wwd.watchers[username] = count
				}
			}()
		}

		if myoptions.watchedthreads {
			go func() {
				m.watchersmu.Lock()
				uwctx, ok := m.watchers[username]
				if ok {
					uwctx.ocmu.Lock()
					uwctx.openConns[client] = true
					uwctx.ocmu.Unlock()
					uwctx.cleanupmu.Lock()
					if uwctx.cleanupTimer != nil {
						uwctx.cleanupTimer.Stop()
						uwctx.cleanupTimer = nil
					}
					m.watchers[username] = uwctx
					uwctx.cleanupmu.Unlock()
				} else {
					uwctx = &userWatcherCtx{
						threadsWatched: make(map[uint32]bool),
						openConns:      make(map[*clientConn]bool, 1),
					}
					uwctx.openConns[client] = true
					m.watchers[username] = uwctx
				}
				m.watchersmu.Unlock()
				if !ok {
					// i don't think that we need watchersmu anymore for this
					// second half, because since we still have our reference to
					// uwctx and since we are an open conn that can't terminate
					// yet, the reference in the m.watchers map can't be deleted
					// by some other conn connecting and disconnecting
					// edited apr 2 2026: this is still true as long as this
					// doesn't take more than 10 seconds and user doesn't
					// immediately disconnect
					threadIDs, err := myoptions.getwatchedthreads(username, r.Context())
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
			}()

			// get the watch context
			// remove client from open conns
			// check remaining
			// if its zero, then we need to queue a cleanup
			// then we get locks to make sure the cleanup happens
			// then we cleanup for each thread i watch
			defer func() {
				m.watchersmu.RLock()
				uwctx := m.watchers[username]
				m.watchersmu.RUnlock()
				uwctx.ocmu.Lock()
				delete(uwctx.openConns, client)
				remaining := len(uwctx.openConns)
				uwctx.ocmu.Unlock()
				if remaining == 0 {
					uwctx.cleanupmu.Lock()
					if uwctx.cleanupTimer != nil {
						panic("i think that this is a weird case! cleanup watch context")
					}
					uwctx.cleanupTimer = time.AfterFunc(10*time.Second, func() {
						// probably i'm being a bit wow shiny toy apropos mutex
						// not necessary and introduce a lot of unnecessary overhead, but
						// anyway, it needs to go this order of nesting since that's
						// how the uwctx initialize goes, to prevent deadlock
						m.watchersmu.Lock()
						defer m.watchersmu.Unlock()
						uwctx.ocmu.RLock()
						defer uwctx.ocmu.RUnlock()
						uwctx.cleanupmu.Lock()
						defer uwctx.cleanupmu.Unlock()

						if len(uwctx.openConns) == 0 {
							delete(m.watchers, username)
						}
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
					})
					uwctx.cleanupmu.Unlock()
				}
			}()
		}
		go client.readloop()
		client.watch()
	}, nil
}
