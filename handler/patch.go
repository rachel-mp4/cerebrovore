package handler

import (
	"log"
	"net/http"

	"github.com/rachel-mp4/cerebrovore/types"
)

func (h *Handler) patchnotes(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	type patchresp struct {
		baseresp
		Patches []types.Patch
	}
	tt, err := h.db.GetBumps(r.Context())
	if err != nil {
		http.Error(w, "error getting threads", http.StatusInternalServerError)
		return
	}
	err = patchT.ExecuteTemplate(w, "base", patchresp{
		baseresp{
			h.ca,
			"brainworm",
			tt,
		},
		[]types.Patch{
			{
				Release:   "α.1.0",
				Timestamp: "2026-03-13",
				Notes: []types.Note{
					{
						Type: "feature",
						Name: "wormwatch current video timestamp",
					},
					{
						Type: "feature",
						Name: "post timestamp",
					},
					{
						Type: "feature",
						Name: "reply & watch notification sfx",
					},
					{
						Type:        "feature",
						Name:        "patch notes",
						Description: "you're reading them rn!",
					},
				},
			},
			{
				Release:   "α.0.1",
				Timestamp: "2026-03-12",
				Notes: []types.Note{
					{
						Type:        "fix",
						Name:        "critical security issue",
						Description: "fail to validate blob content-id leads to serving arbitrary filepath",
					},
				},
			},
			{
				Release:   "α.0.0",
				Timestamp: "2026-03-11",
				Notes: []types.Note{
					{
						Type: "release",
						Name: "initial release yay!!!",
					},
				},
			},
		},
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}

}
