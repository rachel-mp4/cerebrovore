package id

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rachel-mp4/cerebrovore/clog"
)

type ServiceProvider struct {
	port int
	cli  *http.Client
}

func NewServiceProvider(port int) *ServiceProvider {
	return &ServiceProvider{port, http.DefaultClient}
}

type credentials struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
	Code     *string `json:"code,omitempty"`
}

func (s *ServiceProvider) CreateAccount(username string, password string, invite string, ctx context.Context) error {
	err := prevalidate(username)
	if err != nil {
		return err
	}
	if invite == "" {
		return fmt.Errorf("invite empty: %w", ErrBadData)
	}
	creds := &credentials{
		Username: &username,
		Password: &password,
		Code:     &invite,
	}
	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrInternal)
	}
	body := bytes.NewBuffer(data)
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://localhost:%d/create-account", s.port), body)
	if err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrInternal)
	}
	resp, err := s.cli.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrInternal)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%s: %w", err.Error(), ErrInternal)
		}
		d := string(data)
		switch d {
		case "invite does not exist":
			return fmt.Errorf("%s: %w", d, ErrCodeDNE)
		case "invite already used":
			return fmt.Errorf("%s: %w", d, ErrCodeUsed)
		case "invite has expired":
			return fmt.Errorf("%s: %w", d, ErrCodeExpired)
		case "username already exists":
			return fmt.Errorf("%s: %w", d, ErrUserExists)
		default:
			clog.Warn("internal error: %s", d)
			return ErrInternal
		}
	}
	return nil
}

func (s *ServiceProvider) VerifyCredentials(username string, password string, ctx context.Context) error {
	creds := &credentials{
		Username: &username,
		Password: &password,
	}
	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrInternal)
	}
	body := bytes.NewBuffer(data)
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://localhost:%d/validate-account", s.port), body)
	if err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrInternal)
	}
	resp, err := s.cli.Do(req)
	if err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrInternal)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%s: %w", err.Error(), ErrInternal)
		}
		d := string(data)
		switch d {
		case "account does not exist":
			return fmt.Errorf("%s: %w", d, ErrUserDNE)
		case "account banned":
			return fmt.Errorf("%s: %w", d, ErrUserBanned)
		case "account deleted":
			return fmt.Errorf("%s: %w", d, ErrUserDeleted)
		case "incorrect password":
			return fmt.Errorf("%s: %w", d, ErrNotAuthorized)
		default:
			clog.Warn("internal error: %s", d)
			return ErrInternal
		}
	}
	return nil
}

func (s *ServiceProvider) GenerateCode(username string, ctx context.Context) (code string, err error) {
	return s.generateCode(username, ctx, "generate-code")
}

func (s *ServiceProvider) generateCode(username string, ctx context.Context, endpoint string) (code string, err error) {
	creds := &credentials{
		Username: &username,
	}
	data, err := json.Marshal(creds)
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), ErrInternal)
		return
	}
	body := bytes.NewBuffer(data)
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("http://localhost:%d/%s", s.port, endpoint), body)
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), ErrInternal)
		return
	}
	resp, err := s.cli.Do(req)
	if err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), ErrInternal)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var data []byte
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("%s: %w", err.Error(), ErrInternal)
			return
		}
		d := strings.TrimSpace(string(data))
		clog.Info(d)
		switch d {
		case "account does not exist":
			err = fmt.Errorf("%s: %w", d, ErrUserDNE)
			return
		case "account banned":
			err = fmt.Errorf("%s: %w", d, ErrUserBanned)
			return
		case "account deleted":
			err = fmt.Errorf("%s: %w", d, ErrUserDeleted)
			return
		case "out of mana":
			err = fmt.Errorf("%s: %w", d, ErrNotEnoughMana)
			return
		default:
			clog.Warn("internal error: %s", d)
			err = ErrInternal
			return
		}
	}
	var rescreds credentials
	err = json.NewDecoder(resp.Body).Decode(&rescreds)
	if err != nil {
		err = fmt.Errorf("%s: %w", "failed to parse json", ErrInternal)
		return
	}
	if rescreds.Code != nil {
		code = *rescreds.Code
		return
	}
	err = fmt.Errorf("%s: %w", "no code in resp", ErrInternal)
	return
}

func (s *ServiceProvider) GeneratePublicCode(username string, ctx context.Context) (code string, err error) {
	return s.generateCode(username, ctx, "generate-public-code")
}
