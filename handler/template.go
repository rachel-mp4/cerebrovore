package handler

import (
	"context"
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
	"html/template"
)

var homeT *template.Template
var moderateT *template.Template
var meT *template.Template
var loginT *template.Template
var accountT *template.Template
var beepT *template.Template
var threadT *template.Template
var newthreadT *template.Template
var bumpedT *template.Template
var threadsT *template.Template
var catalogT *template.Template
var patchT *template.Template
var codeT *template.Template
var codeerrT *template.Template
var banT *template.Template
var profileT *template.Template
var editprofileT *template.Template

func init() {
	homeT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/home.html",
	)
	patchT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/patch-notes.html",
		"./tmpl/partial/patch.html",
		"./tmpl/partial/note.html",
	)
	moderateT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/moderate.html",
		"./tmpl/partial/post.html",
	)
	meT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/me.html",
	)
	banT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/emptyl.html",
		"./tmpl/empty.html",
		"./tmpl/ban.html",
		"./tmpl/partial/post.html",
	)
	loginT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/emptyl.html",
		"./tmpl/empty.html",
		"./tmpl/login.html",
	)
	accountT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/emptyl.html",
		"./tmpl/empty.html",
		"./tmpl/account.html",
	)
	beepT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/beep.html",
	)
	threadT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/wormwatch.html",
		"./tmpl/thread.html",
		"./tmpl/partial/post.html",
	)
	profileT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/profile.html",
		"./tmpl/partial/post.html",
	)
	editprofileT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/editprofile.html",
	)
	newthreadT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/newthread.html",
	)
	threadsT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/index.html",
		"./tmpl/partial/post.html",
	)
	catalogT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/catalog.html",
		"./tmpl/partial/post.html",
	)
	bumpedT = newTemplate(
		"./tmpl/bumped-threads.html",
		"./tmpl/partial/threadlink.html",
	)
	codeT = newTemplate(
		"./tmpl/code.html",
	)
	codeerrT = newTemplate(
		"./tmpl/codeerr.html",
	)
}

type baseresp struct {
	CompiledAssets *CompiledAssets
	Title          string
	Threads        []types.Thread
	Crack          string
	ReplyCount     *int
	Accent         string
	Websockets     bool
	Username       string
}

func (h *Handler) makebase(title string, username string, ctx context.Context) (*baseresp, error) {
	tt, err := h.db.GetBumps(ctx)
	return &baseresp{
		h.ca,
		title,
		tt,
		h.crack,
		nil,
		"var(--primary)",
		true,
		username,
	}, err
}

func newTemplate(files ...string) *template.Template {
	return template.Must(
		template.New("").Funcs(
			template.FuncMap{
				"idtoa":            utils.IDToA,
				"intto36a":         utils.IntTo36A,
				"renderImageBody":  utils.RenderImageBody,
				"renderAvatarPFP":  utils.RenderAvatarPFP,
				"renderTextBody":   utils.RenderTextBody,
				"colorIsDark":      utils.ColorIsDark,
				"colorToA":         utils.ColorToAp,
				"maxReplies":       utils.MaxReplies,
				"maxBumps":         utils.MaxBumps,
				"formatTime":       utils.FormatTime,
				"remainingTime":    utils.RemainingTime,
				"timeSince":        utils.TimeSince,
				"ftime":            utils.FTime,
				"topicOrIdtoa":     types.TopicOrIdtoa,
				"percentRemaining": utils.PercentRemaining,
				"boolPtrIsTrue":    boolPtrIsTrue,
			}).ParseFiles(files...),
	)
}

func boolPtrIsTrue(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}
