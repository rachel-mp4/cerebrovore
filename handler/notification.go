package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/rachel-mp4/cerebrovore/clog"
)

func (h *Handler) readinbox(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		return
	}
	err := h.db.ReadNotifications(c.Username, r.Context())
	if err != nil {
		inboxT.error(w, "read error")
		return
	}
	h.m.DispatchClear(c.Username)
	w.Write([]byte("all read"))
}

func (h *Handler) inbox(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		return
	}
	cursor := r.URL.Query().Get("cursor")
	var before *int
	if cursor != "" {
		ci, err := strconv.Atoi(cursor)
		if err == nil {
			before = &ci
		}
	}
	notifications, ncursor, includesLastRead, err := h.db.GetNotifications(c.Username, 36, before, r.Context())
	if err != nil {
		if before != nil {
			inboxT.error(w, "db error")
			return
		}
		clog.Warn("%s", err.Error())
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if before != nil {
		inboxT.notifications(w, notificationsResp{notifications, ncursor})
	} else {
		base, err := h.makebase("inbox", c, r.Context())
		if err != nil {
			clog.Warn("%s", err.Error())
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if includesLastRead {
			base.Notifications = 0
		}
		inboxT.exec(w, inboxResp{base, notifications, ncursor})
	}
	if includesLastRead {
		go func() {
			err := h.db.ReadNotifications(c.Username, context.Background())
			if err != nil {
				clog.Warn("%s", err)
				return
			}
			h.m.DispatchClear(c.Username)
		}()
	}
}
