package types

import (
	"time"
)

type ProfileHead struct {
	Username     string
	DisplayName  *string
	Avatar       *string
	IsPixelArt   *bool
	Color        *uint32
	Status       *string
	Bio          *string
	BioIsMono    *bool
	AtIdentifier *string
}

type Profile struct {
	ProfileHead

	// full profile info
	Friends [12]*ProfileFriend
	Posts   [12]*ProfilePost
	Links   [12]*ProfileLink

	FriendsHeader *string
	PostsHeader   *string
	LinksHeader   *string

	// extra info for rendering profile page
	LastSeenLocation *uint32
	LastSeenTopic    *string
	LastSeenTime     *time.Time
}

type ProfileExtras struct {
	FriendsHeader        *string
	FriendInsertUsername []string
	FriendInsertComments []*string
	FriendSelectNames    []*string
	FriendSelectColors   []*uint32
	FriendSelectAvatars  []*string
	FriendSelectBools    []*bool

	PostsHeader         *string
	PostInsertIds       []uint32
	PostInsertComments  []*string
	PostInsertBools     []bool
	PostSelectUsernames []string
	PostSelectAnons     []bool
	PostSelectNicks     []*string
	PostSelectColors    []*uint32
	PostSelectDates     []time.Time
	PostSelectBodys     []*string
	PostSelectCIDS      []*string
	PostSelectAlts      []*string

	LinksHeader        *string
	LinkInsertLinks    []string
	LinkInsertComments []*string
}

type ProfileFriend struct {
	Username string
	Comment  *string
	ProfileHead
}

type ProfilePost struct {
	PostId   uint32
	Comment  *string
	JustBody bool
	Post
}

type ProfileLink struct {
	Link    string
	Comment *string
}
