package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
)

func (s *Store) Ban(ban *types.Ban, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
	INSERT INTO bans (
		username, 
		post_id,
		reason, 
		comment,
		until, 
		moderator
	)
	VALUES ($1, $2, $3, $4, $5, $6)
	`,
		ban.Username,
		ban.PostId,
		ban.Reason,
		ban.Comment,
		ban.Until,
		ban.Moderator,
	)
	return err
}

func (m *MockStore) Ban(ban *types.Ban, ctx context.Context) error {
	return nil
}

func (s *Store) SelfBan(username string, postid *uint32, reason string, til time.Time, ctx context.Context) error {
	comment := "selfBan"
	return s.Ban(&types.Ban{
		Username:  username,
		PostId:    postid,
		Reason:    &reason,
		Comment:   &comment,
		Until:     til,
		Moderator: username,
	}, ctx)
}

func (m *MockStore) SelfBan(username string, postid *uint32, reason string, til time.Time, ctx context.Context) error {
	return nil
}

func (s *Store) Unban(id int, ctx context.Context) error {
	tag, err := s.pool.Exec(ctx, "UPDATE bans SET repealed = TRUE WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		clog.Warn("unban affected %d rows", tag.RowsAffected())
		return errors.New("didn't affect 1 row!")
	}
	return nil
}

func (m *MockStore) Unban(id int, ctx context.Context) error {
	return nil
}

func (s *Store) Reject(id int, ctx context.Context) error {
	tag, err := s.pool.Exec(ctx, "UPDATE bans SET repealed = FALSE WHERE id = $1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		clog.Warn("unban affected %d rows", tag.RowsAffected())
		return errors.New("didn't affect 1 row!")
	}
	return nil
}

func (m *MockStore) Reject(id int, ctx context.Context) error {
	return nil
}

func (s *Store) Appeal(banid int, username string, response string, ctx context.Context) error {
	tag, err := s.pool.Exec(ctx, "UPDATE bans SET response = $1 WHERE id = $2 AND username = $3", response, banid, username)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		clog.Warn("appeal affected %d rows", tag.RowsAffected())
		return errors.New("didn't affect 1 row!")
	}
	return nil
}

func (m *MockStore) Appeal(id int, username string, response string, ctx context.Context) error {
	return nil
}

func (s *Store) GetAppeals(limit int, before *int, ctx context.Context) (bans []types.Appeal, cursor *int, err error) {
	q := `
	SELECT 
		b.id, 
		b.username,
		b.post_id,
		b.reason,
		b.comment,
		b.response,
		b.banned_at,
		b.until,
		b.moderator,
		p.nick,
		p.color,
		p.posted_at,
		t.body,
		i.cid,
		i.alt
	FROM bans b
	LEFT JOIN posts p ON b.post_id = p.id
	LEFT JOIN text_posts t ON b.post_id = t.post_id
	LEFT JOIN image_posts i ON b.post_id = i.post_id
	WHERE b.response IS NOT NULL
	AND b.repealed IS NULL
	AND b.until > now()
	%s
	ORDER BY b.id DESC
	LIMIT $1
	`
	var rows pgx.Rows
	if before != nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "AND id < $2"), limit+1, *before)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, ""), limit+1)
	}
	if err != nil {
		return
	}
	defer rows.Close()
	bans = make([]types.Appeal, 0)
	i := 0
	for rows.Next() {
		i = i + 1
		if i == limit+1 {
			break
		}
		b := types.Ban{}
		var color *uint32
		var nick *string
		var postedAt *time.Time
		var body *string
		var cid *string
		var alt *string
		err = rows.Scan(&b.Id, &b.Username, &b.PostId, &b.Reason, &b.Comment, &b.Response, &b.BannedAt, &b.Until, &b.Moderator, &nick, &color, &postedAt, &body, &cid, &alt)
		if err != nil {
			return
		}
		cursor = &b.Id
		a := types.Appeal{Ban: b}
		if b.PostId != nil {
			p := types.Post{
				ID:       *b.PostId,
				Username: b.Username,
				Nick:     nick,
				Color:    color,
				PostedAt: *postedAt,
			}
			if body != nil {
				p.TextContent = &types.TextContent{Body: *body}
			}
			if cid != nil {
				p.ImageContent = &types.ImageContent{CID: *cid, Alt: alt}
			}
			a.Post = p
		}
		bans = append(bans, a)
	}
	if i != limit+1 {
		cursor = nil
	}
	return
}

func (m *MockStore) GetAppeals(limit int, before *int, ctx context.Context) (appeals []types.Appeal, cursor *int, err error) {
	return
}

func (s *Store) IsBanned(username string, ctx context.Context) (*types.Ban, *types.Post, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT 
			id, 
			reason, 
			post_id, 
			response, 
			until 
		FROM bans 
		WHERE username = $1 
		AND repealed IS NOT TRUE
		AND now() < until`, username)
	b := types.Ban{}
	err := row.Scan(&b.Id, &b.Reason, &b.PostId, &b.Response, &b.Until)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = nil
		}
		return nil, nil, err
	}
	if b.PostId != nil {
		row = s.pool.QueryRow(ctx, `
			SELECT 
				p.id,
				p.username,
				p.anon,
				p.nick,
				p.color,
				p.posted_at,
				t.body,
				i.cid,
				i.alt
			FROM posts p
			LEFT JOIN text_posts t ON t.post_id = p.id
			LEFT JOIN image_posts i ON i.post_id = p.id
			WHERE p.id = $1
  	`, *b.PostId)
		p := types.Post{}
		var body *string
		var cid *string
		var alt *string
		err = row.Scan(&p.ID, &p.Username, &p.Anon, &p.Nick, &p.Color, &p.PostedAt, &body, &cid, &alt)
		if err != nil {
			clog.Warn("%s", err)
			if err != pgx.ErrNoRows {
				return nil, nil, err
			}
			return &b, nil, nil
		}
		if body != nil {
			tc := types.TextContent{Body: *body}
			p.TextContent = &tc
		}
		if cid != nil {
			ic := types.ImageContent{CID: *cid, Alt: alt}
			p.ImageContent = &ic
		}
		return &b, &p, nil
	}
	return &b, nil, nil
}

func (m *MockStore) IsBanned(username string, ctx context.Context) (*types.Ban, *types.Post, error) {
	return nil, nil, nil
}

func (s *Store) ClearOldSelfBans(ctx context.Context) (int64, error) {
	tag, err := s.pool.Exec(ctx, `DELETE FROM bans WHERE username = moderator AND now() > until`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (m *MockStore) ClearOldSelfBans(ctx context.Context) (int64, error) {
	return 0, nil
}

func (s *Store) DeleteUsersPostsAndThreads(username string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE posts
		SET deleted = TRUE
		WHERE username = $1
		`, username)
	_, err = s.pool.Exec(ctx, `
		UPDATE threads t
		SET t.deleted = TRUE
		FROM posts p
		WHERE p.id = t.id
		AND username = $1
		`, username)
	return err
}

func (m *MockStore) DeleteUsersPostsAndThreads(username string, ctx context.Context) error {
	return nil
}
