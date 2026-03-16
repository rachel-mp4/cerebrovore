package handler

import (
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
var patchT *template.Template

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
	)
	meT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/me.html",
	)
	loginT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/login.html",
	)
	accountT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
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
	)
	bumpedT = newTemplate(
		"./tmpl/bumped-threads.html",
		"./tmpl/partial/threadlink.html",
	)
}

type baseresp struct {
	CompiledAssets *CompiledAssets
	Title          string
	Threads        []types.Thread
}

func newTemplate(files ...string) *template.Template {
	return template.Must(
		template.New("").Funcs(
			template.FuncMap{
				"idtoa":           utils.IDToA,
				"intto36a":        utils.IntTo36A,
				"renderImageBody": utils.RenderImageBody,
				"renderTextBody":  utils.RenderTextBody,
				"colorIsDark":     utils.ColorIsDark,
				"colorToA":        utils.ColorToAp,
				"maxReplies":      utils.MaxReplies,
				"formatTime":      utils.FormatTime,
				"ftime":           utils.FTime,
			}).ParseFiles(files...),
	)
}
