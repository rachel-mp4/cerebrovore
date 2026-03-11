package model

import (
	"context"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rachel-mp4/cerebrovore/utils"
)

const bufferDuration = 1 * time.Second

type wormwatchdata struct {
	index int
	// start is the time that the video started at (or will start at)
	start *time.Time
	// pausedAt is the duration after the video start time that we paused at
	pausedAt *time.Duration
	ctx      context.Context
	cancel   context.CancelCauseFunc
	queue    []*wormwatchqueueentry
	watchers map[string]int
}

type wormwatchqueueentry struct {
	username  string
	voteskip  map[string]bool
	votepause map[string]bool
	Data      utils.PlayInput `json:"data"`
	Index     int             `json:"index"`
}

type wormwatchEvent struct {
	Type      string                 `json:"type"`
	Entries   []*wormwatchqueueentry `json:"entries,omitempty"`
	Index     *int                   `json:"index,omitempty"`
	Timestamp *int64                 `json:"timestamp,omitempty"`
}

type wormwatcher struct {
	conn   *websocket.Conn
	ch     chan wormwatchEvent
	ctx    context.Context
	cancel context.CancelFunc
}

const (
	TypeTimeS = "timeS"
	TypeClear = "clear"
	TypeStart = "start"
	TypeQueue = "queue"
	TypePause = "pause"
)

func (m *Model) GetThreadWWHandler(threadID uint32, username string) (http.HandlerFunc, error) {
	m.tmapmu.Lock()
	defer m.tmapmu.Unlock()
	tm, ok := m.tmap[threadID]
	if !ok {
		return nil, ErrThreadDNE
	}
	if tm.full {
		return nil, ErrThreadFull
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
		ctx, cancel := context.WithCancel(context.Background())
		watcher := &wormwatcher{
			ch:     make(chan wormwatchEvent, 10),
			conn:   conn,
			ctx:    ctx,
			cancel: cancel,
		}
		tm.wormwatchersmu.Lock()
		tm.wormwatchers[watcher] = true
		wwd := tm.wormwatchdata
		wwd.watchers[username] = wwd.watchers[username] + 1
		tnowms := time.Now().UnixMilli()
		watcher.ch <- wormwatchEvent{Type: TypeTimeS, Timestamp: &tnowms}
		if wwd.queue != nil {
			watcher.ch <- wormwatchEvent{Type: TypeQueue, Entries: wwd.queue}
		}
		if wwd.start != nil && wwd.pausedAt == nil {
			tstartms := wwd.start.UnixMilli()
			watcher.ch <- wormwatchEvent{Type: TypeStart, Timestamp: &tstartms, Index: &wwd.index}
		}
		tm.wormwatchersmu.Unlock()

		go watcher.readloop()
		watcher.watch()

		tm.wormwatchersmu.Lock()
		delete(tm.wormwatchers, watcher)
		count := wwd.watchers[username] - 1
		if count <= 0 {
			delete(tm.wormwatchdata.watchers, username)
			// should recheck skip / pause condition here if i want to be really cool
			// but that's a pain since locks

		} else {
			tm.wormwatchdata.watchers[username] = count
		}
		tm.wormwatchersmu.Unlock()
	}, nil
}

func (m *Model) Queue(threadID uint32, username string, pis []*utils.PlayInput) {
	log.Println("add video to queue")
	if len(pis) == 0 {
		return
	}
	m.tmapmu.Lock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.Unlock()
	if !ok {
		return
	}
	wwd := tm.wormwatchdata
	// not totally ideal to allocate and copy all this while i hold lock,
	// however otherwise there's a race condition on two queue inserts at once
	tm.wormwatchersmu.Lock()
	clen := len(wwd.queue)
	entries := make([]*wormwatchqueueentry, 0, len(pis))
	initial := clen == 0
	tstart := time.Now().Add(bufferDuration)
	tstartms := tstart.UnixMilli()
	for i, pi := range pis {
		if pi == nil {
			continue
		}
		entries = append(entries, &wormwatchqueueentry{
			username: username,
			Data:     *pi,
			Index:    clen + i,
		})
	}
	log.Println(len(tm.wormwatchers))
	for w := range tm.wormwatchers {
		select {
		case w.ch <- wormwatchEvent{Type: TypeQueue, Entries: entries}:
			if initial {
				select {
				case w.ch <- wormwatchEvent{Type: TypeStart, Index: &clen, Timestamp: &tstartms}:
				default:
					log.Println("need to increase size of wormwatchers ch buffer...")
				}
			}
		default:
			w.cancel()
		}
	}
	wwd.queue = append(wwd.queue, entries...)
	if initial {
		wwd.index = 0
		wwd.start = &tstart
		ctx, cancel := context.WithCancelCause(context.Background())
		time.AfterFunc(entries[0].Data.Duration+bufferDuration, func() {
			cancel(context.DeadlineExceeded)
		})
		wwd.ctx = ctx
		wwd.cancel = cancel
		go tm.handleQueue()
	}
	tm.wormwatchersmu.Unlock()
}

func (tm *threadModel) handleQueue() {
	log.Println("queueing")
	<-tm.wormwatchdata.ctx.Done()
	log.Println("queue state shift")
	switch context.Cause(tm.wormwatchdata.ctx) {
	case context.DeadlineExceeded:
		tm.nextVideo()
	case skipped:
		tm.nextVideo()
	case paused:
		tm.pauseVideo()
	}
}

func (tm *threadModel) pauseVideo() {
	log.Println("queue pause video")
	wwd := tm.wormwatchdata
	tm.wormwatchersmu.Lock()
	tpause := time.Now().Sub(*wwd.start)
	tpausems := tpause.Milliseconds()
	wwd.pausedAt = &tpause

	for w := range tm.wormwatchers {
		select {
		case w.ch <- wormwatchEvent{Type: TypePause, Timestamp: &tpausems}:
		default:
		}
	}
	tm.wormwatchersmu.Unlock()
	log.Println("dequeueing")
}

func (tm *threadModel) nextVideo() {
	wwd := tm.wormwatchdata
	log.Println("queue next video")
	tm.wormwatchersmu.Lock()
	idx := wwd.index + 1
	if idx >= len(wwd.queue) {
		log.Println("actually, queue clear video")
		for w := range tm.wormwatchers {
			select {
			case w.ch <- wormwatchEvent{Type: TypeClear}:
			default:
			}
		}
		wwd.index = 0
		wwd.start = nil
		wwd.pausedAt = nil
		wwd.ctx = nil
		wwd.cancel = nil
		wwd.queue = nil
		tm.wormwatchersmu.Unlock()
		return
	}
	tstart := time.Now().Add(bufferDuration)
	tstartms := tstart.UnixMilli()
	wwd.index = idx
	wwd.start = &tstart
	for w := range tm.wormwatchers {
		select {
		case w.ch <- wormwatchEvent{Type: TypeStart, Index: &idx, Timestamp: &tstartms}:
		default:
		}
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	time.AfterFunc(wwd.queue[idx].Data.Duration+bufferDuration, func() {
		cancel(context.DeadlineExceeded)
	})
	wwd.ctx = ctx
	wwd.cancel = cancel
	tm.wormwatchersmu.Unlock()
	go tm.handleQueue()
}

func (m *Model) Pause(threadID uint32, username string) {
	log.Println("pause video in queue")
	m.tmapmu.Lock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.Unlock()
	wwd := tm.wormwatchdata
	if !ok {
		return
	}
	tm.wormwatchersmu.Lock()
	if wwd.start == nil || wwd.pausedAt != nil {
		tm.wormwatchersmu.Unlock()
		return
	}

	entry := wwd.queue[wwd.index]
	if entry.username == username {
		// i can always pause my own wormwatch entry
		wwd.cancel(paused)
	} else {
		// if its someone elses video, we have to check the number of votes
		if entry.votepause == nil {
			entry.votepause = make(map[string]bool)
		}
		entry.votepause[username] = true
		nvotes := len(entry.votepause)
		nwatchers := len(wwd.watchers)
		if float64(nvotes) >= math.Log2(float64(nwatchers)) {
			entry.votepause = make(map[string]bool)
			wwd.cancel(paused)
		}
	}
	tm.wormwatchersmu.Unlock()
}

func (m *Model) Skip(threadID uint32, username string) {
	log.Println("skip video in queue")
	m.tmapmu.Lock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.Unlock()
	if !ok {
		return
	}
	wwd := tm.wormwatchdata
	tm.wormwatchersmu.Lock()
	if wwd.start == nil {
		tm.wormwatchersmu.Unlock()
		return
	}

	entry := wwd.queue[wwd.index]
	if entry.username == username {
		// i can always skip my own wormwatch entry
		wwd.cancel(skipped)
	} else {
		// if its someone elses video, we have to check the number of votes
		if entry.voteskip == nil {
			entry.voteskip = make(map[string]bool)
		}
		entry.voteskip[username] = true
		nvotes := len(entry.voteskip)
		nwatchers := len(wwd.watchers)
		if float64(nvotes) >= math.Log2(float64(nwatchers)) {
			entry.voteskip = make(map[string]bool)
			wwd.cancel(skipped)
		}
	}
	tm.wormwatchersmu.Unlock()
}

func (m *Model) Unpause(threadID uint32, username string) {
	log.Println("unpause video in queue")
	m.tmapmu.Lock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.Unlock()
	if !ok {
		return
	}
	wwd := tm.wormwatchdata
	tm.wormwatchersmu.Lock()
	if wwd.start == nil || wwd.pausedAt == nil {
		tm.wormwatchersmu.Unlock()
		return
	}

	entry := wwd.queue[wwd.index]
	if entry.username == username {
		// i can always unpause my own wormwatch entry
		entry.votepause = make(map[string]bool)
		tm.unpauseHOLDINGLOCK()
		return
	} else {
		// if its someone elses video, we have to check the number of votes
		if entry.votepause == nil {
			entry.votepause = make(map[string]bool)
		}
		entry.votepause[username] = true
		nvotes := len(entry.votepause)
		nwatchers := len(wwd.watchers)
		if float64(nvotes) >= math.Log2(float64(nwatchers)) {
			entry.votepause = make(map[string]bool)
			tm.unpauseHOLDINGLOCK()
			return
		}
	}
	tm.wormwatchersmu.Unlock()
}

func (tm *threadModel) unpauseHOLDINGLOCK() {
	wwd := tm.wormwatchdata
	index := wwd.index
	tstart := time.Now().Add(-1 * *wwd.pausedAt)
	tstartms := tstart.UnixMilli()
	wwd.start = &tstart

	for w := range tm.wormwatchers {
		select {
		case w.ch <- wormwatchEvent{Type: TypeStart, Index: &index, Timestamp: &tstartms}:
		default:
		}
	}
	idx := wwd.index
	entry := wwd.queue[idx]
	entry.votepause = make(map[string]bool)
	ctx, cancel := context.WithCancelCause(context.Background())
	//paused at is the point in the video where we paused,
	time.AfterFunc(wwd.queue[idx].Data.Duration-*wwd.pausedAt, func() {
		cancel(context.DeadlineExceeded)
	})
	wwd.ctx = ctx
	wwd.cancel = cancel
	wwd.start = &tstart
	wwd.pausedAt = nil
	tm.wormwatchersmu.Unlock()
	go tm.handleQueue()
}

// watch writes events to a watcher's threadsocket
func (w *wormwatcher) watch() {
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

func (w *wormwatcher) readloop() {
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
