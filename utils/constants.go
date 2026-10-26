package utils

import (
	"fmt"
)

const BUMP_LIMIT = 1000
const REPLY_LIMIT = 1296
const THREADS_PER_INDEX_PAGE = 20
const THREADS_PER_CATALOG_PAGE = 100

var PURPLE = AToColorf("#452faa")

func MaxReplies(replyCount int) bool {
	return replyCount >= REPLY_LIMIT
}

func MaxBumps(replyCount int) bool {
	return replyCount >= BUMP_LIMIT
}

func PercentRemaining(replyCount *int) string {
	if replyCount == nil {
		return "100%"
	}
	rem := 100 * float64(REPLY_LIMIT-*replyCount) / float64(REPLY_LIMIT)
	return fmt.Sprintf("%f%%", rem)
}

// EVERY 6 base36 digit word is smaller than uint32max
// EVERY 7 base36 digit word with first letter > 1 is bigger than uint32max
var (
	PLAY_ID        = AToIDp("play")
	SKIP_ID        = AToIDp("skip")
	PAUSE_ID       = AToIDp("pause")
	SEEK_ID        = AToIDp("seek")
	NOSEEK_EX      = AToExp("noseek")
	YESSEEK_EX     = AToExp("yesseek")
	UNPAUSE_EX     = AToExp("unpause")
	DEBRAINWORM_EX = AToExp("debrainworm")
	DESH_ID        = AToIDp("desh")
	DESHELL_EX     = AToExp("deshell")
	MOLT_ID        = AToIDp("molt")
)
