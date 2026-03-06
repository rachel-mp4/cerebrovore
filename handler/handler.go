package handler

import (
	"github.com/gorilla/sessions"
	"github.com/rachel-mp4/cerebrovore/db"
	"github.com/rachel-mp4/cerebrovore/model"
	"github.com/rachel-mp4/cerebrovore/types"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	mux          *http.ServeMux
	ca           *CompiledAssets
	m            *model.Model
	sessionStore *sessions.CookieStore
	db           db.Storer
	idp          withIdentityProvider
}

type CompiledAssets struct {
	ChatPath    string
	ChatCss     []string
	BeepPath    string
	BeepCss     []string
	WatcherPath string
	WatcherCss  []string
}

type withIdentityProvider = bool

func NewHandler(ca *CompiledAssets, m *model.Model, db db.Storer, idp withIdentityProvider) Handler {
	h := Handler{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.AM(h.home))
	mux.HandleFunc("GET /login", h.login)
	mux.HandleFunc("POST /callback", h.callback)
	mux.HandleFunc("GET /beep", h.beep)
	mux.HandleFunc("GET /ts", h.AM(h.getThreadSocket))
	mux.HandleFunc("GET /t-bumped", h.AM(h.getTBumped))
	mux.HandleFunc("POST /t", h.AM(h.postThread))
	mux.HandleFunc("POST /blob", h.AM(h.postBlob))
	mux.HandleFunc("GET /blob", h.AM(h.getBlob))
	mux.HandleFunc("POST /t/{ntid}", h.AM(h.postPost))
	mux.HandleFunc("GET /t/{ntid}", h.AM(h.getThread))
	mux.HandleFunc("GET /p/{npid}", h.AM(h.getPost))
	mux.HandleFunc("POST /w/{ntid}", h.AM(h.watchThread))
	mux.HandleFunc("POST /u/{ntid}", h.AM(h.unwatchThread))
	mux.HandleFunc("GET /t/{ntid}/ws", h.AM(h.getThreadWS))
	mux.HandleFunc("GET /new", h.AM(h.newThread))
	mux.Handle("GET /css/", http.FileServer(http.Dir("./static")))
	mux.Handle("GET /js/", http.FileServer(http.Dir("./static")))
	mux.Handle("GET /font/", Add1YCache(http.FileServer(http.Dir("./static"))))
	mux.Handle("GET /svg/", http.FileServer(http.Dir("./static")))
	mux.Handle("GET /assets/", http.FileServer(http.Dir("./frontend/dist")))
	h.mux = mux
	h.ca = ca
	h.m = m
	sessionStore := sessions.NewCookieStore([]byte("//TODO: FIX ME AAAAAAA"))
	h.sessionStore = sessionStore
	h.db = db
	h.idp = idp

	return h
}

func (h *Handler) home(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.URL.Path != "/" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	type homeresp struct {
		Title          string
		ThreadsCursor  *time.Time
		Threads        []types.Thread
		CompiledAssets *CompiledAssets
	}
	tt, crsr, err := h.db.GetBumpedThreads(nil, 10, r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting threads", http.StatusInternalServerError)
	}
	err = homeT.ExecuteTemplate(w, "base", homeresp{"brainworm",
		crsr, tt, h.ca,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}
}

func (h *Handler) newThread(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	type homeresp struct {
		Title          string
		ThreadsCursor  *time.Time
		Threads        []types.Thread
		CompiledAssets *CompiledAssets
	}
	tt, crsr, err := h.db.GetBumpedThreads(nil, 10, r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting threads", http.StatusInternalServerError)
	}
	err = newthreadT.ExecuteTemplate(w, "base", homeresp{"new thread",
		crsr, tt, h.ca,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if h.idp == false {
		type loginresp struct {
			Title          string
			ThreadsCursor  *time.Time
			Threads        []types.Thread
			CompiledAssets *CompiledAssets
		}
		tt, crsr, err := h.db.GetBumpedThreads(nil, 10, r.Context())
		if err != nil {
			log.Println(err)
			http.Error(w, "error getting threads", http.StatusInternalServerError)
			return
		}
		err = mockloginT.ExecuteTemplate(w, "base", loginresp{"login",
			crsr, tt, h.ca,
		})
		if err != nil {
			log.Println(err)
			http.Error(w, "error templating", http.StatusInternalServerError)
		}
	}
}

func (h *Handler) callback(w http.ResponseWriter, r *http.Request) {
	if h.idp == false {
		r.ParseForm()
		username := r.Form.Get("username")
		id := r.Form.Get("id")
		session, _ := h.sessionStore.Get(r, "session")
		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
		}
		session.Values = map[any]any{}
		session.Values["username"] = username
		session.Values["id"] = id
		err := session.Save(r, w)
		if err != nil {
			log.Println(err)
			http.Error(w, "error saving session", http.StatusInternalServerError)
			return
		}
		h.db.CreateSession(id, username, r.Context())
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (h *Handler) beep(w http.ResponseWriter, r *http.Request) {
	type beepresp struct {
		Title          string
		CompiledAssets *CompiledAssets
		ThreadsCursor  *time.Time
		Threads        []types.Thread
	}
	tt, crsr, err := h.db.GetBumpedThreads(nil, 10, r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting threads", http.StatusInternalServerError)
	}
	beepT.ExecuteTemplate(w, "base", beepresp{"beep",
		h.ca,
		crsr,
		tt,
	})
}

func (h *Handler) Serve() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.mux.ServeHTTP(w, r)
	})
}

type Client struct {
	ID       string
	Username string
}

// AM is an auth middleware function that reads from our cookiestore
func (h *Handler) AM(f func(c *Client, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s, _ := h.sessionStore.Get(r, "session")
		id, ok := s.Values["id"].(string)
		username, bok := s.Values["username"].(string)
		if !ok || !bok {
			f(nil, w, r)
			return
		}
		// check that they haven't been logged out
		_, err := h.db.RetrieveSession(id, r.Context())
		if err != nil {
			s.Options.MaxAge = -1
			s.Save(r, w)
			f(nil, w, r)
			return
		}
		c := &Client{ID: id, Username: username}
		f(c, w, r)
	}
}

func Add1YCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		h.ServeHTTP(w, r)
	})

}
