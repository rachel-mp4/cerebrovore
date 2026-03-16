package handler

import (
	"net/http"

	"github.com/rachel-mp4/cerebrovore/clog"
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
				Release:   "α.2.0",
				Timestamp: "2026-03-16",
				Notes: []types.Note{
					{
						Type: "feature",
						Name: "add volume settings",
					},
					{
						Type: "bts",
						Name: "add a cool start script to streamline development",
					},
					{
						Type: "fix",
						Name: "actually use base 36 encoding of reply count on index",
					},
				},
			},
			{
				Release:   "α.1.3",
				Timestamp: "2026-03-13",
				Notes: []types.Note{
					{
						Type: "fix",
						Name: "use base 36 encoding of reply count on index",
					},
					{
						Type: "bodge",
						Name: "remove post size scaling on mobile for now",
					},
					{
						Type: "feature",
						Name: "store nick and color in localStorage",
					},
					{
						Type: "fix",
						Name: "diff styles respect light vs dark post color",
					},
				},
			},
			{
				Release:   "α.1.2",
				Timestamp: "2026-03-13",
				Notes: []types.Note{
					{
						Type: "fix",
						Name: "stick to bottom when application of local diffs causes line break",
					},
				},
			},
			{
				Release:   "α.1.1",
				Timestamp: "2026-03-13",
				Notes: []types.Note{
					{
						Type: "fix",
						Name: "links that don't start with protocol now work",
					},
					{
						Type: "fix",
						Name: "diffing of recieved message with local copy now works",
					},
				},
			},
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
		clog.Warn("patch notes: %s", err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}

}
