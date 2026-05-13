package handler

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
)

func (h *Handler) profile(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		return
	}
	username := r.PathValue("username")
	// just a lil mapping to make it so users with legacy usernames, their old posts still do an OK job of linking to their profile
	username = strings.ToLower(username)
	profile, err := h.db.GetFullProfile(username, r.Context())
	if err != nil {
		clog.Info("%s", err)
		http.Error(w, "failed to get profile", http.StatusNotFound)
		return
	}
	if profile.Banned {
		http.Error(w, "user is banned", http.StatusForbidden)
		return
	}
	var title = profile.Username
	if profile.DisplayName != nil {
		title = *profile.DisplayName
	}
	br, err := h.makebase(title, c, r.Context())
	if err != nil {
		clog.Info("%s", err.Error())
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	br.Accent = utils.ColorToAp(profile.Color)
	profileT.exec(w, profileresp{br, profile})
}

func (h *Handler) postProfile(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		return
	}
	err := r.ParseForm()
	if err != nil {
		clog.Info("%s", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var p types.ProfileHead
	p.Username = c.Username
	name, ok := r.Form["display-name"]
	if ok && len(name) > 0 {
		p.DisplayName = &name[0]
	}
	color, ok := r.Form["color"]
	if ok && len(color) > 0 {
		c, err := utils.AToColor(color[0])
		if err == nil {
			p.Color = &c
		}
	}
	status, ok := r.Form["status"]
	if ok && len(status) > 0 {
		trimmedStatus := status[0]
		if len(trimmedStatus) > 144 {
			trimmedStatus = trimmedStatus[:144]
		}
		p.Status = &trimmedStatus
	}
	bio, ok := r.Form["bio"]
	if ok && len(bio) > 0 {
		trimmedBio := bio[0]
		if len(trimmedBio) > 1000 {
			trimmedBio = trimmedBio[:1000]
		}
		p.Bio = &bio[0]
		mono := r.Form.Get("mono")
		bim := mono != ""
		if bim {
			p.BioIsMono = &bim
		}
	}

	atid := r.Form.Get("at-identifier")

	if atid != "" {
		if strings.HasPrefix(atid, "did:") {
			// good enough, it's a did, we don't care
			p.AtIdentifier = &atid
		} else {
			if !strings.Contains(atid, ".") || strings.ContainsAny(atid, "/:@") || net.ParseIP(atid) != nil {
				clog.Warn("@%s is being a little too clever - SSRF attempt (atproto ID)", c.Username)
				http.Error(w, "clever, huh?", http.StatusBadRequest)
				return
			}
			records, err := net.DefaultResolver.LookupTXT(r.Context(), fmt.Sprintf("_atproto.%s", atid))
			if err != nil {
				records = nil
				req, err2 := http.NewRequestWithContext(r.Context(), "GET", fmt.Sprintf("https://%s/.well-known/atproto-did", atid), nil)
				if err2 != nil {
					clog.Warn("err1: %s err2: %s", err, err2)
					http.Error(w, "errors with atproto id", http.StatusBadRequest)
					return
				}
				resp, err2 := http.DefaultClient.Do(req)
				if err2 != nil {
					clog.Warn("err1: %s err2: %s", err, err2)
					http.Error(w, "errors with atproto id", http.StatusBadRequest)
					return
				}
				defer resp.Body.Close()
				data, err2 := io.ReadAll(resp.Body)
				if err2 != nil {
					clog.Warn("err1: %s err2: %s", err, err2)
					http.Error(w, "errors with atproto id", http.StatusBadRequest)
					return
				}
				if strings.HasPrefix(string(data), "did:") {
					atid = strings.TrimSpace(string(data))
					p.AtIdentifier = &atid
				}
			}
			for _, record := range records {
				if strings.HasPrefix(record, "did=did:") {
					atid = strings.TrimPrefix(record, "did=")
					p.AtIdentifier = &atid
				}
			}
		}
	}

	err = h.db.UpdateProfile(&p, r.Context())
	if err != nil {
		clog.Info("%s", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/profile/%s", c.Username), http.StatusFound)
}

func (h *Handler) postAvatar(c *Client, w http.ResponseWriter, r *http.Request) {
	var p types.ProfileHead
	p.Username = c.Username
	avatar, _, err := r.FormFile("avatar")
	if err == nil {
		cid, err, code := saveFileToContentAddress(avatar)
		if err != nil {
			clog.Warn("image save: %s", err)
			http.Error(w, "some error apropos image", code)
			return
		}
		var isPixel bool
		pixel, ok := r.MultipartForm.Value["pixel"]
		if ok && len(pixel) != 0 {
			isPixel = pixel[0] != ""
		}
		p.IsPixelArt = &isPixel
		if isPixel {
			err = genThumbnail(cid)
		} else {
			err = genPFPThumb(cid)
		}
		if err != nil {
			clog.Warn("thumbnail: %s", err)
			http.Error(w, "some error apropos image 222", code)
			return
		}
		p.Avatar = &cid
	}
	err = h.db.UpdateProfilePicture(&p, r.Context())
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "failed to update", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/profile/%s", c.Username), http.StatusFound)
}

func (h *Handler) postContents(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	err := r.ParseForm()
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var pe types.ProfileExtras
	fheader := r.Form.Get("f-header")
	if fheader != "" {
		pe.FriendsHeader = &fheader
	}
	pheader := r.Form.Get("p-header")
	if pheader != "" {
		pe.PostsHeader = &pheader
	}
	lheader := r.Form.Get("l-header")
	if lheader != "" {
		pe.LinksHeader = &lheader
	}

	friends, ok := r.Form["username"]
	fcomments, bok := r.Form["f-comment"]
	if !ok || !bok || len(friends) != len(fcomments) || len(friends) != 12 {
		clog.Info("not ok!")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	for i := range 12 {
		if friends[i] == "" {
			continue
		}
		pe.FriendInsertUsername = append(pe.FriendInsertUsername, friends[i])
		if fcomments[i] != "" {
			pe.FriendInsertComments = append(pe.FriendInsertComments, &fcomments[i])
		} else {
			pe.FriendInsertComments = append(pe.FriendInsertComments, nil)
		}
	}

	posts, ok := r.Form["post"]
	pcomments, bok := r.Form["p-comment"]
	bodys, cok := r.Form["body"]
	if !ok || !bok || !cok || len(pcomments) != 12 || len(bodys) < 12 || len(posts) != 12 {
		clog.Info("not ok!")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// stupid hack because unchecked boxes don't submit
	bi := 0
	for i := range 12 {
		if posts[i] == "" {
			continue
		}
		pe.PostInsertIds = append(pe.PostInsertIds, utils.AToIDf(posts[i]))
		if pcomments[i] != "" {
			pe.PostInsertComments = append(pe.PostInsertComments, &pcomments[i])
		} else {
			pe.PostInsertComments = append(pe.PostInsertComments, nil)
		}
		bi += 1
		if bi < len(bodys) && bodys[bi] != "skip" {
			bi += 1
			pe.PostInsertBools = append(pe.PostInsertBools, true)
		} else {
			pe.PostInsertBools = append(pe.PostInsertBools, false)
		}
	}

	links, ok := r.Form["link"]
	lcomments, bok := r.Form["l-comment"]
	if !ok || !bok || len(links) != len(lcomments) || len(links) != 12 {
		clog.Info("not ok!")
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	for i := range 12 {
		if links[i] == "" {
			continue
		}
		trylink := links[i]
		u, err := url.Parse(trylink)
		if err != nil {
			continue
		}
		if !(u.Scheme == "http" || u.Scheme == "https") {
			if u.Scheme != "" {
				continue
			}
			u.Scheme = "https"
		}
		pe.LinkInsertLinks = append(pe.LinkInsertLinks, u.String())
		if lcomments[i] != "" {
			pe.LinkInsertComments = append(pe.LinkInsertComments, &lcomments[i])
		} else {
			pe.LinkInsertComments = append(pe.LinkInsertComments, nil)
		}
	}

	err = h.db.UpdateProfileContents(c.Username, &pe, r.Context())
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/profile/%s", c.Username), http.StatusFound)
}

func (h *Handler) editProfile(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		return
	}
	p, err := h.db.GetFullProfile(c.Username, r.Context())
	if err != nil {
		err2 := h.db.InitializeProfile(c.Username, r.Context())
		if err2 != nil {
			clog.Warn("err: %s, err2: %s", err, err2)
			http.Error(w, "error initializing profile", http.StatusInternalServerError)
			return
		}
		p, err2 = h.db.GetFullProfile(c.Username, r.Context())
		if err2 != nil {
			clog.Warn("err: %s, err2: %s", err, err2)
			http.Error(w, "error initializing profile", http.StatusInternalServerError)
			return
		}
	}
	base, err := h.makebase("edit profile", c, r.Context())
	if err != nil {
		clog.Info("%s", err)
		http.Error(w, "error rendering", http.StatusInternalServerError)
		return
	}
	editprofileT.exec(w, editprofileresp{
		base,
		p,
	})
}

func (h *Handler) poke(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		return
	}
	pokee := r.PathValue("username")
	if pokee == c.Username {
		profileT.error(w, "no poking yourself")
		return
	}
	_, err := h.db.GetProfile(pokee, r.Context())
	if err != nil {
		profileT.error(w, "user does not exist. are you hacking?")
		return
	}
	var msg *string
	m := r.FormValue("message")
	if m != "" {
		msg = &m
	}
	err = h.db.CreatePokeNotification(pokee, c.Username, msg, r.Context())
	if err != nil {
		profileT.error(w, "create poke notification error aaaaa")
		return
	}
	profileT.poked(w)
	go h.m.DispatchNotification(pokee, nil)
}
