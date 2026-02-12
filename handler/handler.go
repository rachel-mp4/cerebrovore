package handler

import (
	"github.com/rachel-mp4/cerebrovore/model"
	"log"
	"net/http"
)

type Handler struct {
	mux  *http.ServeMux
	prod *Prod
	m    *model.Model
}

type Prod struct {
	ChatPath string
	BeepPath string
}

func NewHandler(prod *Prod, m *model.Model) Handler {
	h := Handler{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.home)
	mux.HandleFunc("GET /beep", h.beep)
	mux.HandleFunc("POST /t", h.postThread)
	mux.HandleFunc("GET /t/{ntid}", h.getThread)
	mux.HandleFunc("GET /t/{ntid}/ws", h.getThreadWS)
	mux.Handle("GET /css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css"))))
	mux.Handle("GET /assets/", http.FileServer(http.Dir("./frontend/dist")))
	h.mux = mux
	h.prod = prod
	h.m = m

	return h
}

func (h *Handler) home(w http.ResponseWriter, r *http.Request) {
	type homeresp struct {
		Title   string
		Threads []ThreadResp
	}
	err := homeT.Execute(w, homeresp{"home",
		constructThreadsResp(h.m.GetThreads()),
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}
}

func (h *Handler) beep(w http.ResponseWriter, r *http.Request) {

	type beepresp struct {
		Title   string
		Prod    *Prod
		Threads []ThreadResp
	}
	beepT.Execute(w, beepresp{"beep",
		h.prod,
		constructThreadsResp(h.m.GetThreads()),
	})
}

func (h *Handler) Serve() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.mux.ServeHTTP(w, r)
	})
}
