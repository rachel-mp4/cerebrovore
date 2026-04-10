package handler

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
)

func (h *Handler) moderate(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || c.Username != os.Getenv("ADMIN_USERNAME") {
		http.Error(w, "not authorized to moderate", http.StatusUnauthorized)
		return
	}
	type moderateresp struct {
		baseresp
		Appeals []types.Appeal
	}
	base, err := h.makebase("moderation", c.Username, r.Context())
	if err != nil {
		clog.Warn("bumps %s", err)
	}
	appeals, _, err := h.db.GetAppeals(10, nil, r.Context())
	if err != nil {
		clog.Dbug("%s", err)
	}
	err = moderateT.ExecuteTemplate(w, "base", moderateresp{
		*base,
		appeals,
	})
	if err != nil {
		clog.Warn("moderation: %s", err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}
}

func (h *Handler) postAppealVerdict(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || c.Username != os.Getenv("ADMIN_USERNAME") {
		http.Error(w, "not authorized to moderate", http.StatusUnauthorized)
		return
	}
	r.ParseForm()
	reject := r.Form.Get("reject")
	id := r.Form.Get("id")
	banid, err := strconv.Atoi(id)
	type errorresp struct {
		Message string
	}
	if err != nil {
		moderateT.ExecuteTemplate(w, "errored", errorresp{Message: err.Error()})
		return
	}
	if reject != "" {
		err := h.db.Reject(banid, r.Context())
		if err != nil {
			moderateT.ExecuteTemplate(w, "errored", errorresp{Message: err.Error()})
			return
		}
		moderateT.ExecuteTemplate(w, "rejected", nil)
		return
	}
	err = h.db.Unban(banid, r.Context())
	if err != nil {
		moderateT.ExecuteTemplate(w, "errored", errorresp{Message: err.Error()})
		return
	}
	moderateT.ExecuteTemplate(w, "accepted", nil)
}

func (h *Handler) postDeletePost(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || c.Username != os.Getenv("ADMIN_USERNAME") {
		http.Error(w, "not authorized to moderate", http.StatusUnauthorized)
		return
	}
	type errorresp struct {
		Message string
	}
	err := r.ParseForm()
	if err != nil {
		moderateT.ExecuteTemplate(w, "errored", errorresp{Message: err.Error()})
		return
	}
	nid := r.FormValue("postid")
	id, err := utils.AToID(nid)
	if err != nil {
		clog.Warn("moderation: invalid id %s", nid)
		moderateT.ExecuteTemplate(w, "errored", errorresp{Message: err.Error()})
		return
	}
	thread := r.FormValue("thread")
	isThread := thread != ""
	if isThread {
		err = h.db.DeleteThread(id, r.Context())
		if err != nil {
			moderateT.ExecuteTemplate(w, "errored", errorresp{Message: err.Error()})
			return
		}
		post, err := h.db.GetPost(id, r.Context())
		type resp struct {
			Post *types.Post
		}
		if err != nil {
			moderateT.ExecuteTemplate(w, "errored", errorresp{Message: err.Error()})
			return
		}

		moderateT.ExecuteTemplate(w, "thread-deleted", resp{Post: post})
	} else {
		err = h.db.DeletePost(id, r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		post, err := h.db.GetPost(id, r.Context())
		type resp struct {
			Post *types.Post
		}
		if err != nil {
			moderateT.ExecuteTemplate(w, "errored", errorresp{Message: err.Error()})
			return
		}

		moderateT.ExecuteTemplate(w, "post-deleted", resp{Post: post})
	}
}

func (h *Handler) selfban(username string, postid *uint32, reason string, til time.Time, ctx context.Context) error {
	err := h.db.SelfBan(username, postid, reason, til, ctx)
	if err == nil {
		go h.postBanCleanup(username)
	}
	return err
}

func (h *Handler) postBanUser(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || c.Username != os.Getenv("ADMIN_USERNAME") {
		http.Error(w, "not authorized to ban", http.StatusUnauthorized)
		return
	}
	r.ParseForm()
	b := types.Ban{}
	b.Username = r.Form.Get("username")
	reason := r.Form.Get("reason")
	if reason != "" {
		b.Reason = &reason
	}
	comment := r.Form.Get("comment")
	if comment != "" {
		b.Comment = &comment
	}
	duration := r.Form.Get("duration")
	if duration == "" {
		dur := time.Hour * 24 * 365 * 10
		b.Until = time.Now().Add(dur)
	} else {
		dur, err := strconv.ParseInt(duration, 10, 64)
		if err != nil {
			dur = 24 * 365 * 10
		}
		b.Until = time.Now().Add(time.Duration(dur) * time.Hour)
	}
	postId := r.Form.Get("post-id")
	if postId != "" {
		id, err := utils.AToID(postId)
		if err == nil {
			b.PostId = &id
		}
	}
	err := h.db.Ban(&b, r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(nil)
	go h.postBanCleanup(b.Username)
}

func (h *Handler) postBanCleanup(username string) {
	h.m.BanUser(username)
}

func (h *Handler) postAppeal(w http.ResponseWriter, r *http.Request) {
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
	r.ParseForm()
	banid := r.Form.Get("id")
	banidint, err := strconv.ParseInt(banid, 10, 64)
	if err != nil {
		http.Error(w, "invalid ban id", http.StatusBadRequest)
		return
	}
	response := r.Form.Get("response-field")
	err = h.db.Appeal(int(banidint), username, response, r.Context())
	if err != nil {
		clog.Info("%s", err)
		http.Error(w, "error posting appeal", http.StatusInternalServerError)
		return
	}
	banT.ExecuteTemplate(w, "appeal-submitted", nil)
}
