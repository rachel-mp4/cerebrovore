package handler

import (
	"net/http"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
)

func getNotes() []types.Patch {
	return []types.Patch{
		{
			Release:   "β.5.6",
			Timestamp: "2026-06-03",
			Notes: []types.Note{
				{
					Type: "feature",
					Name: "manually archive your threads, if you wish",
				},
				{
					Type: "upgrade",
					Name: "sticky topic for threads",
				},
			},
		},
		{
			Release:   "β.5.5",
			Timestamp: "2026-06-02",
			Notes: []types.Note{
				{
					Type: "upgrade",
					Name: "new thread page now specifies if it's hrt or forum, and site adapts according to where you are when you navigate to new thread",
				},
			},
		},
		{
			Release:   "β.5.4",
			Timestamp: "2026-06-01",
			Notes: []types.Note{
				{
					Type: "upgrade",
					Name: "don't send notifications if you have a thread open",
				},
				{
					Type: "upgrade",
					Name: "paste images to new thread form",
				},
				{
					Type: "upgrade",
					Name: "alt text also rendered as title (xkcd-style hover) text",
				},
				{
					Type: "upgrade",
					Name: "empty posts don't get saved to database and disappear for anyone who's connected",
				},
				{
					Type: "fix",
					Name: "(you)s within a thread should always render now",
				},
			},
		},
		{
			Release:   "β.5.3",
			Timestamp: "2026-05-31",
			Notes: []types.Note{
				{
					Type: "upgrade",
					Name: "added diffs back",
				},
			},
		},
		{
			Release:   "β.5.2",
			Timestamp: "2026-05-24",
			Notes: []types.Note{
				{
					Type:        "fix",
					Name:        "more frontend optimizations",
					Description: "in particular it should now immediately, regardless of ping, show your post. if you can avoid it, don't send it before you get an id number - it works! spent maybe too much time making a nice system! but there's not enough info at that point in time to add it to database. if you have a post id, it's all good!",
				},
				{
					Type:        "fix",
					Name:        "initial setting for wormwatch now is enable, even if you open settings menu",
					Description: "i hope",
				},
			},
		},
		{
			Release:   "β.5.1",
			Timestamp: "2026-05-22",
			Notes: []types.Note{
				{
					Type:        "fix",
					Name:        "frontend optimizations!",
					Description: "please let me know if frontend still sucks for long threads, optimized scrolling the background text, and some css selectors",
				},
			},
		},
		{
			Release:   "β.5.0",
			Timestamp: "2026-05-17",
			Notes: []types.Note{
				{
					Type:        "feature",
					Name:        "new system for parsing messages",
					Description: "*toggles on and off italic, **toggles on and off bold, `toggles on and off code. additionally, >^<v at the start of a line quote text in various colors",
				},
				{
					Type: "upgrade",
					Name: "you can now delete your threads without mod intervention",
				},
				{
					Type: "upgrade",
					Name: "infinite scrolling in your inbox",
				},
			},
		},
		{
			Release:   "β.4.2",
			Timestamp: "2026-05-14",
			Notes: []types.Note{
				{
					Type: "fix",
					Name: "delete images that don't get linked to a post",
				},
				{
					Type: "upgrade",
					Name: "reload forum on post & scroll to bottom",
				},
			},
		},
		{
			Release:   "β.4.1",
			Timestamp: "2026-05-13",
			Notes: []types.Note{
				{
					Type: "upgrade",
					Name: "improve reporting flow for moderators",
				},
			},
		},
		{
			Release:   "β.4.0",
			Timestamp: "2026-05-08",
			Notes: []types.Note{
				{
					Type: "feature",
					Name: "notifications",
				},
			},
		},
		{
			Release:   "β.3.0",
			Timestamp: "2026-04-29",
			Notes: []types.Note{
				{
					Type: "feature",
					Name: "forums",
				},
				{
					Type: "feature",
					Name: "fill out new thread form a lil based on ur profile",
				},
			},
		},
		{
			Release:   "β.2.2",
			Timestamp: "2026-04-22",
			Notes: []types.Note{
				{
					Type: "feature",
					Name: "yous in thread",
				},
			},
		},
		{
			Release:   "β.2.1",
			Timestamp: "2026-04-21",
			Notes: []types.Note{
				{
					Type:        "feature",
					Name:        "delete your posts!",
					Description: "i don't have capacity to delete your threads yet, report it so a mod can take action",
				},
				{
					Type: "feature",
					Name: "not complete, but your posts now say (you) to help remind you if it's an anonymous post",
				},
				{
					Type: "change",
					Name: "timestamps look different now, sorry if you prefer the old way, ui getting busy",
				},
			},
		},
		{
			Release:   "β.2.0",
			Timestamp: "2026-04-20",
			Notes: []types.Note{
				{
					Type: "feature",
					Name: "capacity to report posts",
				},
				{
					Type: "feature",
					Name: "improve moderation tooling",
				},
				{
					Type: "feature",
					Name: "moderator role",
				},
				{
					Type: "upgrade",
					Name: "refresh session on access website",
				},
				{
					Type: "upgrade",
					Name: "show if user is debrainwormed on profile",
				},
				{
					Type: "fix",
					Name: "last seen no longer exposes posting in deleted threads",
				},
				{
					Type: "fix",
					Name: "ordering of featured accounts on profile is now preserved instead of being based on when they first logged in after the profile update",
				},
			},
		},
		{
			Release:   "β.1.1",
			Timestamp: "2026-04-11",
			Notes: []types.Note{
				{
					Type: "feature",
					Name: "setting to display ping in threads",
				},
			},
		},
		{
			Release:   "β.1.0",
			Timestamp: "2026-04-10",
			Notes: []types.Note{
				{
					Type: "release",
					Name: "beta 1.0 is out!",
				},
				{
					Type: "feature",
					Name: "profiles!",
				},
				{
					Type:        "downgrade :c",
					Name:        "remove a bit of freedom in usernames! sorry!",
					Description: "i had to coerce some pre-existing usernames to fit the new validation pattern. i'm doing this because profiles => usernames should be url safe, and perhaps at some point in the future, you might be able to have a profile page which is a subdomain. but that would actually require even more validation.... so this is the middleground for now, that makes nobody happy :)",
				},
				{
					Type:        "fix",
					Name:        "auto reload after disconnect",
					Description: "click anywhere to cancel! should be self-explanatory",
				},
				{
					Type: "upgrade",
					Name: "improve consistency around click to view fullsize image",
				},
				{
					Type: "meta",
					Name: "add invites to dev / community discord server",
				},
			},
		},
		{
			Release:   "β.0.1",
			Timestamp: "2026-04-05",
			Notes: []types.Note{
				{
					Type: "feature",
					Name: "setting for min and max font size in thread",
				},
				{
					Type: "fix",
					Name: "correct maxlength in new thread topic field in form",
				},
				{
					Type: "upgrade",
					Name: "add details to site footer",
				},
			},
		},
		{
			Release:   "β.0.0",
			Timestamp: "2026-04-03",
			Notes: []types.Note{
				{
					Type: "release",
					Name: "we are now in beta",
				},
				{
					Type: "feature",
					Name: "invite codes",
				},
				{
					Type: "feature",
					Name: "bans",
				},
				{
					Type: "feature",
					Name: "ban appeals",
				},
				{
					Type:        "feature",
					Name:        "self bans",
					Description: "#desh, #molt, #deshell, #debrainworm",
				},
				{
					Type: "feature",
					Name: "setting to disable notifications for new threads",
				},
				{
					Type: "feature",
					Name: "removed ai agent",
				},
				{
					Type: "upgrade",
					Name: "cut down on extra websockets",
				},
				{
					Type: "upgrade",
					Name: "refresh bumped threads on refocus",
				},
			},
		},
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
	base, err := h.makebase("patch notes", c, r.Context())
	if err != nil {
		clog.Warn("bumps %s", err)
	}
	patchT.exec(w, patchresp{
		base,
		h.notes,
	})
}
