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
	noseek   bool
	voteseek map[string]bool
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

func (m *Model) NoSeek(threadID uint32, username string) {
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
	defer tm.wormwatchersmu.Unlock()
	if wwd.noseek {
		delete(wwd.voteseek, username)
		return
	}
	wwd.voteseek[username] = true
	nvotes := len(wwd.voteseek)
	nwatchers := len(wwd.watchers)
	if float64(nvotes) >= math.Log2(float64(nwatchers)) {
		wwd.noseek = true
		wwd.voteseek = make(map[string]bool)
	}
}

func (m *Model) YesSeek(threadID uint32, username string) {
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
	defer tm.wormwatchersmu.Unlock()
	if !wwd.noseek {
		delete(wwd.voteseek, username)
		return
	}
	wwd.voteseek[username] = true
	nvotes := len(wwd.voteseek)
	nwatchers := len(wwd.watchers)
	if float64(nvotes) >= math.Log2(float64(nwatchers)) {
		wwd.noseek = false
		wwd.voteseek = make(map[string]bool)
	}
}

// basically the idea with seeking is that it's time sensitive, so it doesn't
// require voting, instead to help spam the thread can either enter a noseek
// mode where only the queuer is able to seek. determining whether thread should
// be yesseek or noseek is what requires a vote. if you input a seek command
// but you didn't queue the current video and the thread is in noseek mode, we
// add you as a yesseek vote, and if that takes the thread into yesseek mode, your
// vote goes through
func (m *Model) Seek(threadID uint32, username string, to time.Duration, rel bool) {
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
	if wwd.index >= len(wwd.queue) {
		tm.wormwatchersmu.Unlock()
		return
	}
	playing := wwd.queue[wwd.index]
	if playing == nil {
		tm.wormwatchersmu.Unlock()
		return
	}
	if wwd.noseek && playing.username != username {
		wwd.voteseek[username] = true
		nvotes := len(wwd.voteseek)
		nwatchers := len(wwd.watchers)
		if float64(nvotes) >= math.Log2(float64(nwatchers)) {
			wwd.noseek = false
			wwd.voteseek = make(map[string]bool)
		} else {
			tm.wormwatchersmu.Unlock()
			return
		}
	}
	// 4 cases: paused,  seek out of bounds: manually invoke nv (comment@blinko blitzer)
	//          playing, seek out of bounds: skip
	//          paused,  seek in bounds: set paused at to new time
	//          playing, seek in bounds: send typeStart event with new time
	ntr := to + time.Since(*wwd.start)
	seekoutofbounds := (!rel && to > (playing.Data.Duration+bufferDuration)) ||
		(rel && wwd.pausedAt != nil && (to+*wwd.pausedAt) > (playing.Data.Duration+bufferDuration)) ||
		(rel && wwd.pausedAt == nil && ntr > (playing.Data.Duration+bufferDuration))
	if seekoutofbounds {
		if wwd.pausedAt == nil {
			wwd.cancel(skipped)
		} else {
			tm.nv()
			return
		}
	} else {
		if wwd.pausedAt != nil {
			if !rel {
				wwd.pausedAt = &to
			} else {
				ntr = to + *wwd.pausedAt
				wwd.pausedAt = &ntr
			}
		} else {
			wwd.cancel(sought)
			if rel {
				to = ntr
			}
			tstart := time.Now().Add(-1 * to)
			otheroption := time.Now().Add(bufferDuration)
			if tstart.After(otheroption) {
				tstart = otheroption
			}
			tstartms := tstart.UnixMilli()

			for w := range tm.wormwatchers {
				select {
				case w.ch <- wormwatchMessage{"wormwatch", wormwatchEvent{Type: TypeStart, Index: &wwd.index, Timestamp: &tstartms}}:
				default:
				}
			}
			idx := wwd.index
			ctx, cancel := context.WithCancelCause(context.Background())
			time.AfterFunc(wwd.queue[idx].Data.Duration-to, func() {
				cancel(context.DeadlineExceeded)
			})
			wwd.ctx = ctx
			wwd.cancel = cancel
			wwd.start = &tstart
			go tm.handleQueue()
		}
	}
	tm.wormwatchersmu.Unlock()

}

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
	// since i label entries with their indices
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
	case sought, deleted:
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

		// blinko blitzer
		// when the video is paused, there is no routine handling the queue,
		// so we can't cancel it (in fact cancel should be a no-op or something)
		// recall we have a kinda dumb recursive go routine structure here as
		// the video is playing, but it seemed not necessary to have someone waiting
		// while the video was paused since pauses can theoretically last forever
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
