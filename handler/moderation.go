package handler

import (
	"net/http"
	"os"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/utils"
)

func (h *Handler) moderate(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || c.Username != os.Getenv("ADMIN_USERNAME") {
		http.Error(w, "not authorized to moderate", http.StatusUnauthorized)
		return
	}
	type moderateresp struct {
		baseresp
	}
	base, err := h.makebase("moderation", r.Context())
	if err != nil {
		clog.Warn("bumps %s", err)
	}
	err = moderateT.ExecuteTemplate(w, "base", moderateresp{
		*base,
	})
	if err != nil {
		clog.Warn("moderation: %s", err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}
}

func (h *Handler) postModerate(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || c.Username != os.Getenv("ADMIN_USERNAME") {
		http.Error(w, "not authorized to moderate", http.StatusUnauthorized)
		return
	}
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}
	nid := r.FormValue("postid")
	id, err := utils.AToID(nid)
	if err != nil {
		clog.Warn("moderation: invalid id %s", nid)
		http.Error(w, "id should be correct!", http.StatusBadRequest)
		return
	}
	thread := r.FormValue("thread")
	isThread := thread != ""
	if isThread {
		err = h.db.DeleteThread(id, r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		err = h.db.DeletePost(id, r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
