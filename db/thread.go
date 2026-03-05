package db

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rachel-mp4/cerebrovore/types"
)

func (m *MockStore) CreateThread(thread *types.Thread, ctx context.Context) error {
	log.Println(thread.String())
	return nil
}

func (s *Store) CreateThread(thread *types.Thread, ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		INSERT INTO threads (id, topic)
		VALUES ($1, $2)
		`, thread.ID, thread.Topic)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO posts (id, thread_id, username, anon, nick, color)
		VALUES ($1, $1, $2, $3, $4, $5)
		`, thread.OP.ID, thread.OP.Username, thread.OP.Anon, thread.OP.Nick, thread.OP.Color)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if thread.OP.TextContent != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO text_posts (post_id, body)
			VALUES ($1, $2)
			`, thread.ID, thread.OP.TextContent.Body)
		if err != nil {
			log.Println(err.Error())
			return err
		}
	}
	if thread.OP.ImageContent != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO image_posts (post_id, cid, alt)
			VALUES ($1, $2, $3)
			`, thread.OP.ID, thread.OP.ImageContent.CID, thread.OP.ImageContent.Alt)
		if err != nil {
			log.Println(err.Error())
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
			log.Println(err.Error())
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
		log.Println(err.Error())
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM pending_post_replies p
		USING posts target
		WHERE p.to_id = target.id
		AND (p.from_id = $1 OR p.to_id = $1)
		`, thread.ID)
	if err != nil {
		log.Println(err.Error())
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
			topic
		FROM threads
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tt := make([]types.Thread, 0)
	for rows.Next() {
		var t types.Thread
		rows.Scan(&t.ID, &t.Topic)
		tt = append(tt, t)
	}
	return tt, nil
}

func (m *MockStore) GetRecentThreads(before *uint32, limit int, ctx context.Context) (threads []types.Thread, cursor *uint32, err error) {
	return nil, nil, nil
}

func (s *Store) GetRecentThreads(before *uint32, limit int, ctx context.Context) (threads []types.Thread, cursor *uint32, err error) {
	q := `
		SELECT
			id,
			topic
		FROM threads
		ORDER BY id DESC
		%s
		LIMIT $1
	`
	var rows pgx.Rows
	if before != nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "WHERE id < $2"), limit+1, *before)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, ""), limit+1)
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
		err = rows.Scan(&thread.ID, &thread.Topic)
		cursor = &thread.ID
		if err != nil {
			return nil, nil, err
		}
		threads = append(threads, thread)
	}
	if i != limit+1 {
		cursor = nil
	}
	return threads, cursor, nil
}

func (m *MockStore) GetBumpedThreads(before *time.Time, limit int, ctx context.Context) (threads []types.Thread, cursor *time.Time, err error) {
	return nil, nil, nil
}

func (s *Store) GetBumpedThreads(before *time.Time, limit int, ctx context.Context) (threads []types.Thread, cursor *time.Time, err error) {
	q := `
		SELECT
			id,
			topic,
			bumped_at
		FROM threads
		%s
		ORDER BY bumped_at DESC
		LIMIT $1
	`
	var rows pgx.Rows
	if before != nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "WHERE bumped_at < $2"), limit+1, *before)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, ""), limit+1)
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
		err = rows.Scan(&thread.ID, &thread.Topic, &bumpt)
		cursor = &bumpt
		if err != nil {
			return nil, nil, err
		}
		threads = append(threads, thread)
	}
	if i != limit+1 {
		cursor = nil
	}
	return threads, cursor, nil
}

func (m *MockStore) GetThread(id uint32, before *uint32, limit int, ctx context.Context) (*types.Thread, *uint32, error) {
	return nil, nil, nil
}

func (s *Store) GetThread(id uint32, before *uint32, limit int, ctx context.Context) (thread *types.Thread, cursor *uint32, err error) {
	thread = &types.Thread{ID: id}
	row := s.pool.QueryRow(ctx, "SELECT topic FROM threads WHERE id=$1", id)
	err = row.Scan(&thread.Topic)
	if err != nil {
		log.Println(err.Error())
		return nil, nil, err
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
		SELECT to_id, array_agg(from_id) AS replies
		FROM post_replies
		GROUP BY to_id
	) pr ON pr.to_id = p.id
	WHERE p.thread_id = $1 %s
	ORDER BY p.id DESC
	LIMIT $2
	`
	var rows pgx.Rows
	if before == nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, ""), id, limit+1)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "AND p.id < $3"), id, limit+1, *before)
	}
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer rows.Close()
	thread.Posts = make([]types.Post, 0, limit)
	i := 0
	for rows.Next() {

		i = i + 1
		if i == limit+1 {
			break
		}
		p := types.Post{}
		var b *string
		var cid *string
		var alt *string
		err = rows.Scan(&p.ID, &p.Username, &p.Anon, &p.Nick, &p.Color, &p.PostedAt, &b, &cid, &alt, &p.Backlinks)
		if err != nil {
			log.Println(err.Error())
			return
		}
		if b != nil {
			p.TextContent = &types.TextContent{Body: *b}
		}
		if cid != nil {
			p.ImageContent = &types.ImageContent{CID: *cid, Alt: alt}
		}
		thread.Posts = append(thread.Posts, p)
		cursor = &p.ID
	}
	if i != limit+1 {
		cursor = nil
	}
	slices.Reverse(thread.Posts)
	return thread, cursor, nil
}

func (m *MockStore) GetWatchedThreads(username string, ctx context.Context) ([]uint32, error) {
	return nil, nil
}

func (s *Store) GetWatchedThreads(username string, ctx context.Context) ([]uint32, error) {
	return nil, nil
}

func (m *MockStore) WatchThread(username string, id uint32, ctx context.Context) error {
	return nil
}

func (s *Store) WatchThread(username string, id uint32, ctx context.Context) error {
	return nil
}

func (s *Store) UnwatchThread(username string, id uint32, ctx context.Context) error {
	return nil
}

func (m *MockStore) UnwatchThread(username string, id uint32, ctx context.Context) error {

	return nil
}
