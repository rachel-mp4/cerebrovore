package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rachel-mp4/cerebrovore/clog"
	"github.com/rachel-mp4/cerebrovore/types"
)

type Storer interface {
	// auth methods
	CreateAuthRequest(state string, pkceVerifier string, ctx context.Context) error
	DeleteAuthRequest(state string, ctx context.Context) error
	CreateSession(sessionID string, username string, ctx context.Context) error
	RetrieveSession(sessionID string, ctx context.Context) (username string, err error)
	DeleteSession(sessionID string, ctx context.Context) error
	DeleteAllSessions(username string, ctx context.Context) error

	// thread methods

	// CreateThread creates a thread
	CreateThread(thread *types.Thread, ctx context.Context) error

	// GetAllThreads gets the ID, the Topic, and the ReplyCount, for ALL threads
	// for use in model initialization
	GetAllThreads(ctx context.Context) ([]types.Thread, error)

	// GetBumps gets the ID and Topic for the 5 most recently bumped threads.
	// Consider caching this in Storer implementation
	GetBumps(ctx context.Context) ([]types.Thread, error)

	// GetBumpedCatalog gets the thread info and the OP for the limit most recently
	// bumped threads before before (if given). If cursor is non-nil, provide it as
	// the next value of before to get the next limit threads
	GetBumpedCatalog(before *time.Time, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *time.Time, err error)

	// GetRecentCatalog gets the thread info and the OP for the limit most recently
	// posted threads before before (if given). If cursor is non-nil, provide it as
	// the next value of before to get the next limit threads
	GetRecentCatalog(before *uint32, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *uint32, err error)

	// GetRecentThreads gets the ID, the Topic, the ReplyCount, the OP, and the last 5
	// non-OP replies for the most recently posted limit threads before before (if given).
	// If cursor is non-nil, provide it as the next value of before to get the next limit
	// threads.
	GetRecentThreads(before *uint32, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *uint32, err error)
	GetRecentDeadThreads(before *uint32, limit int, archived bool, ctx context.Context) (threads []types.ForumThreadThumb, cursor *uint32, err error)

	// GetBumpedThreads gets the ID, the Topic, the ReplyCount, the OP, and the last 5
	// non-OP replies for the most recently bumped limit threads before before (if given).
	// If cursor is non-nil, provide it as the next value of before to get the next limit
	// threads.
	GetBumpedThreads(before *time.Time, limit int, archived bool, ctx context.Context) (threads []types.Thread, cursor *time.Time, err error)
	GetBumpedDeadThreads(before *time.Time, limit int, archived bool, ctx context.Context) (threads []types.ForumThreadThumb, cursor *time.Time, err error)

	// GetThread gets all the stored information about a thread, and up to limit replies,
	// reverse chronologically, posted before before, if provided. if there are more replies
	// in a thread that aren't provided, the cursor will be non-nil, and
	GetThread(id uint32, viewerIsMod bool, viewerUsername string, ctx context.Context) (threads *types.Thread, err error)
	GetDeadThread(id uint32, viewerIsMod bool, viewerUsername string, ctx context.Context) (threads *types.Thread, err error)

	// DeleteThread sets a thread's deleted column to deleted, to hide it from users
	DeleteThread(id uint32, ctx context.Context) error

	ArchiveThread(id uint32, ctx context.Context) error

	// watch methods

	// StartWatchContext gets all the threads that a username is currently watching
	// & sets their notification status to true so that unneeded persistent notifications
	// aren't created
	StartWatchContext(username string, ctx context.Context) (threadsWatching []uint32, err error)
	EndWatchContext(username string, ctx context.Context) error

	// WatchThread watches a thread for a user, and if they weren't already watching
	// it, returns changed = true
	WatchThread(username string, id uint32, ctx context.Context) (changed bool, err error)

	// UnwatchThread unwatches a thread for a user, and if they were already watching
	// it, returns changed = true
	UnwatchThread(username string, id uint32, ctx context.Context) (changed bool, err error)

	// IsWatched returns if a user is watching a thread
	IsWatched(username string, id uint32, ctx context.Context) bool

	// RemoveWatchesFor removes all watchers for a thread
	RemoveWatchersFor(id uint32, ctx context.Context) error

	// ThreadStatus returns if a thread is at bump limit, or reply limit
	ThreadStatus(id uint32, ctx context.Context) (bumplimit bool, replylimit bool, deleted bool, err error)
	ThreadIsDead(id uint32, ctx context.Context) (bool, error)

	// post methods
	CreatePost(post *types.Post, ctx context.Context) (int, []Backlink, error)
	EZPost(id uint32, ctx context.Context) (*types.Post, error)
	GetPost(id uint32, ctx context.Context) (*types.Post, error)
	GetMaxPostId(ctx context.Context) (uint32, error)
	GetPostThreadID(postId uint32, ctx context.Context) (uint32, error)
	DeletePost(id uint32, ctx context.Context) error

	// ban methods
	Ban(ban *types.Ban, ctx context.Context) error
	SelfBan(username string, postid *uint32, reason string, til time.Time, ctx context.Context) error
	ClearOldSelfBans(ctx context.Context) (int64, error)
	Appeal(banid int, username string, response string, ctx context.Context) error
	GetAppeals(limit int, before *int, ctx context.Context) (actions []types.Action, cursor *int, err error)
	IsBanned(username string, ctx context.Context) (*types.Ban, *types.Post, error)
	// EZBan easily gets if a user is self banned or mod banned
	EZBan(username string, ctx context.Context) (self bool, mod bool, err error)
	Unban(id int, ctx context.Context) error
	Reject(id int, ctx context.Context) error
	DeleteUsersPostsAndThreads(username string, ctx context.Context) error

	// profile methods
	InitializeProfile(username string, ctx context.Context) error
	DeleteProfile(username string, ctx context.Context) error
	UpdateProfile(profile *types.ProfileHead, ctx context.Context) error
	UpdateProfilePicture(profile *types.ProfileHead, ctx context.Context) error
	UpdateProfileContents(username string, profile *types.ProfileExtras, ctx context.Context) error
	GetProfile(username string, ctx context.Context) (*types.ProfileHead, error)
	GetFullProfile(username string, ctx context.Context) (*types.Profile, error)

	// report methods
	Report(report *types.Report, ctx context.Context) (int, error)
	GetReports(limit int, after *int, ctx context.Context) (reports []types.Report, cursor *int, err error)
	GetReport(id int, ctx context.Context) (types.Report, error)
	GetReportersFor(username string, ctx context.Context) (reporters []string, err error)
	ReviewReport(id int, reviewer string, ctx context.Context) error
	ReviewAllReportsBy(reporter string, reviewer string, ctx context.Context) error

	// moderator methods
	MakeModerator(username string, ctx context.Context) error
	RemoveModerator(username string, ctx context.Context) error
	IsModerator(username string, ctx context.Context) (bool, error)
	GetModerators(ctx context.Context) ([]string, error)

	//notification methods
	GetNotifications(username string, limit int, before *int, ctx context.Context) (notifications []types.Notification, cursor *int, includesLastRead bool, err error)
	GetUnreadNotificationCount(username string, ctx context.Context) (int, error)
	ReadNotifications(username string, ctx context.Context) error
	GetAllUsernames(ctx context.Context) ([]string, error)
	CreateReplyNotifications(usernames []string, postid uint32, ctx context.Context) error
	CreateMentionNotifications(usernames []string, postid uint32, ctx context.Context) error
	CreateWatchNotifications(threadid uint32, ctx context.Context) error
	CreatePokeNotification(username string, from string, message *string, ctx context.Context) error
	CreateModNotification(username string, reason string, ctx context.Context) error
	CreateModNotifications(usernames []string, reason string, ctx context.Context) error
	CreateGetNotification(username string, postid uint32, value int, ctx context.Context) error
	CreateReportNotifications(usernames []string, id int, ctx context.Context) error
}

type Backlink struct {
	From       uint32
	To         uint32
	ToUsername string
}

type MockStore struct {
	lastId int
}

func MockInit() (*MockStore, error) {
	return &MockStore{0}, nil
}

type Store struct {
	pool *pgxpool.Pool
}

func Init() (*Store, error) {
	pool, err := initialize()
	return &Store{pool}, err
}

func initialize() (*pgxpool.Pool, error) {
	dbuser := os.Getenv("POSTGRES_USER")
	dbpass := os.Getenv("POSTGRES_PASSWORD")
	dbhost := "localhost"
	dbport := os.Getenv("POSTGRES_PORT")
	dbdb := os.Getenv("POSTGRES_DB")
	dburl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbuser, dbpass, dbhost, dbport, dbdb)
	pool, err := pgxpool.New(context.Background(), dburl)
	if err != nil {
		return nil, err
	}
	pingErr := pool.Ping(context.Background())
	if pingErr != nil {
		return nil, pingErr
	}
	clog.Okay("connected!")
	return pool, nil
}
