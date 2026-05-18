package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/rachel-mp4/cerebrovore/clog"
	"golang.org/x/time/rate"
)

// bucket of buckets
type limitStore struct {
	mu       sync.Mutex
	limiters map[string]*limitEntry // think one spell = 1 mana
	r        rate.Limit             // this is like everyone's mana regen rate
	b        int                    // think this is like everyone's max mana pool
}

type limitEntry struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

// the way how ratelimits are represented is we have a buffer of burst
// that regenerates x tokens per second.
func newLimitStore(regenPeriod time.Duration, burst int) *limitStore {
	s := &limitStore{
		limiters: make(map[string]*limitEntry),
		r:        rate.Every(regenPeriod),
		b:        burst,
	}
	go s.sweep() // clean
	return s
}

func (s *limitStore) allow(key string) bool {
	s.mu.Lock()
	e, ok := s.limiters[key]
	if !ok {
		e = &limitEntry{lim: rate.NewLimiter(s.r, s.b)}
		s.limiters[key] = e
	}
	e.lastSeen = time.Now()
	s.mu.Unlock()
	return e.lim.Allow()
}

func (s *limitStore) sweep() {
	for range time.Tick(10 * time.Minute) {
		cutoff := time.Now().Add(-1 * time.Hour)
		s.mu.Lock()
		for k, e := range s.limiters {
			if e.lastSeen.Before(cutoff) {
				delete(s.limiters, k)
			}
		}
		s.mu.Unlock()
	}
}

// completely sane programming language that is great and I love
// what if instead of writing code we just glued The Gadget to the bin
// I LVOE ERR = NIL RETURN ERR :3
func rateLimit(
	store *limitStore,
	keyFn func(c *Client, r *http.Request) string,
	f func(c *Client, w http.ResponseWriter, r *http.Request),
) func(c *Client, w http.ResponseWriter, r *http.Request) {
	return func(c *Client, w http.ResponseWriter, r *http.Request) {
		key := keyFn(c, r)
		if !store.allow(key) {
			clog.Warn("rate limit hit by @%s (%s)", c.Username, key)
			profileT.error(w, "stop im telling mom")
			return
		}
		f(c, w, r)
	}
}

func rateLimitIP(
	store *limitStore,
	f func(w http.ResponseWriter, r *http.Request),
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Real-IP")
		if key == "" {
			key = r.RemoteAddr // should always be behind nginx but :zany:
		}
		if !store.allow(key) {
			clog.Warn("IP rate limit hit (%s)", key)
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		f(w, r)
	}
}
