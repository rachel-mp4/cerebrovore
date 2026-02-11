package handler

import (
	"net/http"
)

type Handler struct {
	mux  *http.ServeMux
	prod *Prod
}

type Prod struct {
	ChatPath string
	BeepPath string
}

func NewHandler(prod *Prod) Handler {
	h := Handler{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.home)
	mux.HandleFunc("GET /beep", h.beep)
	mux.Handle("GET /css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	mux.Handle("GET /assets/", http.FileServer(http.Dir("./frontend/dist")))
	h.mux = mux
	h.prod = prod

	return h
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	type homeresp struct {
		Title string
		Prod  *Prod
	}
	homeT.Execute(w, homeresp{"home", h.prod})
}

func (h *Handler) beep(w http.ResponseWriter, r *http.Request) {

	type beepresp struct {
		Title string
		Prod  *Prod
	}
	beepT.Execute(w, beepresp{"beep", h.prod})
}

func (h *Handler) Serve() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.mux.ServeHTTP(w, r)
	})
}
