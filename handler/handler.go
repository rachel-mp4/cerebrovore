package handler

import (
	"crypto/rand"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/rachel-mp4/cerebrovore/db"
	"github.com/rachel-mp4/cerebrovore/model"
)

type Handler struct {
	mux          *http.ServeMux
	ca           *CompiledAssets
	m            *model.Model
	sessionStore *sessions.CookieStore
	db           db.Storer
	idp          db.IDStorer
}

type CompiledAssets struct {
	ChatPath    string
	ChatCss     []string
	BeepPath    string
	BeepCss     []string
	WatcherPath string
	WatcherCss  []string
	WormPath    string
	WormCss     []string
}

func NewHandler(ca *CompiledAssets, m *model.Model, db db.Storer, idp db.IDStorer) Handler {
	h := Handler{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.AM(h.home))
	// mux.HandleFunc("GET /catalog", h.AM(h.catalog))
	mux.HandleFunc("GET /me", h.AM(h.me))
	mux.HandleFunc("POST /logout", h.logout)
	mux.HandleFunc("GET /login", h.login)
	mux.HandleFunc("POST /login", h.postLogin)
	mux.HandleFunc("GET /account", h.account)
	mux.HandleFunc("POST /account", h.postAccount)
	mux.HandleFunc("GET /beep", h.beep)
	mux.HandleFunc("GET /ts", h.AM(h.getThreadSocket))
	mux.HandleFunc("GET /t-bumped", h.AM(h.getTBumped))
	mux.HandleFunc("POST /t", h.AM(h.postThread))
	mux.HandleFunc("GET /t", h.AM(h.threads))
	mux.HandleFunc("POST /blob", h.AM(h.postBlob))
	mux.HandleFunc("GET /blob", h.AM(h.getBlob))
	mux.HandleFunc("POST /t/{ntid}", h.AM(h.postPost))
	mux.HandleFunc("GET /t/{ntid}", h.AM(h.getThread))
	mux.HandleFunc("GET /p/{npid}", h.AM(h.getPost))
	mux.HandleFunc("POST /w/{ntid}", h.AM(h.watchThread))
	mux.HandleFunc("POST /u/{ntid}", h.AM(h.unwatchThread))
	mux.HandleFunc("GET /t/{ntid}/ws", h.AM(h.getThreadWS))
	mux.HandleFunc("GET /t/{ntid}/ww", h.AM(h.getThreadWW))
	mux.HandleFunc("GET /new", h.AM(h.newThread))
	mux.Handle("GET /css/", http.FileServer(http.Dir("./static")))
	mux.Handle("GET /js/", http.FileServer(http.Dir("./static")))
	mux.Handle("GET /font/", Add1YCache(http.FileServer(http.Dir("./static"))))
	mux.Handle("GET /svg/", http.FileServer(http.Dir("./static")))
	mux.Handle("GET /assets/", http.FileServer(http.Dir("./frontend/dist")))
	h.mux = mux
	h.ca = ca
	h.m = m
	sessionStore := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
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
		baseresp
	}
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting threads", http.StatusInternalServerError)
		return
	}
	err = homeT.ExecuteTemplate(w, "base", homeresp{
		baseresp{
			h.ca,
			"brainworm",
			tt,
		},
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
	type ntresp struct {
		baseresp
	}
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting threads", http.StatusInternalServerError)
		return
	}
	err = newthreadT.ExecuteTemplate(w, "base", ntresp{
		baseresp{
			h.ca,
			"new thread",
			tt,
		},
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	type loginresp struct {
		baseresp
	}
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting threads", http.StatusInternalServerError)
		return
	}
	err = loginT.ExecuteTemplate(w, "base", loginresp{
		baseresp{
			h.ca,
			"login",
			tt,
		},
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}
}

func (h *Handler) account(w http.ResponseWriter, r *http.Request) {
	type loginresp struct {
		baseresp
	}
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting threads", http.StatusInternalServerError)
		return
	}
	err = accountT.ExecuteTemplate(w, "base", loginresp{
		baseresp{
			h.ca,
			"create account",
			tt,
		},
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}
}

func (h *Handler) postAccount(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	confirmPassword := r.Form.Get("confirm")
	invite := r.Form.Get("invite")
	if username == "" {
		http.Error(w, "enter a username", http.StatusBadRequest)
		return
	}
	if password != confirmPassword {
		http.Error(w, "password must match confirm password", http.StatusBadRequest)
		return
	}
	if h.idp != nil {
		err := h.idp.CreateAccount(username, password, invite)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	id := rand.Text()
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

func (h *Handler) postLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	id := rand.Text()
	if h.idp != nil {
		err := h.idp.VerifyCredentials(username, password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
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

func (h *Handler) beep(w http.ResponseWriter, r *http.Request) {
	type beepresp struct {
		baseresp
	}
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting threads", http.StatusInternalServerError)
	}
	beepT.ExecuteTemplate(w, "base", beepresp{
		baseresp{
			h.ca,
			"beep",
			tt,
		},
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

func (h *Handler) me(c *Client, w http.ResponseWriter, r *http.Request) {
	type meresp struct {
		baseresp
		Username string
	}
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting threads", http.StatusInternalServerError)
	}
	meT.ExecuteTemplate(w, "base", meresp{
		baseresp{
			h.ca,
			"beep",
			tt,
		},
		c.Username,
	})

}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	s, _ := h.sessionStore.Get(r, "session")
	s.Options.MaxAge = -1
	s.Save(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// AM is an auth middleware function that reads from our cookiestore and validates
// it with the database
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
		whynotcheck, err := h.db.RetrieveSession(id, r.Context())
		if err != nil || username != whynotcheck {
			s.Options.MaxAge = -1
			s.Save(r, w)
			f(nil, w, r)
			return
		}
		c := &Client{ID: id, Username: username}
		f(c, w, r)
	}
}

// Add1YCache is a middleware function to add a header to cache our
// response for up to a year. only use this on things that definitely
// never will change, like font or maybe blobs
func Add1YCache(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		h.ServeHTTP(w, r)
	})
}
