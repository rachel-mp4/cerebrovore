package db

import (
	"context"
	"log"
	"slices"

	"github.com/rachel-mp4/cerebrovore/types"
	"github.com/rachel-mp4/cerebrovore/utils"
)

func (s *Store) InitializeProfile(username string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO profiles (
			username,
			color,
			status
		)
		VALUES ($1, $2, $3)`, username, utils.PURPLE, "setting up my profile!")
	return err
}

func (m *MockStore) InitializeProfile(username string, ctx context.Context) error {
	return nil
}

func (s *Store) DeleteProfile(username string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM profiles WHERE username = $1`, username)
	return err
}

func (m *MockStore) DeleteProfile(username string, ctx context.Context) error {
	return nil
}

func (s *Store) UpdateProfile(profile *types.ProfileHead, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `UPDATE profiles 
		SET display_name = $2, 
		color = $3, 
		status = $4, 
		bio = $5,
		is_mono = $6,
		at_identifier = $7
		WHERE username = $1`,
		profile.Username,
		profile.DisplayName,
		profile.Color,
		profile.Status,
		profile.Bio,
		profile.BioIsMono,
		profile.AtIdentifier,
	)
	return err
}

func (m *MockStore) UpdateProfile(profile *types.ProfileHead, ctx context.Context) error {
	return nil
}
func (s *Store) UpdateProfilePicture(profile *types.ProfileHead, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE profiles
		SET avatar = $2, is_pixel = $3
		WHERE username = $1
		`, profile.Username, profile.Avatar, profile.IsPixelArt)
	return err
}

func (m *MockStore) UpdateProfilePicture(profile *types.ProfileHead, ctx context.Context) error {
	return nil
}

func (s *Store) UpdateProfileContents(username string, profile *types.ProfileExtras, ctx context.Context) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Commit(ctx)
	_, err = tx.Exec(ctx, `UPDATE profiles SET friends_header = $1, posts_header = $2, links_header = $3 WHERE username = $4`, profile.FriendsHeader, profile.PostsHeader, profile.LinksHeader, username)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `DELETE FROM profile_friends WHERE profile_username = $1`, username)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM profile_posts WHERE profile_username = $1`, username)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM profile_links WHERE profile_username = $1`, username)
	if err != nil {
		return err
	}

	mynameabunch := slices.Repeat([]string{username}, len(profile.FriendInsertUsername))
	ids := make([]int, len(profile.FriendInsertComments))
	for i := range ids {
		ids[i] = i
	}

	_, err = tx.Exec(ctx, `INSERT INTO profile_friends
		SELECT profile_username, friend, comment, id
		FROM unnest($1::text[], $2::text[], $3::text[], $4::int8[]) q(profile_username, friend, comment, id)
		`, mynameabunch, profile.FriendInsertUsername, profile.FriendInsertComments, ids)

	if err != nil {
		return err
	}

	mynameabunch = slices.Repeat([]string{username}, len(profile.PostInsertIds))
	ids = make([]int, len(profile.PostInsertComments))
	for i := range ids {
		ids[i] = i
	}
	_, err = tx.Exec(ctx, `INSERT INTO profile_posts
		SELECT profile_username, post_id, comment, just_body, id
		FROM unnest($1::text[], $2::int8[], $3::text[], $4::boolean[], $5::int8[]) q(profile_username, post_id, comment, just_body, id)
		`, mynameabunch, profile.PostInsertIds, profile.PostInsertComments, profile.PostInsertBools, ids)
	if err != nil {
		return err
	}

	mynameabunch = slices.Repeat([]string{username}, len(profile.LinkInsertLinks))
	ids = make([]int, len(profile.LinkInsertComments))
	for i := range ids {
		ids[i] = i
	}
	_, err = tx.Exec(ctx, `INSERT INTO profile_links
		SELECT profile_username, link, comment, id
		FROM unnest($1::text[], $2::text[], $3::text[], $4::int8[]) q(profile_username, link, comment, id)
		`, mynameabunch, profile.LinkInsertLinks, profile.LinkInsertComments, ids)
	if err != nil {
		return err
	}
	return nil
}

func (m *MockStore) UpdateProfileContents(username string, profile *types.ProfileExtras, ctx context.Context) error {
	return nil
}

func (s *Store) GetProfile(username string, ctx context.Context) (*types.ProfileHead, error) {
	row := s.pool.QueryRow(ctx, `SELECT 
		display_name,
		avatar,
		is_pixel,
		color,
		status,
		bio,
		deleted
		FROM profiles WHERE username = $1`, username)
	var p types.ProfileHead
	err := row.Scan(&p.DisplayName, &p.Avatar, &p.IsPixelArt, &p.Color, &p.Status, &p.Bio, &p.Deleted)
	if err != nil {
		return nil, err
	}
	p.Username = username
	p.SelfBanned, p.Banned, err = s.EZBan(username, ctx)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *MockStore) GetProfile(username string, ctx context.Context) (*types.ProfileHead, error) {
	return nil, nil
}

func (s *Store) GetFullProfile(username string, ctx context.Context) (*types.Profile, error) {
	row := s.pool.QueryRow(ctx, `SELECT
		p.display_name,
		p.avatar,
		p.is_pixel,
		p.color,
		p.status,
		p.bio,
		p.is_mono,
		p.at_identifier,
		p.deleted,
		p.friends_header,
		p.posts_header,
		p.links_header,
		COALESCE(pf.friends, '{}'),
		COALESCE(pf.comments, '{}'),
		COALESCE(pf.names, '{}'),
		COALESCE(pf.colors, '{}'),
		COALESCE(pf.avatars, '{}'),
		COALESCE(pf.pixels, '{}'),
		COALESCE(pp.ids, '{}'),
		COALESCE(pp.comments, '{}'),
		COALESCE(pp.bools, '{}'),
		COALESCE(pp.usernames, '{}'),
		COALESCE(pp.anons, '{}'),
		COALESCE(pp.nicks, '{}'),
		COALESCE(pp.colors, '{}'),
		COALESCE(pp.dates, '{}'),
		COALESCE(pp.bodys, '{}'),
		COALESCE(pp.cids, '{}'),
		COALESCE(pp.alts, '{}'),
		COALESCE(pl.links, '{}'),
		COALESCE(pl.comments, '{}')
		FROM profiles p
		LEFT JOIN LATERAL (
			SELECT 
				array_agg(pf.friend ORDER BY pf.id) AS friends, 
				array_agg(pf.comment ORDER BY pf.id) AS comments,
				array_agg(pro.display_name ORDER BY pf.id) AS names,
				array_agg(pro.color ORDER BY pf.id) AS colors,
				array_agg(pro.avatar ORDER BY pf.id) AS avatars,
				array_agg(pro.is_pixel ORDER BY pf.id) AS pixels
			FROM (
				SELECT *
				FROM profile_friends
				WHERE profile_username = p.username
				ORDER BY id
				LIMIT 12
			) pf
			LEFT JOIN profiles pro ON pro.username = pf.friend
		) pf ON true
		LEFT JOIN LATERAL (
			SELECT 
				array_agg(pp.post_id ORDER BY pp.id) AS ids, 
				array_agg(pp.comment ORDER BY pp.id) AS comments, 
				array_agg(pp.just_body ORDER by pp.id) AS bools,
				array_agg(ps.username ORDER BY pp.id) AS usernames,
				array_agg(ps.nick ORDER BY pp.id) AS nicks,
				array_agg(ps.color ORDER BY pp.id) AS colors,
				array_agg(ps.anon ORDER BY pp.id) AS anons,
				array_agg(ps.posted_at ORDER BY pp.id) AS dates,
				array_agg(t.body ORDER BY pp.id) AS bodys,
				array_agg(i.cid ORDER BY pp.id) AS cids,
				array_agg(i.alt ORDER BY pp.id) AS alts
			FROM (
				SELECT *
				FROM profile_posts
				WHERE profile_username = p.username
				ORDER BY id
				LIMIT 12
			) pp
			LEFT JOIN posts ps ON ps.id = pp.post_id
			LEFT JOIN text_posts t ON t.post_id = ps.id
			LEFT JOIN image_posts i ON i.post_id = ps.id
		) pp ON true
		LEFT JOIN LATERAL (
			SELECT 
				array_agg(pl.link ORDER BY pl.id) AS links, 
				array_agg(pl.comment ORDER BY pl.id) AS comments
			FROM (
				SELECT *
				FROM profile_links
				WHERE profile_username = p.username
				ORDER BY id
				LIMIT 12
			) pl
		) pl ON true
		WHERE p.username = $1
		`, username)
	var p types.Profile
	var pe types.ProfileExtras
	err := row.Scan(
		&p.ProfileHead.DisplayName,
		&p.ProfileHead.Avatar,
		&p.ProfileHead.IsPixelArt,
		&p.ProfileHead.Color,
		&p.ProfileHead.Status,
		&p.ProfileHead.Bio,
		&p.ProfileHead.BioIsMono,
		&p.ProfileHead.AtIdentifier,
		&p.ProfileHead.Deleted,
		&p.FriendsHeader,
		&p.PostsHeader,
		&p.LinksHeader,
		&pe.FriendInsertUsername,
		&pe.FriendInsertComments,
		&pe.FriendSelectNames,
		&pe.FriendSelectColors,
		&pe.FriendSelectAvatars,
		&pe.FriendSelectBools,
		&pe.PostInsertIds,
		&pe.PostInsertComments,
		&pe.PostInsertBools,
		&pe.PostSelectUsernames,
		&pe.PostSelectAnons,
		&pe.PostSelectNicks,
		&pe.PostSelectColors,
		&pe.PostSelectDates,
		&pe.PostSelectBodys,
		&pe.PostSelectCIDS,
		&pe.PostSelectAlts,
		&pe.LinkInsertLinks,
		&pe.LinkInsertComments)
	if err != nil {
		return nil, err
	}

	for idx := range p.Friends {
		if idx >= len(pe.FriendInsertComments) {
			break
		}
		profile := types.ProfileHead{
			Username:    pe.FriendInsertUsername[idx],
			DisplayName: pe.FriendSelectNames[idx],
			Color:       pe.FriendSelectColors[idx],
			Avatar:      pe.FriendSelectAvatars[idx],
			IsPixelArt:  pe.FriendSelectBools[idx],
		}

		p.Friends[idx] = &types.ProfileFriend{
			Username:    pe.FriendInsertUsername[idx],
			Comment:     pe.FriendInsertComments[idx],
			ProfileHead: profile,
		}
	}

	for idx := range p.Posts {
		if idx >= len(pe.PostInsertComments) {
			break
		}
		post := types.Post{
			ID:       pe.PostInsertIds[idx],
			Username: pe.PostSelectUsernames[idx],
			Anon:     pe.PostSelectAnons[idx],
			Nick:     pe.PostSelectNicks[idx],
			Color:    pe.PostSelectColors[idx],
			PostedAt: pe.PostSelectDates[idx],
		}
		if pe.PostSelectBodys[idx] != nil {
			post.TextContent = &types.TextContent{Body: *pe.PostSelectBodys[idx]}
		}
		if pe.PostSelectCIDS[idx] != nil {
			post.ImageContent = &types.ImageContent{CID: *pe.PostSelectCIDS[idx], Alt: pe.PostSelectAlts[idx]}
		}
		p.Posts[idx] = &types.ProfilePost{
			PostId:   pe.PostInsertIds[idx],
			Comment:  pe.PostInsertComments[idx],
			JustBody: pe.PostInsertBools[idx],
			Post:     post,
		}
	}

	for idx := range p.Links {
		if idx >= len(pe.LinkInsertComments) {
			break
		}
		p.Links[idx] = &types.ProfileLink{
			Link:    pe.LinkInsertLinks[idx],
			Comment: pe.LinkInsertComments[idx],
		}
	}

	p.Username = username
	row = s.pool.QueryRow(ctx, `SELECT 
			p.thread_id, 
			t.topic, 
			p.posted_at 
		FROM posts p 
		LEFT JOIN threads t ON p.thread_id = t.id 
		WHERE p.username = $1 
		AND p.anon = FALSE 
		AND t.deleted = FALSE 
		AND p.deleted = FALSE 
		ORDER BY p.id DESC`,
		username)
	err = row.Scan(&p.LastSeenLocation, &p.LastSeenTopic, &p.LastSeenTime)
	if err != nil {
		log.Println(err)
	}
	if p.LastSeenTopic != nil && *p.LastSeenTopic == "" {
		p.LastSeenTopic = nil
	}
	p.SelfBanned, p.Banned, err = s.EZBan(username, ctx)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *MockStore) GetFullProfile(username string, ctx context.Context) (*types.Profile, error) {
	return nil, nil
}
