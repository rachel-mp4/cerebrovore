package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
)

func (m *MockStore) CreateThread(thread *types.Thread, ctx context.Context) error {
	clog.Dbug("create thread: %s", thread.String())
	return nil
}

func (s *Store) CreateThread(thread *types.Thread, ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		clog.Warn("db: %s", err)
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		INSERT INTO threads (id, topic, dead)
		VALUES ($1, $2, $3)
		`, thread.ID, thread.Topic, thread.Dead)
	if err != nil {
		clog.Warn("db: %s", err)
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO posts (id, thread_id, username, anon, nick, color)
		VALUES ($1, $1, $2, $3, $4, $5)
		`, thread.OP.ID, thread.OP.Username, thread.OP.Anon, thread.OP.Nick, thread.OP.Color)
	if err != nil {
		clog.Warn("db: %s", err)
		return err
	}
	if thread.OP.TextContent != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO text_posts (post_id, body)
			VALUES ($1, $2)
			`, thread.ID, thread.OP.TextContent.Body)
		if err != nil {
			clog.Warn("db: %s", err)
			return err
		}
	}
	if thread.OP.ImageContent != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO image_posts (post_id, cid, alt)
			VALUES ($1, $2, $3)
			`, thread.OP.ID, thread.OP.ImageContent.CID, thread.OP.ImageContent.Alt)
		if err != nil {
			clog.Warn("db: %s", err)
			return err
		}
	}

	if len(thread.OP.Backlinks) > 0 {
		_, err = tx.Exec(ctx, `
			INSERT INTO pending_post_replies (from_id, to_id)
			SELECT $1, unnest($2::int[])
			ON CONFLICT DO NOTHING
			`, thread.ID, thread.OP.Backlinks)
		if err != nil {
			clog.Warn("db: %s", err)
			return err
		}
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO post_replies (from_id, to_id)
		SELECT p.from_id, p.to_id
		FROM pending_post_replies p
		JOIN posts target ON target.id = p.to_id
		WHERE p.from_id = $1 OR p.to_id = $1
		ON CONFLICT DO NOTHING
		`, thread.ID)
	if err != nil {
		clog.Warn("db: %s", err)
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM pending_post_replies p
		USING posts target
		WHERE p.to_id = target.id
		AND (p.from_id = $1 OR p.to_id = $1)
		`, thread.ID)
	if err != nil {
		clog.Warn("db: %s", err)
		return err
	}
	return tx.Commit(ctx)
}

func (m *MockStore) GetAllThreads(ctx context.Context) ([]types.Thread, error) {
	return nil, nil
}

func (s *Store) GetAllThreads(ctx context.Context) ([]types.Thread, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			id,
			topic,
			reply_count,
			dead,
			manually_archived
		FROM threads
		WHERE deleted = FALSE
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tt := make([]types.Thread, 0)
	for rows.Next() {
		var t types.Thread
		rows.Scan(&t.ID, &t.Topic, &t.ReplyCount, &t.Dead, &t.ManuallyArchived)
		tt = append(tt, t)
	}
	return tt, nil
}

func (m *MockStore) GetRecentThreads(before *uint32, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *uint32, err error) {
	return nil, nil, nil
}

func (m *MockStore) GetRecentDeadThreads(before *uint32, limit int, archived bool, ctx context.Context) (threads []types.ForumThreadThumb, cursor *uint32, err error) {
	return nil, nil, nil
}

func (s *Store) GetRecentThreads(before *uint32, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *uint32, err error) {
	q := `
		WITH limited_threads AS (
			SELECT id, topic, bumped_at, reply_count, manually_archived
			FROM threads
			WHERE deleted = FALSE AND dead = FALSE %s %s
			ORDER BY id DESC
			LIMIT $1
		)
		SELECT
			t.id,
			t.topic,
			t.reply_count,
			t.manually_archived,
			p.id,
			p.nick,
			p.username,
			p.anon,
			p.color,
			p.posted_at,
			tp.body,
			ip.cid,
			ip.alt,
			COALESCE(pr.replies, '{}') AS replies
		FROM limited_threads t
		LEFT JOIN LATERAL (
			SELECT *
			FROM (
				SELECT
					p.*,
					ROW_NUMBER() OVER (ORDER BY p.id ASC) AS rn_asc,
					ROW_NUMBER() OVER (ORDER BY p.id DESC) AS rn_desc
				FROM posts p
				WHERE p.thread_id = t.id AND p.deleted = FALSE
			) ranked
			WHERE rn_asc = 1 OR rn_desc <= 5
			ORDER BY id ASC
		) p ON TRUE
		LEFT JOIN text_posts tp ON tp.post_id = p.id
		LEFT JOIN image_posts ip ON ip.post_id = p.id
		LEFT JOIN (
			SELECT pr.to_id, array_agg(pr.from_id) AS replies
			FROM post_replies pr
			JOIN posts rp ON rp.id = pr.from_id
			WHERE rp.deleted = FALSE
			GROUP BY pr.to_id
		) pr ON pr.to_id = p.id
		ORDER BY t.id DESC, p.id ASC
	`
	var rows pgx.Rows
	astring := "AND manually_archived = FALSE"
	if archived {
		astring = ""
	}
	if before != nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "AND id < $2", astring), limit+1, *before)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "", astring), limit+1)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer rows.Close()
	threads = make([]types.Thread, 0)
	var lastthread *types.Thread
	i := 0
	for rows.Next() {
		var thread types.Thread
		var post types.Post
		var body *string
		var alt *string
		var cid *string
		err = rows.Scan(&thread.ID, &thread.Topic, &thread.ReplyCount, &thread.ManuallyArchived, &post.ID, &post.Nick, &post.Username, &post.Anon, &post.Color, &post.PostedAt, &body, &cid, &alt, &post.Backlinks)
		if err != nil {
			return nil, nil, err
		}
		if body != nil {
			post.TextContent = &types.TextContent{Body: *body}
		}
		if cid != nil {
			post.ImageContent = &types.ImageContent{CID: *cid, Alt: alt}
		}
		if lastthread == nil {
			thread.Posts = []types.Post{post}
			lastthread = &thread
		} else {
			if lastthread.ID == thread.ID {
				lastthread.Posts = append(lastthread.Posts, post)
			} else {
				lastthread.OP = lastthread.Posts[0]
				lastthread.Posts = lastthread.Posts[1:]
				threads = append(threads, *lastthread)
				thread.Posts = []types.Post{post}
				lastthread = &thread
				i = i + 1
				if i == limit {
					break
				}
			}
		}
		cursor = &thread.ID
	}
	if i != limit {
		cursor = nil
		threads = append(threads, *lastthread)
	}
	return threads, cursor, nil
}

func (s *Store) GetRecentDeadThreads(before *uint32, limit int, archived bool, ctx context.Context) (threads []types.ForumThreadThumb, cursor *uint32, err error) {
	q := `
		WITH limited_threads AS (
			SELECT id, topic, bumped_at, reply_count, manually_archived
			FROM threads
			WHERE deleted = FALSE AND dead = TRUE %s %s
			ORDER BY id DESC
			LIMIT $1
		)
		SELECT
			t.id,
			t.topic,
			t.reply_count,
			t.manually_archived,
			p.id,
			p.nick,
			p.username,
			p.anon,
			p.color,
			p.posted_at
		FROM limited_threads t
		LEFT JOIN LATERAL (
			SELECT *
			FROM (
				SELECT
					p.*,
					ROW_NUMBER() OVER (ORDER BY p.id ASC) AS rn_asc,
					ROW_NUMBER() OVER (ORDER BY p.id DESC) AS rn_desc
				FROM posts p
				WHERE p.thread_id = t.id AND p.deleted = FALSE
			) ranked
			WHERE rn_asc = 1 OR rn_desc = 1
			ORDER BY id ASC
		) p ON TRUE
		ORDER BY t.id DESC, p.id ASC
	`
	astring := "AND manually_archived = FALSE"
	if archived {
		astring = ""
	}
	var rows pgx.Rows
	if before != nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "AND id < $2", astring), limit+1, *before)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "", astring), limit+1)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer rows.Close()
	var lastthread *types.ForumThreadThumb
	i := 0
	for rows.Next() {
		thread := types.ForumThreadThumb{}
		var post types.Post
		err = rows.Scan(&thread.ID, &thread.Topic, &thread.ReplyCount, &thread.ManuallyArchived, &post.ID, &post.Nick, &post.Username, &post.Anon, &post.Color, &post.PostedAt)
		if err != nil {
			return nil, nil, err
		}
		if lastthread == nil {
			thread.OP = post
			lastthread = &thread
		} else {
			if lastthread.ID == thread.ID {
				lastthread.LP = &post
			}
			threads = append(threads, *lastthread)
			i = i + 1
			if i == limit {
				break
			}
			if lastthread.ID == thread.ID {
				lastthread = nil
				continue
			}
			thread.OP = post
			lastthread = &thread
		}
		cursor = &thread.ID
	}
	if i != limit {
		cursor = nil
		if lastthread != nil {
			threads = append(threads, *lastthread)
		}
	}
	return threads, cursor, nil
}

func (m *MockStore) GetBumpedThreads(before *time.Time, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *time.Time, err error) {
	return nil, nil, nil
}
func (m *MockStore) GetBumpedDeadThreads(before *time.Time, limit int, archived bool, ctx context.Context) (threads []types.ForumThreadThumb, cursor *time.Time, err error) {
	return nil, nil, nil
}

func (s *Store) GetBumps(ctx context.Context) (threads []types.Thread, err error) {
	var rows pgx.Rows
	rows, err = s.pool.Query(ctx, `
		SELECT
			id,
			topic
		FROM threads
		WHERE deleted = FALSE
		AND reply_count + 1 < $1 AND manually_archived = FALSE
		ORDER BY bumped_at DESC
		LIMIT 5
		`, utils.REPLY_LIMIT)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()
	threads = make([]types.Thread, 0, 5)
	for rows.Next() {
		var thread types.Thread
		err = rows.Scan(&thread.ID, &thread.Topic)
		if err != nil {
			return nil, nil
		}
		threads = append(threads, thread)
	}
	return threads, nil
}

func (m *MockStore) GetBumps(ctx context.Context) (threads []types.Thread, err error) {
	return nil, nil
}

func (s *Store) GetBumpedThreads(before *time.Time, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *time.Time, err error) {
	q := `
		WITH limited_threads AS (
			SELECT id, topic, bumped_at, reply_count, manually_archived
			FROM threads
			WHERE deleted = FALSE AND dead = FALSE %s %s
			ORDER BY bumped_at DESC
			LIMIT $1
		)
		SELECT
			t.id,
			t.topic,
			t.bumped_at,
			t.reply_count,
			t.manually_archived,
			p.id,
			p.nick,
			p.username,
			p.anon,
			p.color,
			p.posted_at,
			tp.body,
			ip.cid,
			ip.alt,
			COALESCE(pr.replies, '{}') AS replies
		FROM limited_threads t
		LEFT JOIN LATERAL (
			SELECT *
			FROM (
				SELECT
					p.*,
					ROW_NUMBER() OVER (ORDER BY p.id ASC) AS rn_asc,
					ROW_NUMBER() OVER (ORDER BY p.id DESC) AS rn_desc
				FROM posts p
				WHERE p.thread_id = t.id AND p.deleted = FALSE
			) ranked
			WHERE rn_asc = 1 OR rn_desc <= 5
			ORDER BY id ASC
		) p ON TRUE
		LEFT JOIN text_posts tp ON tp.post_id = p.id
		LEFT JOIN image_posts ip ON ip.post_id = p.id
		LEFT JOIN (
			SELECT pr.to_id, array_agg(pr.from_id) AS replies
			FROM post_replies pr
			JOIN posts rp ON rp.id = pr.from_id
			WHERE rp.deleted = FALSE
			GROUP BY pr.to_id
		) pr ON pr.to_id = p.id
		ORDER BY bumped_at DESC, p.id ASC
	`
	astring := "AND manually_archived = FALSE"
	if archived {
		astring = ""
	}
	var rows pgx.Rows
	if before != nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "AND bumped_at < $2", astring), limit+1, *before)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "", astring), limit+1)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer rows.Close()
	threads = make([]types.Thread, 0)
	var lastthread *types.Thread
	i := 0
	for rows.Next() {
		var thread types.Thread
		var bumpt time.Time
		var post types.Post
		var body *string
		var cid *string
		var alt *string
		err = rows.Scan(&thread.ID, &thread.Topic, &bumpt, &thread.ReplyCount, &thread.ManuallyArchived, &post.ID, &post.Nick, &post.Username, &post.Anon, &post.Color, &post.PostedAt, &body, &cid, &alt, &post.Backlinks)
		if err != nil {
			return nil, nil, err
		}
		if body != nil {
			post.TextContent = &types.TextContent{Body: *body}
		}
		if cid != nil {
			post.ImageContent = &types.ImageContent{CID: *cid, Alt: alt}
		}
		if lastthread == nil {
			thread.Posts = []types.Post{post}
			lastthread = &thread
		} else {
			if lastthread.ID == thread.ID {
				lastthread.Posts = append(lastthread.Posts, post)
			} else {
				lastthread.OP = lastthread.Posts[0]
				lastthread.Posts = lastthread.Posts[1:]
				threads = append(threads, *lastthread)
				thread.Posts = []types.Post{post}
				lastthread = &thread
				i = i + 1
				if i == limit {
					break
				}
			}
		}
		cursor = &bumpt
	}
	// this means we ran out of threads before we hit limit, so we
	// didn't add the lastthread we were creating to the threads,
	// and there should be NO cursor
	if i != limit {
		cursor = nil
		threads = append(threads, *lastthread)
	}
	return threads, cursor, nil
}

func (s *Store) GetBumpedDeadThreads(before *time.Time, limit int, archived bool, ctx context.Context) (threads []types.ForumThreadThumb, cursor *time.Time, err error) {
	q := `
		WITH limited_threads AS (
			SELECT id, topic, bumped_at, reply_count, manually_archived
			FROM threads
			WHERE deleted = FALSE AND dead = TRUE %s %s
			ORDER BY bumped_at DESC
			LIMIT $1
		)
		SELECT
			t.id,
			t.topic,
			t.bumped_at,
			t.reply_count,
			t.manually_archived,
			p.id,
			p.nick,
			p.username,
			p.anon,
			p.color,
			p.posted_at
		FROM limited_threads t
		LEFT JOIN LATERAL (
			SELECT *
			FROM (
				SELECT
					p.*,
					ROW_NUMBER() OVER (ORDER BY p.id ASC) AS rn_asc,
					ROW_NUMBER() OVER (ORDER BY p.id DESC) AS rn_desc
				FROM posts p
				WHERE p.thread_id = t.id AND p.deleted = FALSE
			) ranked
			WHERE rn_asc = 1 OR rn_desc = 1
			ORDER BY id ASC
		) p ON TRUE
		ORDER BY bumped_at DESC, p.id ASC
	`
	astring := "AND manually_archived = FALSE"
	if archived {
		astring = ""
	}
	var rows pgx.Rows
	if before != nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "AND bumped_at < $2", astring), limit+1, *before)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "", astring), limit+1)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer rows.Close()
	var lastthread *types.ForumThreadThumb
	i := 0
	for rows.Next() {
		thread := types.ForumThreadThumb{}
		var bumpt time.Time
		var post types.Post
		err = rows.Scan(&thread.ID, &thread.Topic, &bumpt, &thread.ReplyCount, &thread.ManuallyArchived, &post.ID, &post.Nick, &post.Username, &post.Anon, &post.Color, &post.PostedAt)
		if err != nil {
			return nil, nil, err
		}
		if lastthread == nil {
			thread.OP = post
			lastthread = &thread
		} else {
			if lastthread.ID == thread.ID {
				lastthread.LP = &post
			}
			threads = append(threads, *lastthread)
			i = i + 1
			if i == limit {
				break
			}
			if lastthread.ID == thread.ID {
				lastthread = nil
				continue
			}
			thread.OP = post
			lastthread = &thread
		}
		cursor = &bumpt
	}
	// this means we ran out of threads before we hit limit, so we
	// didn't add the lastthread we were creating to the threads,
	// and there should be NO cursor
	if i != limit {
		cursor = nil
		if lastthread != nil {
			threads = append(threads, *lastthread)
		}
	}
	return threads, cursor, nil
}

func (m *MockStore) GetBumpedCatalog(before *time.Time, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *time.Time, err error) {
	return nil, nil, nil
}

func (s *Store) GetBumpedCatalog(before *time.Time, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *time.Time, err error) {
	q := `
		SELECT
			t.id,
			t.topic,
			t.bumped_at,
			t.reply_count,
			t.manually_archived,
			p.id,
			p.nick,
			p.username,
			p.anon,
			p.color,
			p.posted_at,
			tp.body,
			ip.cid,
			ip.alt
		FROM threads t
		LEFT JOIN posts p ON p.id = t.id
		LEFT JOIN text_posts tp ON tp.post_id = p.id
		LEFT JOIN image_posts ip ON ip.post_id = p.id
		WHERE t.deleted = FALSE AND t.dead = false %s %s
		ORDER BY bumped_at DESC
		LIMIT $1
	`
	astring := "AND t.manually_archived = FALSE"
	if archived {
		astring = ""
	}
	var rows pgx.Rows
	if before != nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "AND bumped_at < $2", astring), limit+1, *before)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "", astring), limit+1)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer rows.Close()
	threads = make([]types.Thread, 0)
	i := 0
	for rows.Next() {
		i = i + 1
		if i == limit+1 {
			break
		}
		var thread types.Thread
		var bumpt time.Time
		var post types.Post
		var body *string
		var cid *string
		var alt *string
		err = rows.Scan(&thread.ID, &thread.Topic, &bumpt, &thread.ReplyCount, &thread.ManuallyArchived, &post.ID, &post.Nick, &post.Username, &post.Anon, &post.Color, &post.PostedAt, &body, &cid, &alt)
		if err != nil {
			return nil, nil, err
		}
		if body != nil {
			post.TextContent = &types.TextContent{Body: *body}
		}
		if cid != nil {
			post.ImageContent = &types.ImageContent{CID: *cid, Alt: alt}
		}
		thread.OP = post
		threads = append(threads, thread)
		cursor = &bumpt
	}
	// this means we ran out of threads before we hit limit
	// so there should be NO cursor
	if i != limit+1 {
		cursor = nil
	}
	return threads, cursor, nil
}

func (m *MockStore) GetRecentCatalog(before *uint32, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *uint32, err error) {
	return nil, nil, nil
}

func (s *Store) GetRecentCatalog(before *uint32, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *uint32, err error) {
	q := `
		SELECT
			t.id,
			t.topic,
			t.reply_count,
			t.manually_archived,
			p.id,
			p.nick,
			p.username,
			p.anon,
			p.color,
			p.posted_at,
			tp.body,
			ip.cid,
			ip.alt
		FROM threads t
		LEFT JOIN posts p ON p.id = t.id
		LEFT JOIN text_posts tp ON tp.post_id = p.id
		LEFT JOIN image_posts ip ON ip.post_id = p.id
		WHERE t.deleted = FALSE AND t.dead = FALSE %s %s
		ORDER BY t.id DESC
		LIMIT $1
	`
	astring := "AND t.manually_archived = FALSE"
	if archived {
		astring = ""
	}
	var rows pgx.Rows
	if before != nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "AND bumped_at < $2", astring), limit+1, *before)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "", astring), limit+1)
	}
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer rows.Close()
	threads = make([]types.Thread, 0)
	i := 0
	for rows.Next() {
		i = i + 1
		if i == limit+1 {
			break
		}
		var thread types.Thread
		var post types.Post
		var body *string
		var cid *string
		var alt *string
		err = rows.Scan(&thread.ID, &thread.Topic, &thread.ReplyCount, &thread.ManuallyArchived, &post.ID, &post.Nick, &post.Username, &post.Anon, &post.Color, &post.PostedAt, &body, &cid, &alt)
		if err != nil {
			return nil, nil, err
		}
		if body != nil {
			post.TextContent = &types.TextContent{Body: *body}
		}
		if cid != nil {
			post.ImageContent = &types.ImageContent{CID: *cid, Alt: alt}
		}
		thread.OP = post
		threads = append(threads, thread)
		cursor = &post.ID
	}
	// this means we ran out of threads before we hit limit
	// so there should be NO cursor
	if i != limit+1 {
		cursor = nil
	}
	return threads, cursor, nil
}

func (m *MockStore) GetThread(id uint32, viewerIsMod bool, viewerUsername string, ctx context.Context) (*types.Thread, error) {
	return nil, nil
}

func (m *MockStore) GetDeadThread(id uint32, viewerIsMod bool, viewerUsername string, ctx context.Context) (*types.Thread, error) {
	return nil, nil
}

func (s *Store) GetThread(id uint32, viewerIsMod bool, viewerUsername string, ctx context.Context) (thread *types.Thread, err error) {
	thread = &types.Thread{ID: id}
	row := s.pool.QueryRow(ctx, "SELECT topic, reply_count, manually_archived FROM threads WHERE id=$1 AND deleted=FALSE AND dead=FALSE", id)
	err = row.Scan(&thread.Topic, &thread.ReplyCount, &thread.ManuallyArchived)
	if err != nil {
		clog.Warn("db: %s", err)
		return nil, err
	}
	q := `
	SELECT 
		p.id,
		p.username,
		p.anon,
		p.nick,
		p.color,
		p.posted_at,
		t.body,
		i.cid,
		i.alt,
		COALESCE(pr.replies, '{}') AS replies
	FROM posts p
	LEFT JOIN text_posts t ON t.post_id = p.id
	LEFT JOIN image_posts i ON i.post_id = p.id
	LEFT JOIN (
		SELECT pr.to_id, array_agg(pr.from_id) AS replies
		FROM post_replies pr
		JOIN posts rp ON rp.id = pr.from_id
		WHERE rp.deleted = FALSE
		GROUP BY pr.to_id
	) pr ON pr.to_id = p.id
	WHERE p.thread_id = $1 AND p.deleted = FALSE
	ORDER BY p.id ASC
	`
	var rows pgx.Rows
	rows, err = s.pool.Query(ctx, q, id)
	if err != nil {
		clog.Warn("db: %s", err)
		return
	}
	defer rows.Close()
	thread.Posts = make([]types.Post, 0)
	for rows.Next() {
		p := types.Post{}
		var b *string
		var cid *string
		var alt *string
		err = rows.Scan(&p.ID, &p.Username, &p.Anon, &p.Nick, &p.Color, &p.PostedAt, &b, &cid, &alt, &p.Backlinks)
		if err != nil {
			clog.Warn("db: %s", err)
			return
		}
		if b != nil {
			p.TextContent = &types.TextContent{Body: *b}
		}
		if cid != nil {
			p.ImageContent = &types.ImageContent{CID: *cid, Alt: alt}
		}
		p.ViewerIsYou = viewerUsername == p.Username
		p.LinkToModerate = viewerIsMod
		thread.Posts = append(thread.Posts, p)
	}
	thread.OP = thread.Posts[0]
	return thread, nil
}

func (s *Store) GetDeadThread(id uint32, viewerIsMod bool, viewerUsername string, ctx context.Context) (thread *types.Thread, err error) {
	thread = &types.Thread{ID: id, Dead: true}
	row := s.pool.QueryRow(ctx, "SELECT topic, reply_count, manually_archived FROM threads WHERE id=$1 AND deleted=FALSE AND dead=TRUE", id)
	err = row.Scan(&thread.Topic, &thread.ReplyCount, &thread.ManuallyArchived)
	if err != nil {
		clog.Warn("db: %s", err)
		return nil, err
	}
	q := `
	SELECT 
		p.id,
		p.username,
		p.anon,
		p.nick,
		p.color,
		p.posted_at,
		t.body,
		i.cid,
		i.alt,
		COALESCE(pr.replies, '{}') AS replies,
		pro.display_name,
		pro.avatar,
		pro.is_pixel,
		pro.color,
		pro.status
	FROM posts p
	LEFT JOIN text_posts t ON t.post_id = p.id
	LEFT JOIN image_posts i ON i.post_id = p.id
	LEFT JOIN (
		SELECT pr.to_id, array_agg(pr.from_id) AS replies
		FROM post_replies pr
		JOIN posts rp ON rp.id = pr.from_id
		WHERE rp.deleted = FALSE
		GROUP BY pr.to_id
	) pr ON pr.to_id = p.id
	LEFT JOIN profiles pro ON pro.username = p.username
	WHERE p.thread_id = $1 AND p.deleted = FALSE
	ORDER BY p.id ASC
	`
	var rows pgx.Rows
	rows, err = s.pool.Query(ctx, q, id)
	if err != nil {
		clog.Warn("db: %s", err)
		return
	}
	defer rows.Close()
	thread.Posts = make([]types.Post, 0)
	for rows.Next() {
		p := types.Post{}
		var b *string
		var cid *string
		var alt *string
		var dname *string
		var ava *string
		var ispix *bool
		var color *uint32
		var status *string
		err = rows.Scan(&p.ID, &p.Username, &p.Anon, &p.Nick, &p.Color, &p.PostedAt, &b, &cid, &alt, &p.Backlinks, &dname, &ava, &ispix, &color, &status)
		if err != nil {
			clog.Warn("db: %s", err)
			return
		}
		if b != nil {
			p.TextContent = &types.TextContent{Body: *b}
		}
		if cid != nil {
			p.ImageContent = &types.ImageContent{CID: *cid, Alt: alt}
		}
		p.ViewerIsYou = viewerUsername == p.Username
		p.LinkToModerate = viewerIsMod
		if dname != nil || ava != nil || ispix != nil || color != nil || status != nil {
			pro := types.ProfileHead{Username: p.Username, DisplayName: dname, Avatar: ava, IsPixelArt: ispix, Color: color, Status: status}
			p.ProfileHead = &pro
		}
		thread.Posts = append(thread.Posts, p)
	}
	thread.OP = thread.Posts[0]
	return thread, nil
}

func (m *MockStore) StartWatchContext(username string, ctx context.Context) ([]uint32, error) {
	return nil, nil
}

func (s *Store) StartWatchContext(username string, ctx context.Context) ([]uint32, error) {
	rows, err := s.pool.Query(ctx, `
		UPDATE watched_threads
		SET notified = TRUE
		WHERE username = $1
		RETURNING thread_id
		`, username)
	if err != nil {
		return nil, err
	}
	res := make([]uint32, 0)
	defer rows.Close()
	for rows.Next() {
		var v uint32
		err := rows.Scan(&v)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return res, nil
}

func (s *Store) EndWatchContext(username string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
	UPDATE watched_threads
	SET notified = FALSE
	WHERE username = $1`, username)
	return err
}

func (m *MockStore) EndWatchContext(username string, ctx context.Context) error {
	return nil
}

func (m *MockStore) WatchThread(username string, id uint32, ctx context.Context) (changed bool, err error) {
	return true, nil
}

func (s *Store) WatchThread(username string, id uint32, ctx context.Context) (changed bool, err error) {
	bl, _, err := s.ThreadStatus(id, ctx)
	if err != nil || bl {
		return
	}
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO watched_threads (username, thread_id, notified) VALUES ($1, $2, true) ON CONFLICT DO NOTHING
		`, username, id)
	if err != nil {
		return
	}
	changed = tag.RowsAffected() > 0
	return
}

func (s *Store) UnwatchThread(username string, id uint32, ctx context.Context) (changed bool, err error) {
	bl, _, err := s.ThreadStatus(id, ctx)
	if err != nil || bl {
		return
	}
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM watched_threads WHERE username = $1 AND thread_id = $2
		`, username, id)
	if err != nil {
		return
	}
	changed = tag.RowsAffected() > 0
	return
}

func (m *MockStore) UnwatchThread(username string, id uint32, ctx context.Context) (changed bool, err error) {
	return true, nil
}

func (s *Store) IsWatched(username string, id uint32, ctx context.Context) bool {
	row := s.pool.QueryRow(ctx, `
		SELECT username FROM watched_threads WHERE username = $1 AND thread_id = $2
		`, username, id)
	err := row.Scan(&username)
	return err == nil
}

func (m *MockStore) IsWatched(username string, id uint32, ctx context.Context) bool {
	return false
}

func (s *Store) RemoveWatchersFor(id uint32, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM watched_threads WHERE thread_id = $1
		`, id)
	return err
}
func (m *MockStore) RemoveWatchersFor(id uint32, ctx context.Context) error {
	return nil
}

func (s *Store) DeleteThread(id uint32, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE threads SET deleted = TRUE WHERE id = $1
		`, id)
	return err
}

func (m *MockStore) DeleteThread(id uint32, ctx context.Context) error {
	return nil
}

func (s *Store) ThreadStatus(id uint32, ctx context.Context) (bumplimit bool, replylimit bool, err error) {
	row := s.pool.QueryRow(ctx, `
		SELECT reply_count, manually_archived
		FROM threads
		WHERE id = $1
		`, id)
	var rc int
	var ma bool
	err = row.Scan(&rc, &ma)
	if err != nil {
		return
	}
	if ma {
		bumplimit = true
		replylimit = true
		return
	}
	bumplimit = utils.MaxBumps(rc)
	replylimit = utils.MaxReplies(rc)
	return
}

func (m *MockStore) ThreadStatus(id uint32, ctx context.Context) (bumplimit bool, replylimit bool, err error) {
	return
}

func (s *Store) ThreadIsDead(id uint32, ctx context.Context) (bool, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT dead
		FROM threads
		WHERE id = $1
		`, id)
	var dead bool
	err := row.Scan(&dead)
	return dead, err
}

func (m *MockStore) ThreadIsDead(id uint32, ctx context.Context) (bool, error) {
	return false, nil
}

func (s *Store) ArchiveThread(id uint32, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `UPDATE threads SET manually_archived = TRUE WHERE id = $1`, id)
	return err
}

func (m *MockStore) ArchiveThread(id uint32, ctx context.Context) error {
	return nil
}
