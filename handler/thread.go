package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func (h *Handler) postThread(w http.ResponseWriter, r *http.Request) {
	tid := h.m.AddThread()
	ntid := strconv.FormatInt(int64(tid), 36)
	http.Redirect(w, r, fmt.Sprintf("/t/%s", ntid), http.StatusSeeOther)
}

func (h *Handler) getThread(w http.ResponseWriter, r *http.Request) {
	ntid := r.PathValue("ntid")
	tid, err := strconv.ParseInt(ntid, 36, 64)
	if err != nil {
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}
	err = h.m.GetThread(uint32(tid))
	if err != nil {
		http.Error(w, "thread does not exist", http.StatusNotFound)
		return
	}
	tt := h.m.GetThreads()
	title := ntid
	me := constructThreadResp(uint32(tid))
	threads := constructThreadsResp(tt)

	type getthreadresp struct {
		Title   string
		Me      ThreadResp
		Threads []ThreadResp
	}
	gtr := getthreadresp{
		Title:   title,
		Me:      me,
		Threads: threads,
	}
	err = threadT.Execute(w, gtr)
	if err != nil {
		http.Error(w, "error templating", http.StatusInternalServerError)
	}
}

type ThreadResp struct {
	TID   uint32
	NTID  string
	WSURL string
}

func constructThreadsResp(threads []uint32) []ThreadResp {
	res := make([]ThreadResp, len(threads))
	for i, v := range threads {
		res[i] = constructThreadResp(v)
	}
	return res
}

func constructThreadResp(thread uint32) ThreadResp {
	ntid := strconv.FormatInt(int64(thread), 36)
	return ThreadResp{
		TID:   thread,
		NTID:  ntid,
		WSURL: fmt.Sprintf("/t/%s/ws", ntid),
	}
}

func (h *Handler) getThreadWS(w http.ResponseWriter, r *http.Request) {
	ntid := r.PathValue("ntid")
	tid, err := strconv.ParseInt(ntid, 36, 64)
	if err != nil {
		http.Error(w, "invalid thread id", http.StatusBadRequest)
		return
	}
	err = h.m.GetThread(uint32(tid))
	if err != nil {
		http.Error(w, "thread does not exist", http.StatusNotFound)
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
