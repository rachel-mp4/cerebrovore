package handler

import (
	"log"
	"net/http"
	"os"

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
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting threads", http.StatusInternalServerError)
		return
	}
	err = moderateT.ExecuteTemplate(w, "base", moderateresp{
		baseresp{
			h.ca,
			"moderation",
			tt,
		},
	})
	if err != nil {
		log.Println(err)
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
		log.Println(nid)
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
