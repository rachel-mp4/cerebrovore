package handler

import (
	"crypto/rand"
	"errors"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/sessions"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/db"
	"github.com/rachel-mp4/cerebrovore/id"
	"github.com/rachel-mp4/cerebrovore/model"
	"github.com/rachel-mp4/cerebrovore/types"
)

type Handler struct {
	mux          *http.ServeMux
	ca           *CompiledAssets
	m            *model.Model
	sessionStore *sessions.CookieStore
	db           db.Storer
	idp          id.Provider

	// crack is a string we append to static assets that we want clients to
	// redownload
	crack string

	// notes is the list of patch notes, for flavor
	notes []types.Patch

	// live is the time that the site went live last, just for flavor
	live *time.Time

	// reqcode is true if the current id provider requires invite codes for account registration
	reqcode bool

	commit string
}

type CompiledAssets struct {
	ChatPath     string
	ChatCss      []string
	BeepPath     string
	BeepCss      []string
	WatcherPath  string
	WatcherCss   []string
	WormPath     string
	WormCss      []string
	SettingsPath string
	SettingsCss  []string
}

func NewHandler(ca *CompiledAssets, m *model.Model, db db.Storer, idp id.Provider, reqcode bool) Handler {
	h := Handler{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.AM(h.home))
	mux.HandleFunc("GET /catalog", h.AM(h.catalog))
	mux.HandleFunc("GET /patch-notes", h.AM(h.patchnotes))
	mux.HandleFunc("GET /settings", h.AM(h.me))
	mux.HandleFunc("GET /profile/{username}", h.AM(h.profile))
	mux.HandleFunc("POST /profile", h.AM(h.postProfile))
	mux.HandleFunc("GET /profile", h.AM(h.editProfile))
	mux.HandleFunc("POST /avatar", h.AM(h.postAvatar))
	mux.HandleFunc("POST /profile-contents", h.AM(h.postContents))
	mux.HandleFunc("GET /moderate", h.AM(h.moderate))
	mux.HandleFunc("GET /administrate", h.AM(h.administrate))
	mux.HandleFunc("POST /add-moderator", h.AM(h.addModerator))
	mux.HandleFunc("POST /take-action", h.AM(h.takeAction))
	mux.HandleFunc("GET /cancel-action", h.AM(h.cancelAction))
	mux.HandleFunc("GET /inspect-post", h.AM(h.inspectPost))
	mux.HandleFunc("POST /appeal-verdict", h.AM(h.postAppealVerdict))
	mux.HandleFunc("POST /report", h.AM(h.postReport))
	mux.HandleFunc("POST /review-report", h.AM(h.reviewReport))
	mux.HandleFunc("GET /reports", h.AM(h.getReports))
	mux.HandleFunc("POST /logout", h.logout)
	mux.HandleFunc("GET /login", h.login)
	mux.HandleFunc("POST /login", h.postLogin)
	mux.HandleFunc("GET /account", h.account)
	mux.HandleFunc("POST /account", h.postAccount)
	mux.HandleFunc("POST /appeal", h.postAppeal)
	mux.HandleFunc("GET /beep", h.AM(h.beep))
	mux.HandleFunc("GET /t-bumped", h.AM(h.getTBumped))
	mux.HandleFunc("POST /t", h.AM(h.postThread))
	mux.HandleFunc("GET /t", h.AM(h.threads))
	mux.HandleFunc("GET /ft", h.AM(h.forumthreads))
	mux.HandleFunc("POST /blob", h.AM(h.postBlob))
	mux.HandleFunc("GET /blob", h.AM(h.getBlob))
	mux.HandleFunc("POST /t/{ntid}", h.AM(h.postPost))
	mux.HandleFunc("POST /ft/{ntid}", h.AM(h.postForumPost))
	mux.HandleFunc("GET /fp/{npid}", h.AM(h.forumPost))
	mux.HandleFunc("GET /t/{ntid}", h.AM(h.getThread))
	mux.HandleFunc("GET /ft/{ntid}", h.AM(h.getForumThread))
	mux.HandleFunc("GET /p/{npid}", h.AM(h.getPost))
	mux.HandleFunc("DELETE /p/{npid}", h.AM(h.deletePost))
	mux.HandleFunc("POST /w/{ntid}", h.AM(h.watchThread))
	mux.HandleFunc("POST /u/{ntid}", h.AM(h.unwatchThread))
	mux.HandleFunc("GET /t/{ntid}/ws", h.AM(h.getThreadWS))
	mux.HandleFunc("GET /ws", h.AM(h.getJSONWebsockets))
	mux.HandleFunc("GET /new", h.AM(h.newThread))
	mux.HandleFunc("POST /gen-code", h.AM(h.gencode))
	mux.HandleFunc("POST /gen-public-code", h.AM(h.genpubliccode))
	mux.Handle("GET /css/", h.StripCrack(http.FileServer(http.Dir("./static"))))
	mux.Handle("GET /js/", h.StripCrack(http.FileServer(http.Dir("./static"))))
	mux.Handle("GET /font/", Add1YCache(http.FileServer(http.Dir("./static"))))
	mux.Handle("GET /wav/", Add1YCache(http.FileServer(http.Dir("./static"))))
	mux.Handle("GET /svg/", http.FileServer(http.Dir("./static")))
	mux.Handle("GET /assets/", http.FileServer(http.Dir("./frontend/dist")))
	h.mux = mux
	h.ca = ca
	h.m = m
	sessionStore := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	h.sessionStore = sessionStore
	h.db = db
	h.idp = idp
	if h.idp == nil {
		panic("you shouldn't be allowed to do that anymore")
	}
	h.crack = "-" + string(rand.Text()[:5])
	h.notes = getNotes()
	t := time.Now()
	h.live = &t
	h.reqcode = reqcode
	out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		panic(err)
	}
	h.commit = strings.TrimSpace(string(out))

	return h
}

func (h *Handler) gencode(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		return
	}
	code, err := h.idp.GenerateCode(c.Username, r.Context())
	if err != nil {
		codeerrT.exec(w, codeerrresp{err.Error()})
		return
	}
	codeT.exec(w, coderesp{code})
}

func (h *Handler) genpubliccode(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		return
	}
	code, err := h.idp.GeneratePublicCode(c.Username, r.Context())
	if err != nil {
		codeerrT.exec(w, codeerrresp{err.Error()})
		return
	}
	codeT.exec(w, coderesp{code})
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
	base, err := h.makebase("brainworm", c, r.Context())
	if err != nil {
		clog.Warn("getbumps %s", err)
	}
	homeT.exec(w, homeresp{
		base,
		h.notes[0].Release,
		h.live,
		h.commit,
		os.Getenv("DISCORD_LINK"),
	})
}

func (h *Handler) newThread(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	base, err := h.makebase("new thread", c, r.Context())
	if err != nil {
		clog.Warn("bumps %s", err)
	}
	newthreadT.exec(w, newthreadresp{
		base,
	})
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	loginT.exec(w, loginresp{
		h.makejustbase("login", false),
		os.Getenv("DISCORD_LINK"),
	})
}

func (h *Handler) account(w http.ResponseWriter, r *http.Request) {
	type loginresp struct {
	}
	invite := r.URL.Query().Get("invite")
	accountT.exec(w, accountresp{
		h.makejustbase("create account", false),
		invite,
		h.reqcode,
		os.Getenv("DISCORD_LINK"),
	})
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
		err := h.idp.CreateAccount(username, password, invite, r.Context())
		if err != nil {
			if errors.Is(err, id.ErrCodeDNE) ||
				errors.Is(err, id.ErrCodeUsed) ||
				errors.Is(err, id.ErrCodeExpired) ||
				errors.Is(err, id.ErrUserExists) ||
				errors.Is(err, id.ErrBadData) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			clog.Warn("%s", err.Error())
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}
	id := rand.Text()
	session, _ := h.sessionStore.Get(r, "session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 14,
		HttpOnly: true,
	}
	session.Values = map[any]any{}
	session.Values["username"] = username
	session.Values["id"] = id
	err := session.Save(r, w)
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "error saving session", http.StatusInternalServerError)
		return
	}
	h.db.CreateSession(id, username, r.Context())
	h.db.InitializeProfile(username, r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) postLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	myid := rand.Text()
	if h.idp != nil {
		err := h.idp.VerifyCredentials(username, password, r.Context())
		if err != nil {
			if errors.Is(err, id.ErrNotAuthorized) ||
				errors.Is(err, id.ErrUserDNE) ||
				errors.Is(err, id.ErrUserBanned) ||
				errors.Is(err, id.ErrUserDeleted) ||
				errors.Is(err, id.ErrBadData) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			clog.Warn("%s", err.Error())
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}
	session, _ := h.sessionStore.Get(r, "session")
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 14,
		HttpOnly: true,
	}
	session.Values = map[any]any{}
	session.Values["username"] = username
	session.Values["id"] = myid
	err := session.Save(r, w)
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "error saving session", http.StatusInternalServerError)
		return
	}
	h.db.CreateSession(myid, username, r.Context())
	h.db.InitializeProfile(username, r.Context())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) beep(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "this is really serious.... you are not authz...", http.StatusUnauthorized)
		return
	}
	base, err := h.makebase("beep", c, r.Context())
	if err != nil {
		clog.Warn("bumps %s", err)
	}
	beepT.exec(w, beepresp{
		base,
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
	IsMod    bool
}

func (h *Handler) me(c *Client, w http.ResponseWriter, r *http.Request) {
	base, err := h.makebase("me", c, r.Context())
	if err != nil {
		clog.Warn("bumps %s", err)
	}
	meT.exec(w, meresp{
		base,
		c.Username,
		h.reqcode,
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
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// check that they haven't been logged out
		whynotcheck, err := h.db.RetrieveSession(id, r.Context())

		if err != nil || username != whynotcheck {
			s.Options.MaxAge = -1
			s.Save(r, w)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		s.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 14,
			HttpOnly: true,
		}
		s.Save(r, w)
		isadmin := os.Getenv("ADMIN_USERNAME") == username
		ismod, err := h.db.IsModerator(username, r.Context())
		if err != nil {
			clog.Info("%s", err)
			http.Error(w, "error getting if moderator", http.StatusInternalServerError)
			return
		}
		if isadmin {
			ismod = true
		}
		ban, post, err := h.db.IsBanned(username, r.Context())
		if err != nil {
			clog.Info("%s", err)
			http.Error(w, "error getting ban state", http.StatusInternalServerError)
			return
		}
		if ban != nil && !ismod {
			banT.exec(w, banresp{h.makejustbase("ban", false), false, *ban, post})
			return
		}
		c := &Client{ID: id, Username: username, IsMod: ismod}
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

// StripCrack is a middleware function to remove the cache crack string
// from a path that might have it
func (hdlr *Handler) StripCrack(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, hdlr.crack)
		h.ServeHTTP(w, r)
	})
}
