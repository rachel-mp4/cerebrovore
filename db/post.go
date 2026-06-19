package db

import (
	"context"

	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
)

func (m *MockStore) CreatePost(post *types.Post, ctx context.Context) (int, []Backlink, error) {
	clog.Dbug("create post: %s", post.String())
	return 0, nil, nil
}

func (s *Store) CreatePost(post *types.Post, ctx context.Context) (int, []Backlink, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		clog.Warn("db: %s", err)
		return 0, nil, err
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(ctx, `
		UPDATE threads
		SET
			reply_count = reply_count + 1,
			bumped_at = CASE
				WHEN reply_count + 1 < $2 THEN now()
				ELSE bumped_at
			END
		WHERE id = $1
		AND reply_count < $3
		AND manually_archived = FALSE
		RETURNING reply_count
		`, post.ThreadID, utils.BUMP_LIMIT, utils.REPLY_LIMIT)
	var rc int
	err = row.Scan(&rc)
	if err != nil {
		return 0, nil, err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO posts (id, thread_id, username, anon, nick, color)
		VALUES ($1, $2, $3, $4, $5, $6)
		`, post.ID, post.ThreadID, post.Username, post.Anon, post.Nick, post.Color)
	if err != nil {
		clog.Warn("db: %s", err)
		return 0, nil, err
	}
	if post.TextContent != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO text_posts (post_id, body)
			VALUES ($1, $2)
			`, post.ID, post.TextContent.Body)
		if err != nil {
			clog.Warn("db: %s", err)
			return 0, nil, err
		}
	}
	if post.ImageContent != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO image_posts (post_id, cid, alt)
			VALUES ($1, $2, $3)
			`, post.ID, post.ImageContent.CID, post.ImageContent.Alt)
		if err != nil {
			clog.Warn("db: %s", err)
			return 0, nil, err
		}
	}

	if len(post.Backlinks) > 0 {
		_, err = tx.Exec(ctx, `
			INSERT INTO pending_post_replies (from_id, to_id)
			SELECT $1, unnest($2::int[])
			ON CONFLICT DO NOTHING
			`, post.ID, post.Backlinks)
		if err != nil {
			clog.Warn("db: %s", err)
			return 0, nil, err
		}
	}
	rows, err := tx.Query(ctx, `
		WITH inserted AS (
			INSERT INTO post_replies (from_id, to_id)
			SELECT p.from_id, p.to_id
			FROM pending_post_replies p
			JOIN posts target ON target.id = p.to_id
			WHERE p.from_id = $1 OR p.to_id = $1
			ON CONFLICT DO NOTHING
			RETURNING from_id, to_id
		)
		SELECT i.from_id, i.to_id, p.username
		FROM inserted i 
		JOIN POSTS p ON p.id = to_id
		`, post.ID)
	if err != nil {
		clog.Warn("db: %s", err)
		return 0, nil, err
	}
	defer rows.Close()
	res := make([]Backlink, 0)
	for rows.Next() {
		var bl Backlink
		err := rows.Scan(&bl.From, &bl.To, &bl.ToUsername)
		if err != nil {
			clog.Warn("db: %s", err)
			return 0, nil, err
		}
		res = append(res, bl)
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM pending_post_replies p
		USING posts target
		WHERE p.to_id = target.id
		AND (p.from_id = $1 OR p.to_id = $1)
		`, post.ID)
	if err != nil {
		clog.Warn("db: %s", err)
		return 0, nil, err
	}
	return rc, res, tx.Commit(ctx)
}

func (m *MockStore) EZPost(id uint32, ctx context.Context) (*types.Post, error) {
	return nil, nil
}

func (s *Store) EZPost(id uint32, ctx context.Context) (*types.Post, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT
			thread_id,
			username,
			nick
		FROM posts
		WHERE id = $1
		`, id)
	p := types.Post{ID: id}
	err := row.Scan(&p.ThreadID, &p.Username, &p.Nick)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetPost(id uint32, ctx context.Context) (*types.Post, error) {
	row := s.pool.QueryRow(ctx, `
	SELECT 
		p.thread_id,
		p.username,
		p.anon,
		p.nick,
		p.color,
		p.posted_at,
		p.deleted,
		t.body,
		i.cid,
		i.alt,
		COALESCE(pr.replies, '{}') AS replies
	FROM posts p
	LEFT JOIN text_posts t ON p.id = t.post_id
	LEFT JOIN image_posts i ON p.id = i.post_id
	LEFT JOIN (
		SELECT pr.to_id, array_agg(pr.from_id) AS replies
		FROM post_replies pr
		JOIN posts rp ON rp.id = pr.from_id
		WHERE rp.deleted = FALSE
		GROUP BY pr.to_id
	) pr ON pr.to_id = p.id
	WHERE p.id = $1
  `, id)
	p := types.Post{}
	var body *string
	var cid *string
	var alt *string
	err := row.Scan(&p.ThreadID, &p.Username, &p.Anon, &p.Nick, &p.Color, &p.PostedAt, &p.Deleted, &body, &cid, &alt, &p.Backlinks)
	if err != nil {
		return nil, err
	}
	if body != nil {
		p.TextContent = &types.TextContent{Body: *body}
	}
	if cid != nil {
		p.ImageContent = &types.ImageContent{CID: *cid, Alt: alt}
	}
	p.ID = id
	return &p, nil
}

func (m *MockStore) GetPost(id uint32, ctx context.Context) (*types.Post, error) {
	return nil, nil
}

func (m *MockStore) GetMaxPostId(ctx context.Context) (uint32, error) {
	return 0, nil
}
func (s *Store) GetMaxPostId(ctx context.Context) (uint32, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id FROM posts ORDER BY id DESC
		`)
	var id uint32
	err := row.Scan(&id)
	if err != nil {
		clog.Warn("db: GetMaxPostId: %s", err)
		return 0, err
	}
	return id, nil
}

func (s *Store) GetPostThreadID(postId uint32, ctx context.Context) (uint32, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT thread_id FROM posts WHERE id=$1 AND deleted=FALSE
		`, postId)
	var tid uint32
	err := row.Scan(&tid)
	if err != nil {
		return 0, err
	}
	return tid, nil
}

func (m *MockStore) GetPostThreadID(postId uint32, ctx context.Context) (uint32, error) {
	return 0, nil
}

func (s *Store) DeletePost(id uint32, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE posts SET deleted = TRUE WHERE id = $1
		`, id)
	return err
}

func (m *MockStore) DeletePost(id uint32, ctx context.Context) error {
	return nil
}
