package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/rachel-mp4/cerebrovore/types"
)

func (s *Store) GetNotifications(username string, limit int, before *int, ctx context.Context) (notifications []types.Notification, cursor *int, includesLastRead bool, err error) {
	q := `
	SELECT
		n.id,
		r.post_id,
		p.username,
		p.anon,
		p.nick,
		p.color,
		p.posted_at,
		p.deleted,
		tp.body,
		ip.cid,
		ip.alt,
		COALESCE(pr.replies, '{}') AS replies,
		m.post_id,
		mp.username,
		mp.anon,
		mp.nick,
		w.thread_id,
		t.topic,
		poke.username,
		poke.message,
		mod.reason,
		get.post_id,
		get.value,
		report.report_id,
		rs.reviewed_by,
		read.username
	FROM notifications n
	LEFT JOIN reply_notifications r ON r.notification_id = n.id
	LEFT JOIN posts p ON p.id = r.post_id
	LEFT JOIN text_posts tp ON p.id = tp.post_id
	LEFT JOIN image_posts ip ON p.id = ip.post_id
	LEFT JOIN (
		SELECT pr.to_id, array_agg(pr.from_id) AS replies
		FROM post_replies pr
		JOIN posts rp ON rp.id = pr.from_id
		WHERE rp.deleted = FALSE
		GROUP BY pr.to_id
	) pr ON pr.to_id = p.id
	LEFT JOIN mention_notifications m ON m.notification_id = n.id
	LEFT JOIN posts mp ON mp.id = m.post_id
	LEFT JOIN watch_notifications w ON w.notification_id = n.id
	LEFT JOIN threads t ON t.id = w.thread_id
	LEFT JOIN poke_notifications poke ON poke.notification_id = n.id
	LEFT JOIN mod_notifications mod ON mod.notification_id = n.id
	LEFT JOIN get_notifications get ON get.notification_id = n.id
	LEFT JOIN report_notifications report ON report.notification_id = n.id
	LEFT JOIN reports rs ON rs.id = report.report_id
	LEFT JOIN read_notifications read ON read.notification_id = n.id
	WHERE n.username = $2 %s
	ORDER BY n.id DESC
	LIMIT $1
	`

	var rows pgx.Rows
	if before != nil {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, "AND n.id < $3"), limit+1, username, *before)
	} else {
		rows, err = s.pool.Query(ctx, fmt.Sprintf(q, ""), limit+1, username)
	}
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
		return
	}
	defer rows.Close()
	i := 0
	for rows.Next() {
		i += 1
		if i == limit+1 {
			break
		}
		n := types.Notification{}
		post := types.Post{}
		var postusername *string
		var postanon *bool
		var postpostedat *time.Time
		var postdeleted *bool
		var body *string
		var cid *string
		var alt *string
		var read *string
		err = rows.Scan(&n.Id, &n.ReplyId, &postusername, &postanon, &post.Nick, &post.Color, &postpostedat, &postdeleted, &body, &cid, &alt, &post.Backlinks, &n.MentionId, &n.Mentioner, &n.MentionAnon, &n.MentionNick, &n.ThreadId, &n.Topic, &n.Poker, &n.PokeMessage, &n.Reason, &n.GetId, &n.Value, &n.Report, &n.Reviewer, &read)
		if err != nil {
			return
		}
		if n.ReplyId != nil {
			post.ID = *n.ReplyId
			if postusername != nil {
				post.Username = *postusername
			} else {
				err = errors.New("null username in non nil reply")
				return
			}
			if postanon != nil {
				post.Anon = *postanon
			} else {
				err = errors.New("null anon in non nil reply")
				return
			}
			if postpostedat != nil {
				post.PostedAt = *postpostedat
			} else {
				err = errors.New("null postedat in non nil reply")
				return
			}
			if postdeleted != nil {
				post.Deleted = *postdeleted
			} else {
				err = errors.New("null deleted in non nil reply")
				return
			}
			if body != nil {
				post.TextContent = &types.TextContent{Body: *body}
			}
			if cid != nil {
				post.ImageContent = &types.ImageContent{CID: *cid, Alt: alt}
			}
			n.Reply = &post
		}
		if read != nil {
			includesLastRead = true
			n.IsLastRead = true
		}
		notifications = append(notifications, n)
		cursor = &n.Id
	}
	if i != limit+1 {
		cursor = nil
	}
	return
}

func (m *MockStore) GetNotifications(username string, limit int, before *int, ctx context.Context) (notifications []types.Notification, cursor *int, includesLastRead bool, err error) {
	return
}

func (s *Store) GetUnreadNotificationCount(username string, ctx context.Context) (int, error) {
	row := s.pool.QueryRow(ctx, `
	SELECT COUNT(*)
	FROM notifications
	WHERE username = $1
	AND id > COALESCE((
		SELECT notification_id
		FROM read_notifications
		WHERE username = $1
	),0)`, username)
	var res int
	err := row.Scan(&res)
	return res, err
}

func (m *MockStore) GetUnreadNotificationCount(username string, ctx context.Context) (int, error) {
	return 0, nil
}

func (s *Store) ReadNotifications(username string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
	INSERT INTO read_notifications (username, notification_id)
	SELECT username, id
	FROM notifications
	WHERE username = $1
	ORDER BY id DESC
	LIMIT 1
	ON CONFLICT (username)
	DO UPDATE SET
		notification_id = EXCLUDED.notification_id
	`, username)
	return err
}

func (m *MockStore) ReadNotifications(username string, ctx context.Context) error {
	return nil
}

func (s *Store) CreateReplyNotifications(usernames []string, postid uint32, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
	WITH inserted_notifications AS (
		INSERT INTO notifications (username)
		SELECT unnest($1::text[])
		RETURNING id
	)
	INSERT INTO reply_notifications (notification_id, post_id)
	SELECT id, $2
	FROM inserted_notifications
	`, usernames, postid)
	return err
}

func (m *MockStore) CreateReplyNotifications(username []string, postid uint32, ctx context.Context) error {
	return nil
}

func (s *Store) CreateMentionNotifications(usernames []string, postid uint32, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
	WITH inserted_notifications AS (
		INSERT INTO notifications (username)
		SELECT unnest($1::text[])
		RETURNING id
	)
	INSERT INTO mention_notifications (notification_id, post_id)
	SELECT id, $2
	FROM inserted_notifications
	`, usernames, postid)
	return err
}

func (m *MockStore) CreateMentionNotifications(username []string, postid uint32, ctx context.Context) error {
	return nil
}

// maybe an index is appropriate & justified on wt.notified, but my reasoning is that
// number of people who watch a given thread is going to be quite low and it's not
// that big a deal and i don't know enough about sql optimization here to stand behind
// this decision without seeing what things are like in the real world for a bit.
// TBH this initially returned the list of users notified, but they're supposed to be
// offline, so that seemed unnecessary. can turn the insert clause into another with,
// and return from users_notified if situation changes
func (s *Store) CreateWatchNotifications(threadid uint32, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
	WITH users_notified AS (
		UPDATE watched_threads
		SET notified = TRUE
		WHERE thread_id = $1 AND notified = FALSE
		RETURNING username
	), 
	inserted_notifications AS (
		INSERT INTO notifications (username)
		SELECT username
		FROM users_notified
		RETURNING id
	)
	INSERT INTO watch_notifications (notification_id, thread_id)
	SELECT id, $1
	FROM inserted_notifications
	`, threadid)
	return err
}

func (m *MockStore) CreateWatchNotifications(threadid uint32, ctx context.Context) error {
	return nil
}

func (s *Store) CreatePokeNotification(username string, from string, message *string, ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(ctx, `
	INSERT INTO notifications (
		username
	) VALUES ($1)
	RETURNING id`, username)
	var id int
	err = row.Scan(&id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
	INSERT INTO poke_notifications (
		notification_id, username, message
	) VALUES ($1, $2, $3)`, id, from, message)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (m *MockStore) CreatePokeNotification(username string, from string, message *string, ctx context.Context) error {
	return nil
}

func (s *Store) CreateModNotification(username string, reason string, ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(ctx, `
	INSERT INTO notifications (
		username
	) VALUES ($1)
	RETURNING id`, username)
	var id int
	err = row.Scan(&id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
	INSERT INTO mod_notifications (
		notification_id, reason
	) VALUES ($1, $2)`, id, reason)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (m *MockStore) CreateModNotification(username string, reason string, ctx context.Context) error {
	return nil
}

func (s *Store) CreateModNotifications(usernames []string, reason string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
	WITH inserted_notifications AS (
		INSERT INTO notifications (username)
		SELECT unnest($1::text[])
		RETURNING id
	)
	INSERT INTO mod_notifications (notification_id, reason)
	SELECT id, $2
	FROM inserted_notifications
	`, usernames, reason)
	return err
}

func (m *MockStore) CreateModNotifications(usernames []string, reason string, ctx context.Context) error {
	return nil
}

func (s *Store) CreateGetNotification(username string, postid uint32, value int, ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	row := tx.QueryRow(ctx, `
	INSERT INTO notifications (
		username
	) VALUES ($1)
	RETURNING id`, username)
	var id int
	err = row.Scan(&id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
	INSERT INTO get_notifications (
		notification_id, post_id, value
	) VALUES ($1, $2, $3)`, id, postid, value)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (m *MockStore) CreateGetNotification(username string, postid uint32, value int, ctx context.Context) error {
	return nil
}

func (s *Store) GetAllUsernames(ctx context.Context) (usernames []string, err error) {
	rows, err := s.pool.Query(ctx, `SELECT username FROM profiles`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var username string
		err = rows.Scan(&username)
		if err != nil {
			return
		}
		usernames = append(usernames, username)
	}
	return
}

func (m *MockStore) GetAllUsernames(ctx context.Context) (usernames []string, err error) {
	return
}

func (s *Store) CreateReportNotifications(usernames []string, id int, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
	WITH inserted_notifications AS (
		INSERT INTO notifications (username)
		SELECT unnest($1::text[])
		RETURNING id
	)
	INSERT INTO report_notifications (notification_id, report_id)
	SELECT id, $2
	FROM inserted_notifications
	`, usernames, id)
	return err
}

func (m *MockStore) CreateReportNotifications(usernames []string, id int, ctx context.Context) error {
	return nil
}
