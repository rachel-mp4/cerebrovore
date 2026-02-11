package handler

import (
	"html/template"
)

var homeT = template.Must(template.ParseFiles("./tmpl/base.html", "./tmpl/home.html", "./tmpl/threads.html", "./tmpl/empty.html"))
var beepT = template.Must(template.ParseFiles("./tmpl/base.html", "./tmpl/beep.html", "./tmpl/threads.html", "./tmpl/empty.html"))
