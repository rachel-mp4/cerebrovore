package model

import (
	"context"
	"math"
	"time"

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
	watchers map[string]int // keys are usernames
}

type wormwatchqueueentry struct {
	username  string
	voteskip  map[string]bool // keys are usernames
	votepause map[string]bool // keys are usernames
	Data      utils.PlayInput `json:"data"`
	Index     int             `json:"index"`
}

type wormwatchEvent struct {
	Type      string                 `json:"type"`
	Entries   []*wormwatchqueueentry `json:"entries,omitempty"`
	Index     *int                   `json:"index,omitempty"`
	Timestamp *int64                 `json:"timestamp,omitempty"`
}

type wormwatchMessage struct {
	Type string         `json:"type"`
	Data wormwatchEvent `json:"data"`
}

const (
	TypeTimeS = "timeS"
	TypeClear = "clear"
	TypeStart = "start"
	TypeQueue = "queue"
	TypePause = "pause"
)

func (m *Model) Queue(threadID uint32, username string, pis []*utils.PlayInput) {
	if len(pis) == 0 {
		return
	}
	m.tmapmu.RLock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	if tm.dead {
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
	for w := range tm.wormwatchers {
		select {
		case w.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypeQueue, Entries: entries}}:
			if initial {
				select {
				case w.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypeStart, Index: &clen, Timestamp: &tstartms}}:
				default:
				}
			}
		default:
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
	<-tm.wormwatchdata.ctx.Done()
	switch context.Cause(tm.wormwatchdata.ctx) {
	case context.DeadlineExceeded:
		tm.nextVideo()
	case skipped:
		tm.nextVideo()
	case paused:
		tm.pauseVideo()
	case deleted:
		return
	}
}

func (tm *threadModel) pauseVideo() {
	wwd := tm.wormwatchdata
	tm.wormwatchersmu.Lock()
	tpause := time.Since(*wwd.start)
	tpausems := tpause.Milliseconds()
	wwd.pausedAt = &tpause

	for w := range tm.wormwatchers {
		select {
		case w.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypePause, Timestamp: &tpausems, Index: &wwd.index}}:
		default:
		}
	}
	tm.wormwatchersmu.Unlock()
}

func (tm *threadModel) nextVideo() {
	tm.wormwatchersmu.Lock()
	tm.nv()
}

// nv plays the nextVideo in queue, or clears it if the queue is currently on
// the last video in queue. THIS ASSUMES WE HAVE THE LOCK
func (tm *threadModel) nv() {
	wwd := tm.wormwatchdata
	idx := wwd.index + 1
	if idx >= len(wwd.queue) {
		for w := range tm.wormwatchers {
			select {
			case w.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypeClear}}:
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
	wwd.pausedAt = nil
	for w := range tm.wormwatchers {
		select {
		case w.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypeStart, Index: &idx, Timestamp: &tstartms}}:
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
	m.tmapmu.RLock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	if tm.dead {
		return
	}
	wwd := tm.wormwatchdata
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
	m.tmapmu.RLock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	if tm.dead {
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
		if wwd.pausedAt == nil {
			wwd.cancel(skipped)
		} else {
			tm.nv()
			return
		}
	} else {
		// if its someone elses video, we have to check the number of votes
		if entry.voteskip == nil {
			entry.voteskip = make(map[string]bool)
		}
		entry.voteskip[username] = true
		nvotes := len(entry.voteskip)
		nwatchers := len(wwd.watchers)
		if float64(nvotes) >= math.Log2(float64(nwatchers)) {
			if wwd.pausedAt == nil {
				wwd.cancel(skipped)
			} else {
				tm.nv()
				return
			}
		}
	}
	tm.wormwatchersmu.Unlock()
}

func (m *Model) Unpause(threadID uint32, username string) {
	m.tmapmu.RLock()
	tm, ok := m.tmap[threadID]
	m.tmapmu.RUnlock()
	if !ok {
		return
	}
	if tm.dead {
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
		case w.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypeStart, Index: &index, Timestamp: &tstartms}}:
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
