package handler

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"

	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/rachel-mp4/cerebrovore/db"
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
		if len(topic) > maxlen {
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
		thread.OP.Backlinks = utils.ParseBodyForBacklinks(b)
	}
	img, _, err := r.FormFile("image")
	if err == nil {
		cid, err, code := saveFileToContentAddress(img)
		if err != nil {
			log.Println(err)
			http.Error(w, "some error apropos image", code)
			return
		}
		err = genThumbnail(cid)
		if err != nil {
			log.Println(err)
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
		log.Println(err.Error())
		err = h.m.DeleteThread(tid)
		if err != nil {
			log.Println("real bad case!!!!!")
			http.Error(w, "failed to create thread", http.StatusInternalServerError)
			return
		}
		http.Error(w, "failed to create thread", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/t/%s", ntid), http.StatusSeeOther)
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
		return "", errors.New(fmt.Sprintf("disallowed contentType: %s", contentType))
	}
}

func (h *Handler) getBlob(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		log.Println("hi")
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	cid := r.URL.Query().Get("cid")
	if len(cid) < 4 {
		http.Error(w, "cid too short", http.StatusBadRequest)
	}
	thumb := r.URL.Query().Get("thumb")
	ext := ""
	if thumb != "" {
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
		log.Println(err)
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
		log.Println(err)
		http.Error(w, "encountered an error", code)
		return
	}
	err = genThumbnail(cid)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}
}

// genThumbnail generates a thumnail for a given content id that
// is currently in our uploads folder. i save it as png, and since it should
// already have an extension in the cid, the final path for the thumbanil
// should look like "uploads/74d/670818d4421f572b6.jpeg.png". this way i can
// easily find the thumbnail when user requests it by just appending .png to
// the content id
func genThumbnail(cid string) error {
	dir := filepath.Join("uploads", cid[:3], cid[3:])
	file, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	thumb := imaging.Fit(img, 200, 200, imaging.NearestNeighbor)
	thumbpath := filepath.Join("uploads", cid[:3], fmt.Sprintf("%s.png", cid[3:]))
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
	out, err := os.CreateTemp("", "upload-*")
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
	err = os.Link(out.Name(), path)
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
	if ok && len(body) > 0 {
		b := body[0]
		post.TextContent = &types.TextContent{Body: b}
		post.Backlinks = utils.ParseBodyForBacklinks(b)
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
		log.Println(err.Error())
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}
	post.ThreadID = tid
	id, ok := r.MultipartForm.Value["id"]
	if !ok || len(id) < 1 {
		log.Println("require post id")
		http.Error(w, "requires post id", http.StatusBadRequest)
		return
	}
	ii, err := utils.AToID(id[0])
	if err != nil {
		log.Println(err.Error())
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
		log.Println(nonce[0])
		log.Println(err.Error())
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
		log.Println(err.Error())
		http.Error(w, "failed to create post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/t/%s", ntid), http.StatusSeeOther)
	go h.postPostPostFunFunc(c, &post, rc, backlinks, context.Background())
}

// postPostPostFunFunc is a func that runs post postpost, does assorted fun we want
// like informing lrc clients of parsed backlinks, and sending events to watchers
func (h *Handler) postPostPostFunFunc(c *Client, post *types.Post, replyCount int, backlinks []db.Backlink, ctx context.Context) {
	if len(backlinks) != 0 {
		log.Println("sending backlinks!")
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
				log.Println("BAD BACKLINK!")
			}
		}

		batch := lrcpb.Event_Replybatch{
			Replybatch: &lrcpb.ReplyBatch{
				Replies: replies,
			},
		}
		h.m.AddBacklinks(post.ThreadID, batch)
	}
	if replyCount < utils.BUMP_LIMIT {
		h.m.NotifyWatchers(post.ThreadID)
		err := h.db.WatchThread(c.Username, post.ThreadID, ctx)
		if err != nil {
			log.Println(err)
		}
	} else if replyCount == utils.BUMP_LIMIT {
		h.m.NotifyBumpLimit(post.ThreadID)
		err := h.db.RemoveWatchersFor(post.ThreadID, ctx)
		if err != nil {
			log.Println(err.Error())
		}
	} else if utils.MaxReplies(replyCount) {
		h.m.ReplyLimit(post.ThreadID)
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
	type tbumpedresp struct {
		Threads []types.Thread
	}
	err = bumpedT.ExecuteTemplate(w, "bumped-threads", tbumpedresp{tt})
	if err != nil {
		log.Println(err.Error())
	}
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
	t, tcrsr, err := h.db.GetThread(tid, nil, utils.REPLY_LIMIT, r.Context())
	if err != nil {
		http.Error(w, "failed to get thread", http.StatusNotFound)
		return
	}
	if tcrsr != nil {
		log.Println(fmt.Sprintf("thread %d (#%s) has over REPLY_LIMIT replies (%d)", tid, ntid, t.ReplyCount))
	}
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		http.Error(w, "failed to get threads", http.StatusInternalServerError)
		return
	}
	title := ntid
	watched := h.db.IsWatched(c.Username, tid, r.Context())

	type getthreadresp struct {
		baseresp
		Thread   *types.Thread
		Archived bool
		Watched  bool
	}
	gtr := getthreadresp{
		baseresp: baseresp{
			h.ca,
			title,
			tt,
		},
		Thread:   t,
		Watched:  watched,
		Archived: utils.MaxReplies(t.ReplyCount),
	}
	err = threadT.ExecuteTemplate(w, "base", gtr)
	if err != nil {
		log.Println(err.Error())
	}
}

func (h *Handler) getThreadWS(c *Client, w http.ResponseWriter, r *http.Request) {
	ntid := r.PathValue("ntid")
	tid, err := utils.AToID(ntid)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}
	f, err := h.m.GetThreadWSHandler(uint32(tid))
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "error getting ws handler", http.StatusInternalServerError)
		return
	}
	f(w, r)
}

func (h *Handler) getThreadSocket(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	ids, err := h.db.GetWatchedThreads(c.Username, r.Context())
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "error finding watched threads", http.StatusInternalServerError)
		return
	}

	f := h.m.GetThreadSocketHandler(ids)
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
	err = h.db.WatchThread(c.Username, tid, r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, "error watching thread", http.StatusInternalServerError)
		return
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
	err = h.db.UnwatchThread(c.Username, tid, r.Context())
	if err != nil {
		http.Error(w, "error watching thread", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/t/%s", ntid), http.StatusSeeOther)
}

func (h *Handler) threads(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}
	type threadsresp struct {
		baseresp
		IsChrono     bool
		ChronoCursor *uint32
		BumpCursor   *time.Time
		ThreadThumbs []types.Thread
	}
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "error getting bumps", http.StatusInternalServerError)
		return
	}
	chrono := r.URL.Query().Get("chrono")
	isChrono := chrono != ""
	tr := threadsresp{
		baseresp{
			h.ca,
			"threads",
			tt,
		},
		isChrono,
		nil,
		nil,
		nil,
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
			log.Println(err.Error())
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
			log.Println(err.Error())
			http.Error(w, "error getting threads", http.StatusInternalServerError)
			return
		}
		tr.BumpCursor = nc
		tr.ThreadThumbs = fts
	}
	err = threadsT.ExecuteTemplate(w, "base", tr)
	if err != nil {
		log.Println(err.Error())
	}
}
