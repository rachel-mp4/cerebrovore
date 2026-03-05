package handler

import (
	"github.com/rachel-mp4/cerebrovore/utils"
	"html/template"
)

var homeT *template.Template
var mockloginT *template.Template
var beepT *template.Template
var threadT *template.Template
var newthreadT *template.Template

func init() {
	homeT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/empty.html",
		"./tmpl/home.html",
	)
	mockloginT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/empty.html",
		"./tmpl/mocklogin.html",
	)
	beepT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/empty.html",
		"./tmpl/beep.html",
	)
	threadT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/empty.html",
		"./tmpl/thread.html",
		"./tmpl/partial/post.html",
	)
	newthreadT = newTemplate(
		"./tmpl/base.html",
		"./tmpl/threads.html",
		"./tmpl/partial/threadlink.html",
		"./tmpl/empty.html",
		"./tmpl/newthread.html",
	)
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
			}).ParseFiles(files...),
	)
}
