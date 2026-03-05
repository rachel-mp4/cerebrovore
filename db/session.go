package db

import (
	"context"
)

func (s *Store) CreateAuthRequest(state string, pkceVerifier string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO requests (state, pkce_verifier)
		VALUES ($1, $2)
		`, state, pkceVerifier)
	return err
}

func (s *Store) DeleteAuthRequest(state string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM requests WHERE state = $1
		`, state)
	return err
}

func (s *Store) CreateSession(sessionID string, username string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (session_id, username)
		VALUES ($1, $2)
		`, sessionID, username)
	return err
}

func (s *Store) RetrieveSession(sessionID string, ctx context.Context) (username string, err error) {
	row := s.pool.QueryRow(ctx, `
		SELECT username FROM sessions WHERE session_id = $1
		`, sessionID)
	err = row.Scan(&username)
	return
}

func (s *Store) DeleteSession(sessionID string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM sessions WHERE session_id = $1
		`, sessionID)
	return err
}
func (s *Store) DeleteAllSessions(username string, ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM sessions WHERE username = $1
		`, username)
	return err
}

func (m *MockStore) CreateAuthRequest(state string, pkceVerifier string, ctx context.Context) error {
	return nil
}

func (m *MockStore) DeleteAuthRequest(state string, ctx context.Context) error {
	return nil

}

func (m *MockStore) CreateSession(sessionID string, username string, ctx context.Context) error {
	return nil
}

func (m *MockStore) RetrieveSession(sessionID string, ctx context.Context) (username string, err error) {
	return "mock.username", nil
}

func (m *MockStore) DeleteSession(sessionID string, ctx context.Context) error {
	return nil
}

func (m *MockStore) DeleteAllSessions(username string, ctx context.Context) error {
	return nil
}
