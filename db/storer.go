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
	CreateThread(thread *types.Thread, ctx context.Context) error
	GetAllThreads(ctx context.Context) ([]types.Thread, error)
	GetRecentThreads(before *uint32, limit int, ctx context.Context) ([]types.Thread, *uint32, error)
	GetBumpedThreads(before *time.Time, limit int, ctx context.Context) ([]types.Thread, *time.Time, error)
	GetThread(id uint32, before *uint32, limit int, ctx context.Context) (*types.Thread, *uint32, error)
	GetWatchedThreads(username string, ctx context.Context) ([]uint32, error)
	WatchThread(username string, id uint32, ctx context.Context) error
	UnwatchThread(username string, id uint32, ctx context.Context) error

	// post methods
	CreatePost(post *types.Post, ctx context.Context) ([]Backlink, error)
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
