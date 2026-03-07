package handler

import (
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
	"html/template"
)

var homeT *template.Template
var mockloginT *template.Template
var beepT *template.Template
var threadT *template.Template
var newthreadT *template.Template
var bumpedT *template.Template
var threadsT *template.Template

func init() {
	homeT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/home.html",
	)
	mockloginT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/bumped-threads.html",
		"./tmpl/empty.html",
		"./tmpl/mocklogin.html",
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
		"./tmpl/empty.html",
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
				"renderImageBody": utils.RenderImageBody,
				"renderTextBody":  utils.RenderTextBody,
				"colorIsDark":     utils.ColorIsDark,
				"colorToA":        utils.ColorToAp,
				"maxReplies":      utils.MaxReplies,
				"formatTime":      utils.FormatTime,
			}).ParseFiles(files...),
	)
}
