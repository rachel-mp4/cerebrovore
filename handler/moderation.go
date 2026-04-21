package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
)

func (h *Handler) inspectPost(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || !c.IsMod {
		http.Error(w, "not authorized to moderate", http.StatusUnauthorized)
		return
	}
	nid := r.FormValue("postid")
	id, err := utils.AToID(nid)
	if err != nil {
		moderateT.error(w, err.Error())
		return
	}
	post, err := h.db.GetPost(id, r.Context())
	if err != nil {
		moderateT.error(w, err.Error())
		return
	}
	moderateT.inspect(w, post)
}

func (h *Handler) moderate(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || !c.IsMod {
		http.Error(w, "not authorized to moderate", http.StatusUnauthorized)
		return
	}
	id := r.URL.Query().Get("id")
	var autofillid *string
	if id != "" {
		autofillid = &id
	}
	base, _ := h.makebase("moderation", c, r.Context())
	appeals, _, err := h.db.GetAppeals(10, nil, r.Context())
	if err != nil {
		clog.Dbug("%s", err)
		return
	}
	reports, cursor, err := h.db.GetReports(10, nil, r.Context())
	if err != nil {
		clog.Dbug("%s", err)
		return
	}
	moderateT.exec(w, moderateresp{
		base,
		appeals,
		reports,
		cursor,
		autofillid,
	})
}

func (h *Handler) administrate(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || c.Username != os.Getenv("ADMIN_USERNAME") {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	mods, err := h.db.GetModerators(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	base, err := h.makebase("administrate", c, r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	adminT.exec(w, adminresp{
		base,
		mods,
	})
}

func (h *Handler) addModerator(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || c.Username != os.Getenv("ADMIN_USERNAME") {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	username := r.FormValue("username")
	err := h.db.MakeModerator(username, r.Context())
	if err != nil {
		moderateT.error(w, err.Error())
		return
	}
	adminT.plusmodsuccess(w, username)
}

func (h *Handler) postAppealVerdict(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || !c.IsMod {
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
		moderateT.error(w, err.Error())
		return
	}
	if reject != "" {
		err := h.db.Reject(banid, r.Context())
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		moderateT.reject(w)
		return
	}
	err = h.db.Unban(banid, r.Context())
	if err != nil {
		moderateT.error(w, err.Error())
		return
	}
	moderateT.accept(w)
}

func (h *Handler) cancelAction(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || !c.IsMod {
		http.Error(w, "not authorized to not moderate wait-", http.StatusUnauthorized)
		return
	}
	moderateT.canceled(w)
}

func (h *Handler) takeAction(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || !c.IsMod {
		http.Error(w, "not authorized to moderate", http.StatusUnauthorized)
		return
	}
	err := r.ParseForm()
	if err != nil {
		moderateT.error(w, err.Error())
		return
	}
	nid := r.FormValue("postid")
	username := r.FormValue("username")
	reason := r.FormValue("reason")
	comment := r.FormValue("comment")
	until := r.FormValue("until")
	confirm := r.FormValue("confirm")

	var action types.Action
	var actioning string
	var actioned string
	if nid == "" && username == "" {
		// no id and no username = no info to work with
		moderateT.error(w, "must fill one field")
		return
	} else if nid == "" && username != "" {
		// no id and username = just ban them
		actioning = "just performing a ban!"
		actioned = "just performed a ban!"

		ban := composeBan(username, reason, comment, until)
		action.Ban = &ban
	} else if nid != "" && username != "" {
		// id and username = delete and ban them if the post with that id is really that user's
		actioning = "deleting a %s AND banning"
		actioned = "deleted a %s AND banned"

		id, err := utils.AToID(nid)
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		post, err := h.db.GetPost(id, r.Context())
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		action.Post = post
		if username != post.Username {
			moderateT.error(w, fmt.Sprintf("username: %s does not match post username: %s", username, post.Username))
			return
		}
		ban := composeBan(username, reason, comment, until)
		ban.PostId = &id
		action.Ban = &ban
		if post.ID == post.ThreadID {
			actioning = fmt.Sprintf(actioning, "THREAD")
			actioned = fmt.Sprintf(actioned, "THREAD")
		} else {
			actioning = fmt.Sprintf(actioning, "post")
			actioned = fmt.Sprintf(actioned, "post")
		}
	} else if reason != "" || comment != "" || until != "" {
		// id and no username but ban info = delete and ban because the mod is probably lazy, however this is a bit worrisome so it should be loud
		actioning = "deleting a %s AND banning WITHOUT CONFIRMING username"
		actioned = "deleted a %s AND banned WITHOUT CONFIRMING username"

		id, err := utils.AToID(nid)
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		post, err := h.db.GetPost(id, r.Context())
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		action.Post = post
		ban := composeBan(post.Username, reason, comment, until)
		ban.PostId = &id
		action.Ban = &ban
		if post.ID == post.ThreadID {
			actioning = fmt.Sprintf(actioning, "THREAD")
			actioned = fmt.Sprintf(actioned, "THREAD")
		} else {
			actioning = fmt.Sprintf(actioning, "post")
			actioned = fmt.Sprintf(actioned, "post")
		}
	} else if nid != "" && username == "" && reason == "" && comment == "" && until == "" {
		// this should be the last case, in which only the id was filled in
		actioning = "just deleting a %s"
		actioned = "just deleted a %s"

		id, err := utils.AToID(nid)
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		post, err := h.db.GetPost(id, r.Context())
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		action.Post = post
		if post.ID == post.ThreadID {
			actioning = fmt.Sprintf(actioning, "THREAD")
			actioned = fmt.Sprintf(actioned, "THREAD")
		} else {
			actioning = fmt.Sprintf(actioning, "post")
			actioned = fmt.Sprintf(actioned, "post")
		}
	} else {
		// i think i exhausted all cases but i am stupid maybe
		clog.Fail("YOU ARE STUPID")
		panic("YOU ARE STUPID")
	}

	if confirm != "" {
		if action.Post != nil {
			if action.Post.ThreadID == action.Post.ID {
				err := h.db.DeleteThread(action.Post.ThreadID, r.Context())
				if err != nil {
					moderateT.error(w, err.Error())
					return
				}
				err = h.m.DeleteThread(action.Post.ThreadID)
				if err != nil {
					moderateT.error(w, err.Error())
					return
				}
			}
			err = h.db.DeletePost(action.Post.ID, r.Context())
			if err != nil {
				moderateT.error(w, err.Error())
				return
			}
		}
		if action.Ban != nil {
			err = h.db.Ban(action.Ban, r.Context())
			if err != nil {
				moderateT.error(w, err.Error())
				return
			}
			go h.postBanCleanup(action.Ban.Username)
		}
		moderateT.confirmed(w, &action, actioned)
		return
	}
	moderateT.confirm(w, &action, actioning, nid, username, reason, comment, until)
}

func composeBan(username, reason, comment, until string) types.Ban {
	b := types.Ban{
		Username: username,
	}
	if reason != "" {
		b.Reason = &reason
	}
	if comment != "" {
		b.Comment = &comment
	}
	if until == "" {
		b.Until = time.Now().Add(1 * time.Hour)
	} else {
		untild, err := time.ParseDuration(until)
		if err != nil {
			b.Until = time.Now().Add(1 * time.Hour)
		} else {
			b.Until = time.Now().Add(untild)
		}
	}
	return b
}

func (h *Handler) selfban(username string, postid *uint32, reason string, til time.Time, ctx context.Context) error {
	err := h.db.SelfBan(username, postid, reason, til, ctx)
	if err == nil {
		go h.postBanCleanup(username)
	}
	return err
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
	banT.appeal(w)
}

func (h *Handler) getReports(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil || !c.IsMod {
		http.Error(w, "not authorized to view reports", http.StatusUnauthorized)
		return
	}
	l := r.FormValue("limit")
	limit, err := strconv.Atoi(l)
	if err != nil {
		moderateT.error(w, err.Error())
		return
	}
	a := r.FormValue("after")
	var after *int
	if a != "" {
		af, err := strconv.Atoi(a)
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		after = &af
	}

	reports, cursor, err := h.db.GetReports(limit, after, r.Context())
	if err != nil {
		moderateT.error(w, err.Error())
		return
	}
	moderateT.reports(w, reports, cursor)
}

func (h *Handler) postReport(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized to report", http.StatusUnauthorized)
		return
	}
	if c.IsMod {
		aid := r.FormValue("id")
		iid, err := strconv.Atoi(aid)
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		err = h.db.ReviewReport(iid, c.Username, r.Context())
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		moderateT.review(w)
		return
	}
	report := types.Report{Reporter: c.Username}
	postid := r.FormValue("postid")
	var post *types.Post
	if postid != "" {
		id, err := utils.AToID(postid)
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		report.PostId = &id
		post, err = h.db.GetPost(id, r.Context())
		if err != nil {
			moderateT.error(w, err.Error())
			return
		}
		if post != nil {
			report.Reported = post.Username
		}
	}
	profile := r.FormValue("profile")
	report.ForProfile = profile != ""
	username := r.FormValue("username")
	if username != "" {
		if report.Reported == "" {
			report.Reported = username
		} else if report.Reported != username {
			clog.Info("provided username %s does not match the one we have on file %s, ignoring provided username", username, report.Reported)
		}
	}
	if report.Reported == "" {
		moderateT.error(w, "post does not exist yet, yell at devs to improve lrcd (or add this yourself!)")
		return
	}
	reason := r.FormValue("reason")
	if reason != "" {
		report.Reason = &reason
	}
	err := h.db.Report(&report, r.Context())
	if err != nil {
		clog.Warn("%s", err)
		moderateT.error(w, "error saving report")
		return
	}
	moderateT.report(w)
}
