package utils

const BUMP_LIMIT = 1000
const REPLY_LIMIT = 1296
const THREADS_PER_INDEX_PAGE = 20

func MaxReplies(replyCount int) bool {
	return replyCount >= REPLY_LIMIT
}
