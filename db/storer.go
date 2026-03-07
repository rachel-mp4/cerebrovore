package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rachel-mp4/cerebrovore/types"
	"os"
	"time"
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

	// GetRecentThreads gets the ID, the Topic, the ReplyCount, the OP, and the last 3
	// non-OP replies for the most recently posted limit threads before before (if given).
	// If cursor is non-nil, provide it as the next value of before to get the next limit
	// threads.
	// DOES NOT CURRENTLY GET ANY POSTS
	GetRecentThreads(before *uint32, limit int, ctx context.Context) (threads []types.Thread, cursor *uint32, err error)

	// GetBumpedThreads gets the ID, the Topic, the ReplyCount, the OP, and the last 3
	// non-OP replies for the most recently bumped limit threads before before (if given).
	// If cursor is non-nil, provide it as the next value of before to get the next limit
	// threads.
	// DOES NOT CURRENTLY GET ANY POSTS
	GetBumpedThreads(before *time.Time, limit int, ctx context.Context) (threads []types.Thread, cursor *time.Time, err error)

	// GetThread gets all the stored information about a thread, and up to limit replies,
	// reverse chronologically, posted before before, if provided. if there are more replies
	// in a thread that aren't provided, the cursor will be non-nil, and
	GetThread(id uint32, before *uint32, limit int, ctx context.Context) (threads *types.Thread, cursor *uint32, err error)

	// watch methods
	// GetWatchedThreads gets all the threads that a username is watching
	GetWatchedThreads(username string, ctx context.Context) ([]uint32, error)
	WatchThread(username string, id uint32, ctx context.Context) error
	UnwatchThread(username string, id uint32, ctx context.Context) error
	IsWatched(username string, id uint32, ctx context.Context) bool
	RemoveWatchersFor(id uint32, ctx context.Context) error

	// post methods
	CreatePost(post *types.Post, ctx context.Context) (int, []Backlink, error)
	GetMaxPostId(ctx context.Context) (uint32, error)
	GetPostThreadID(postId uint32, ctx context.Context) (uint32, error)
}

type Backlink struct {
	From uint32
	To   uint32
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
	fmt.Println("connected!")
	return pool, nil
}
