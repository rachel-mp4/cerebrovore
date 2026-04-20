package handler

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/disintegration/imaging"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	_ "golang.org/x/image/webp"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/db"
	"github.com/rachel-mp4/cerebrovore/model"
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
	"github.com/rachel-mp4/lrcd"
	lrcpb "github.com/rachel-mp4/lrcproto/gen/go"
)

func (h *Handler) postThread(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	r.ParseMultipartForm(10 << 20)
	var thread types.Thread
	topic, ok := r.MultipartForm.Value["topic"]
	// len probably should always be non nil, but form is map to slice of string
	// so we check just in case maybe it's a nil slice for some strange reason,
	// if form is filled out correctly the first entry in slice is what we want
	if ok && len(topic) > 0 {
		maxlen := len("brevity is the soul of wit")
		mytopic := topic[0]
		if len(mytopic) > maxlen {
			mytopic = mytopic[:maxlen]
		}
		thread.Topic = &mytopic
	}
	thread.OP.Username = c.Username
	_, ok = r.MultipartForm.Value["anon"]
	thread.OP.Anon = ok
	nick, ok := r.MultipartForm.Value["nick"]
	if ok && len(nick) > 0 {
		thread.OP.Nick = &nick[0]
	}
	color, ok := r.MultipartForm.Value["color"]
	if ok && len(color) > 0 {
		c, err := utils.AToColor(color[0])
		if err == nil {
			thread.OP.Color = &c
		}
	}
	body, ok := r.MultipartForm.Value["body"]
	if ok && len(body) > 0 {
		b := body[0]
		thread.OP.TextContent = &types.TextContent{Body: b}
		thread.OP.Backlinks, _ = utils.ParseBodyForBacklinks(b)
	}
	img, _, err := r.FormFile("image")
	if err == nil {
		cid, err, code := saveFileToContentAddress(img)
		if err != nil {
			clog.Warn("image save: %s", err)
			http.Error(w, "some error apropos image", code)
			return
		}
		err = genThumbnail(cid)
		if err != nil {
			clog.Warn("thumbnail: %s", err)
			http.Error(w, "some error apropos image 222", code)
			return
		}

		thread.OP.ImageContent = &types.ImageContent{CID: cid}
		alt, ok := r.MultipartForm.Value["alt"]
		if ok && len(alt) > 0 {
			b := alt[0]
			thread.OP.ImageContent.Alt = &b
		}
	}

	tid := h.m.AddThread(thread.Topic)
	thread.ID = tid
	thread.OP.ID = tid
	thread.OP.ThreadID = tid
	ntid := utils.IDToA(tid)
	err = h.db.CreateThread(&thread, r.Context())
	if err != nil {
		clog.Warn("create thread: %s", err)
		err = h.m.DeleteThread(tid)
		if err != nil {
			clog.Fail("create thread rollback failed: %s", err)
			http.Error(w, "failed to create thread", http.StatusInternalServerError)
			return
		}
		http.Error(w, "failed to create thread", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/t/%s", ntid), http.StatusSeeOther)
	go h.m.NotifyNewThread(tid)
}

func mimeToExt(contentType string) (string, error) {
	switch contentType {
	case "image/jpeg":
		return ".jpeg", nil
	case "image/png":
		return ".png", nil
	case "image/gif":
		return ".gif", nil
	case "image/webp":
		return ".webp", nil
	default:
		return "", fmt.Errorf("disallowed contentType: %s", contentType)
	}
}

func (h *Handler) getBlob(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	cid := r.URL.Query().Get("cid")
	if len(cid) < 4 {
		http.Error(w, "cid too short", http.StatusBadRequest)
		return
	}
	if strings.ContainsAny(cid, "\\/") {
		http.Error(w, "not today, hacker!", http.StatusBadRequest)
		return
	}
	thumb := r.URL.Query().Get("thumb")
	ext := ""
	if thumb == "jpg" {
		ext = ".jpg"
	} else if thumb != "" {
		ext = ".png"
	}
	dir := filepath.Join("uploads", cid[:3], fmt.Sprintf("%s%s", cid[3:], ext))
	file, err := os.Open(dir)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer file.Close()
	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		clog.Warn("blob read: %s", err)
		http.Error(w, "can't read file mime", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(buf[:n])
	w.Header().Add("Content-Type", contentType)
	// TODO: be fully sure that this doesn't ruin things because we use query params
	// but it seems ok rn
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Write(buf[:n])
	file.WriteTo(w)
}

func (h *Handler) postBlob(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	r.ParseMultipartForm(10 << 20)
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "requires file", http.StatusBadRequest)
		return
	}
	cid, err, code := saveFileToContentAddress(file)
	if err != nil {
		clog.Warn("blob save: %s", err)
		http.Error(w, "encountered an error", code)
		return
	}
	err = genThumbnail(cid)
	if err != nil {
		clog.Warn("blob thumbnail: %s", err)
		http.Error(w, "encountered an error 2", http.StatusInternalServerError)
	}
	// i'm not really sure why there's a uuid, but i remember a few months ago
	// being certain it was necessary. not gonna think too hard about it haha
	type blobresp struct {
		CID  string `json:"cid"`
		UUID string `json:"uuid"`
	}
	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(blobresp{cid, r.FormValue("uuid")})
	if err != nil {
		clog.Warn("blob json encode: %s", err)
	}
}

// genThumbnail generates a thumnail for a given content id that
// is currently in our uploads folder. i save it as png, and since it should
// already have an extension in the cid, the final path for the thumbanil
// should look like "uploads/74d/670818d4421f572b6.jpeg.png". this way i can
// easily find the thumbnail when user requests it by just appending .png to
// the content id
// this DOES work for gifs, but i just ignore the gif thumbnails and only
// serve the full size gifs
func genThumbnail(cid string) error {
	if strings.HasSuffix(cid, ".gif") {
		return nil
	}
	dir := filepath.Join("uploads", cid[:3], cid[3:])
	file, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer file.Close()
	img, err := imaging.Decode(file, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}
	thumb := imaging.Fit(img, 192, 192, imaging.NearestNeighbor)
	thumbpath := filepath.Join("uploads", cid[:3], fmt.Sprintf("%s.png", cid[3:]))
	err = imaging.Save(thumb, thumbpath)
	return nil
}

func genPFPThumb(cid string) error {
	if strings.HasSuffix(cid, ".gif") {
		return nil
	}
	dir := filepath.Join("uploads", cid[:3], cid[3:])
	file, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer file.Close()
	img, err := imaging.Decode(file, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}
	thumb := imaging.Fit(img, 384, 384, imaging.Lanczos)
	thumbpath := filepath.Join("uploads", cid[:3], fmt.Sprintf("%s.jpg", cid[3:]))
	err = imaging.Save(thumb, thumbpath)
	return nil
}

func saveFileToContentAddress(file multipart.File) (cid string, err error, code int) {
	defer file.Close()
	// read first 512 byte into a buffer, so we can detect content type
	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		code = http.StatusInternalServerError
		return
	}
	contentType := http.DetectContentType(buf[:n])
	ext, err := mimeToExt(contentType)
	if err != nil {
		code = http.StatusBadRequest
		return
	}

	// want to hash file so we can store it as content address
	hasher := sha256.New()

	// remember to include the first up to 512 byte in hash
	_, err = hasher.Write(buf[:n])
	if err != nil {
		code = http.StatusInternalServerError
		return
	}
	_, err = io.Copy(hasher, file)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}

	hash := hasher.Sum(nil)
	// sha256 = 64 hex encode bytes
	hexhash := make([]byte, 64)
	hex.Encode(hexhash, hash)
	// only include first 20 hex encode bytes, since that's more than
	// enough and we have prettier url
	cid = fmt.Sprintf("%s%s", string(hexhash)[:20], ext)
	// nested directory structure for content id means that we don't have every
	// single file in one directory, maybe this is better or something
	dir := filepath.Join("uploads", cid[:3])
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}
	out, err := os.CreateTemp(dir, "upload-*")
	if err != nil {
		code = http.StatusInternalServerError
		return
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}
	_, err = io.Copy(out, file)
	if err != nil {
		code = http.StatusInternalServerError
		return
	}
	err = out.Close()
	if err != nil {
		code = http.StatusInternalServerError
		return
	}
	// full path is uploads > first 3 of hex encode hash > next 17 of hex encode hash .ext
	path := filepath.Join(dir, cid[3:])
	_, err = os.Stat(path)
	// err == nil => we already have this file
	if err == nil {
		os.Remove(out.Name())
		return
	} else if !os.IsNotExist(err) {
		// we want the is not exist error so we can make it exist
		code = http.StatusInternalServerError
		return
	}
	err = os.Rename(out.Name(), path)
	if err != nil {
		if os.IsExist(err) {
			// perhaps there's a race here, seems kinda unlikely
			os.Remove(out.Name())
			err = nil
			return
		}
		code = http.StatusInternalServerError
		return
	}
	os.Remove(out.Name())
	return
}

func (h *Handler) postPost(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	r.ParseMultipartForm(10 << 20)
	var post types.Post
	post.Username = c.Username
	_, ok := r.MultipartForm.Value["anon"]
	post.Anon = ok
	nick, ok := r.Form["nick"]
	if ok && len(nick) > 0 {
		post.Nick = &nick[0]
	}
	color, ok := r.MultipartForm.Value["color"]
	if ok && len(color) > 0 {
		c, err := utils.AToColor(color[0])
		if err == nil {
			post.Color = &c
		}
	}
	body, ok := r.MultipartForm.Value["body"]
	var extras []uint64
	if ok && len(body) > 0 {
		b := body[0]
		post.TextContent = &types.TextContent{Body: b}
		post.Backlinks, extras = utils.ParseBodyForBacklinks(b)
	}
	cid, ok := r.MultipartForm.Value["cid"]
	if ok && len(cid) > 0 {
		c := cid[0]
		alt, ok := r.MultipartForm.Value["alt"]
		if ok && len(alt) > 0 {
			post.ImageContent = &types.ImageContent{CID: c, Alt: &alt[0]}
		} else {
			post.ImageContent = &types.ImageContent{CID: c}
		}
	}

	ntid := r.PathValue("ntid")
	tid, err := utils.AToID(ntid)
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}
	post.ThreadID = tid
	id, ok := r.MultipartForm.Value["id"]
	if !ok || len(id) < 1 {
		clog.Warn("require post id")
		http.Error(w, "requires post id", http.StatusBadRequest)
		return
	}
	ii, err := utils.AToID(id[0])
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	post.ID = ii
	nonce, ok := r.MultipartForm.Value["nonce"]
	if !ok || len(nonce) < 1 {
		http.Error(w, "requires nonce", http.StatusBadRequest)
		return
	}
	mynonce, err := base64.StdEncoding.DecodeString(nonce[0])
	if err != nil {
		clog.Warn("bad nonce: %s", nonce[0])
		clog.Warn("%s", err)
		http.Error(w, "nonce encoded wrong", http.StatusBadRequest)
		return
	}

	truenonce := lrcd.GenerateNonce(ii, ntid, os.Getenv("LRCD_SECRET"))
	if !slices.Equal(mynonce, truenonce) {
		http.Error(w, "i think user tried to submit wrong post", http.StatusUnauthorized)
		return
	}

	rc, backlinks, err := h.db.CreatePost(&post, r.Context())
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "failed to create post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/t/%s", ntid), http.StatusSeeOther)
	go h.postPostPostFunFunc(c, &post, rc, backlinks, context.Background(), extras)
}

// postPostPostFunFunc is a func that runs post postpost, does assorted fun we want
// like informing lrc clients of parsed backlinks, and sending events to watchers
func (h *Handler) postPostPostFunFunc(c *Client, post *types.Post, replyCount int, backlinks []db.Backlink, ctx context.Context, extras []uint64) {
	if len(backlinks) != 0 {
		replies := make([]*lrcpb.Reply, 0, len(backlinks))
		for _, bl := range backlinks {
			if bl.From == post.ID {
				reply := lrcpb.Reply{
					Reply: &lrcpb.Reply_Attachreply{Attachreply: &lrcpb.AttachReply{From: &post.ID, To: bl.To}},
				}
				replies = append(replies, &reply)
			} else if bl.To == post.ID {
				reply := lrcpb.Reply{
					Reply: &lrcpb.Reply_Attachreply{Attachreply: &lrcpb.AttachReply{To: post.ID, From: &bl.From}},
				}
				replies = append(replies, &reply)
			} else {
				clog.Fail("bad backlink: from=%d to=%d", bl.From, bl.To)
			}
		}

		batch := lrcpb.Event_Replybatch{
			Replybatch: &lrcpb.ReplyBatch{
				Replies: replies,
			},
		}
		h.m.AddBacklinks(post.ThreadID, batch)
	}
	if post.Anon {
		h.m.NotifyReply(post.ThreadID, post.ID, nil, replyCount)
	} else {
		h.m.NotifyReply(post.ThreadID, post.ID, &c.Username, replyCount)
	}
	if replyCount < utils.BUMP_LIMIT {
		h.m.NotifyWatchers(post.ThreadID)
		changed, err := h.db.WatchThread(c.Username, post.ThreadID, ctx)
		if err != nil {
			clog.Warn("%s", err)
		}
		if changed {
			h.m.Watch(c.Username, post.ThreadID)
		}
	} else if replyCount == utils.BUMP_LIMIT {
		h.m.NotifyBumpLimit(post.ThreadID)
		err := h.db.RemoveWatchersFor(post.ThreadID, ctx)
		if err != nil {
			clog.Warn("%s", err)
		}
	} else if utils.MaxReplies(replyCount) {
		h.m.ReplyLimit(post.ThreadID)
	}
	type commands struct {
		play        bool
		skip        bool
		pause       bool
		unpause     bool
		molt        bool
		desh        bool
		deshell     bool
		debrainworm bool
	}
	cmd := commands{}
	for _, bl := range post.Backlinks {
		switch bl {
		case utils.PLAY_ID:
			cmd.play = true
		case utils.SKIP_ID:
			cmd.skip = true
		case utils.PAUSE_ID:
			cmd.pause = true
		case utils.DESH_ID:
			cmd.desh = true
		case utils.MOLT_ID:
			cmd.molt = true
		}
	}
	for _, ex := range extras {
		switch ex {
		case utils.UNPAUSE_EX:
			cmd.unpause = true
		case utils.DEBRAINWORM_EX:
			cmd.debrainworm = true
		case utils.DESHELL_EX:
			cmd.deshell = true
		}

	}
	if cmd.play {
		res, unpause := utils.ParseBodyForPlays(post.TextContent.Body)
		if len(res) != 0 {
			h.m.Queue(post.ThreadID, c.Username, res)
		}
		if unpause {
			h.m.Unpause(post.ThreadID, c.Username)
		}
	}
	if cmd.skip {
		h.m.Skip(post.ThreadID, c.Username)
	}
	if cmd.pause {
		h.m.Pause(post.ThreadID, c.Username)
	}
	if cmd.unpause {
		h.m.Unpause(post.ThreadID, c.Username)
	}
	if cmd.desh {
		h.selfban(c.Username, &post.ID, "*deshs u*", time.Now().Add(5*time.Minute), ctx)
	}
	if cmd.molt {
		h.selfban(c.Username, &post.ID, "*molts u*", time.Now().Add(5*time.Minute), ctx)
	}
	if cmd.deshell {
		h.selfban(c.Username, &post.ID, "*deshells u*", time.Now().Add(5*time.Minute), ctx)
	}
	if cmd.debrainworm {
		h.selfban(c.Username, &post.ID, "*debrainworms u*", time.Now().Add(5*time.Minute), ctx)
	}
}

func (h *Handler) getPost(c *Client, w http.ResponseWriter, r *http.Request) {
	npid := r.PathValue("npid")
	pid, err := utils.AToID(npid)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}
	tid, err := h.db.GetPostThreadID(pid, r.Context())
	if err != nil {
		http.Error(w, "post does not exist", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/t/%s#%s", utils.IDToA(tid), npid), http.StatusFound)
}

func (h *Handler) getTBumped(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		http.Error(w, "failed to get threads", http.StatusInternalServerError)
		return
	}
	bumpedT.exec(w, bumpedresp{tt})
}

func (h *Handler) getThread(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	ntid := r.PathValue("ntid")
	tid, err := utils.AToID(ntid)
	if err != nil {
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}
	t, err := h.db.GetThread(tid, r.Context())
	if err != nil {
		http.Error(w, "failed to get thread", http.StatusNotFound)
		return
	}
	title := ntid
	if t.Topic != nil {
		title = *t.Topic
	}
	br, err := h.makebase(title, c, r.Context())
	if err != nil {
		http.Error(w, "failed to get bumps", http.StatusInternalServerError)
		return
	}
	br.Accent = utils.ColorToAp(t.OP.Color)
	watched := h.db.IsWatched(c.Username, tid, r.Context())
	gtr := threadresp{
		baseresp: br,
		Thread:   t,
		Watched:  watched,
		Archived: utils.MaxReplies(t.ReplyCount),
	}
	threadT.exec(w, gtr)
}

func (h *Handler) getThreadWS(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "you seem like you're not logged in", http.StatusBadRequest)
		return
	}
	ntid := r.PathValue("ntid")
	tid, err := utils.AToID(ntid)
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}
	f, err := h.m.GetThreadWSHandler(uint32(tid), c.Username)
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "error getting ws handler", http.StatusInternalServerError)
		return
	}
	f(w, r)
}

func (h *Handler) getJSONWebsockets(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	opts := make([]model.Option, 0)
	ntid := r.URL.Query().Get("thread")
	if ntid != "" {
		tid, err := utils.AToID(ntid)
		if err == nil {
			opts = append(opts, model.WithThreadSocket(tid))
		}
	}
	watcher := r.URL.Query().Get("watcher")
	if watcher != "" {
		opts = append(opts, model.WithWatchedThreads(h.db.GetWatchedThreads, watcher == "and-new-threads"))
	}
	ntid = r.URL.Query().Get("wormwatch")
	if ntid != "" {
		tid, err := utils.AToID(ntid)
		if err == nil {
			opts = append(opts, model.WithWormwatch(tid))
		}
	}

	f, err := h.m.GetWebSockets(c.Username, opts...)
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	f(w, r)
}

func (h *Handler) watchThread(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	ntid := r.PathValue("ntid")
	tid, err := utils.AToID(ntid)
	if err != nil {
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}
	changed, err := h.db.WatchThread(c.Username, tid, r.Context())
	if err != nil {
		clog.Warn("%s", err)
		http.Error(w, "error watching thread", http.StatusInternalServerError)
		return
	}
	if changed {
		bl := h.m.Watch(c.Username, tid)
		if bl {
			changed, err := h.db.UnwatchThread(c.Username, tid, r.Context())
			if err != nil {
				clog.Warn("%s", err)
				http.Error(w, "error watching thread", http.StatusInternalServerError)
				return
			}
			if !changed {
				clog.Warn("change back fail?")
			}
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/t/%s", ntid), http.StatusSeeOther)
}

func (h *Handler) unwatchThread(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	ntid := r.PathValue("ntid")
	tid, err := utils.AToID(ntid)
	if err != nil {
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}
	changed, err := h.db.UnwatchThread(c.Username, tid, r.Context())
	if err != nil {
		http.Error(w, "error watching thread", http.StatusInternalServerError)
		return
	}
	if changed {
		h.m.Unwatch(c.Username, tid)
	}
	http.Redirect(w, r, fmt.Sprintf("/t/%s", ntid), http.StatusSeeOther)
}

func (h *Handler) catalog(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	base, err := h.makebase("catalog", c, r.Context())
	if err != nil {
		clog.Warn("bumps %s", err)
	}
	chrono := r.URL.Query().Get("chrono")
	isChrono := chrono != ""
	tr := catalogthreadsresp{
		baseresp: base,
		IsChrono: isChrono,
	}
	cursor := r.URL.Query().Get("cursor")
	if isChrono {
		var cc *uint32
		id, err := utils.AToID(cursor)
		if err == nil {
			cc = &id
		}
		fts, nc, err := h.db.GetRecentCatalog(cc, utils.THREADS_PER_CATALOG_PAGE, r.Context())
		if err != nil {
			clog.Warn("%s", err)
			http.Error(w, "error getting threads", http.StatusInternalServerError)
			return
		}
		tr.ChronoCursor = nc
		tr.ThreadThumbs = fts
	} else {
		var bc *time.Time
		t, err := utils.ParseTime(cursor)
		if err == nil {
			bc = &t
		}
		fts, nc, err := h.db.GetBumpedCatalog(bc, utils.THREADS_PER_CATALOG_PAGE, r.Context())
		if err != nil {
			clog.Warn("%s", err)
			http.Error(w, "error getting threads", http.StatusInternalServerError)
			return
		}
		tr.BumpCursor = nc
		tr.ThreadThumbs = fts
	}
	catalogT.exec(w, tr)
}

func (h *Handler) threads(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	base, err := h.makebase("threads", c, r.Context())
	if err != nil {
		clog.Warn("bumps %s", err)
	}
	chrono := r.URL.Query().Get("chrono")
	isChrono := chrono != ""
	tr := catalogthreadsresp{
		baseresp: base,
		IsChrono: isChrono,
	}
	cursor := r.URL.Query().Get("cursor")
	if isChrono {
		var cc *uint32
		id, err := utils.AToID(cursor)
		if err == nil {
			cc = &id
		}
		fts, nc, err := h.db.GetRecentThreads(cc, utils.THREADS_PER_INDEX_PAGE, r.Context())
		if err != nil {
			clog.Warn("%s", err)
			http.Error(w, "error getting threads", http.StatusInternalServerError)
			return
		}
		tr.ChronoCursor = nc
		tr.ThreadThumbs = fts
	} else {
		var bc *time.Time
		t, err := utils.ParseTime(cursor)
		if err == nil {
			bc = &t
		}
		fts, nc, err := h.db.GetBumpedThreads(bc, utils.THREADS_PER_INDEX_PAGE, r.Context())
		if err != nil {
			clog.Warn("%s", err)
			http.Error(w, "error getting threads", http.StatusInternalServerError)
			return
		}
		tr.BumpCursor = nc
		tr.ThreadThumbs = fts
	}
	threadsT.exec(w, tr)
	if err != nil {
		clog.Warn("%s", err)
	}
}
