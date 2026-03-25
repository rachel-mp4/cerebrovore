package utils

import (
	"fmt"
)

const BUMP_LIMIT = 1000
const REPLY_LIMIT = 1296
const THREADS_PER_INDEX_PAGE = 20
const THREADS_PER_CATALOG_PAGE = 100

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

var (
	PLAY_ID        = AToIDp("play")
	SKIP_ID        = AToIDp("skip")
	PAUSE_ID       = AToIDp("pause")
	UNPAUSE_EX     = AToExp("unpause")
	DEBRAINWORM_EX = AToExp("debrainworm")
	DESH_ID        = AToIDp("desh")
	DESHELL_EX     = AToExp("deshell")
	MOLT_ID        = AToIDp("molt")
)
