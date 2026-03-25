package handler

import (
	"net/http"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
)

func getNotes() []types.Patch {
	return []types.Patch{
		{
			Release:   "β.-0.4",
			Timestamp: "2026-03-24",
			Notes: []types.Note{
				{
					Type: "upgrade",
					Name: "resizing wormwatch now works horizontally (a bit jank still)",
				},
				{
					Type: "upgrade",
					Name: "alias lines with just #play to #unpause",
				},
				{
					Type: "fix",
					Name: "refresh now seeks to current timestamp properly",
				},
				{
					Type: "feature",
					Name: "guesses if autoplay works and mutes videos accordingly with info message",
				},
			},
		},
		{
			Release:   "β.-0.3",
			Timestamp: "2026-03-23",
			Notes: []types.Note{
				{
					Type: "feature",
					Name: "click and drag wormwatch to resize it",
				},
				{
					Type: "upgrade",
					Name: "skip during pause should work right now",
				},
			},
		},
		{
			Release:   "β.-0.2",
			Timestamp: "2026-03-23",
			Notes: []types.Note{
				{
					Type: "fix",
					Name: "watch thread query typo",
				},
				{
					Type: "upgrade",
					Name: "hide invite codes in UI if identity provider does not use them",
				},
			},
		},
		{
			Release:   "β.-0.1",
			Timestamp: "2026-03-21",
			Notes: []types.Note{
				{
					Type: "feature",
					Name: "grey out archived and bump limit threads in index/catalog",
				},
				{
					Type: "feature",
					Name: "warn browsers that don't support masonry",
				},
				{
					Type: "feature",
					Name: "add version number to homepage",
				},
				{
					Type: "upgrade",
					Name: "improve queries for catalog",
				},
			},
		},
		{
			Release:   "β.-0.0",
			Timestamp: "2026-03-20",
			Notes: []types.Note{
				{
					Type:        "release",
					Name:        "this is beta prelease zero",
					Description: "that's why it's negative zero",
				},
				{
					Type: "upgrade",
					Name: "thread watcher no longer requires refresh to update preferences",
				},
				{
					Type: "feature",
					Name: "thread watcher informs you of new threads",
				},
				{
					Type: "feature",
					Name: "invite codes and identity provider service",
				},
				{
					Type: "feature",
					Name: "index page shows last 5 posts + op",
				},
				{
					Type: "feature",
					Name: "add catalog",
				},
				{
					Type: "feature",
					Name: "preview pasted image in thread",
				},
			},
		},
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
	}
}

func (h *Handler) patchnotes(c *Client, w http.ResponseWriter, r *http.Request) {
	if c == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	type patchresp struct {
		baseresp
		Patches []types.Patch
	}
	base, err := h.makebase("patch notes", r.Context())
	if err != nil {
		clog.Warn("bumps %s", err)
	}
	err = patchT.ExecuteTemplate(w, "base", patchresp{
		*base,
		h.notes,
	})
	if err != nil {
		clog.Warn("patch notes: %s", err)
		http.Error(w, "error templating", http.StatusInternalServerError)
	}

}
