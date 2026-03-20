package id

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Provider interface {
	// CreateAccount attempts to create an account. returns nil if account was successfully created, otherwise
	// returns ErrCodeDNE, ErrCodeUsed, ErrCodeExpired, ErrUserExists, ErrBadData or ErrInternal
	CreateAccount(username string, password string, invite string, ctx context.Context) error

	// VerifyCredentials attempts to verify a user's credentials. returns nil if password is correct, otherwise
	// returns ErrNotAuthorized, ErrUserDNE, ErrUserBanned, ErrUserDeleted or ErrInternal
	VerifyCredentials(username string, password string, ctx context.Context) error

	// GenerateCode attempts to generate an invite code. returns a code and nil if code was generated succesfully, otherwise
	// returns ErrNotEnoughMana, ErrUserDNE, ErrUserBanned, ErrUserDeleted, or ErrInternal
	GenerateCode(username string, ctx context.Context) (code string, err error)

	// GenerateCode attempts to generate a public invite code. returns a code and nil if code was generated succesfully, otherwise
	// returns ErrNotEnoughMana, ErrUserDNE, ErrUserBanned, ErrUserDeleted, or ErrInternal
	GeneratePublicCode(username string, ctx context.Context) (code string, err error)
}

var (
	ErrBadData       = errors.New("bad data")
	ErrUserDNE       = errors.New("account does not exist")
	ErrCodeDNE       = errors.New("invite code does not exist")
	ErrInternal      = errors.New("internal provider error")
	ErrCodeUsed      = errors.New("invite code already used")
	ErrUserBanned    = errors.New("account has been banned")
	ErrUserExists    = errors.New("username already exists")
	ErrUserDeleted   = errors.New("account has been Deleted")
	ErrCodeExpired   = errors.New("invite code expired")
	ErrNotEnoughMana = errors.New("not enough mana to generate code")
	ErrNotAuthorized = errors.New("credentials don't match")
)

var stripper = regexp.MustCompile(`\s`)

func prevalidate(username string) error {
	stripped := stripper.ReplaceAllString(username, "")
	if stripped != username || strings.Contains(username, ",") {
		return fmt.Errorf("username contains illegal characters: %w", ErrBadData)
	}
	return nil
}
