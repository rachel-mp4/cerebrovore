package id

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type MemoryProvider struct {
	nameToHash map[string]string
	mu         sync.RWMutex
}

func NewMemoryProvider() *MemoryProvider {
	file, err := os.Open(".fileStore")
	if err != nil {
		return &MemoryProvider{
			nameToHash: make(map[string]string),
		}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	nth := make(map[string]string, 10)
	for scanner.Scan() {
		l := scanner.Text()
		split := strings.Split(l, ",")
		if len(split) != 2 {
			continue
		}
		nth[split[0]] = split[1]
	}
	return &MemoryProvider{
		nameToHash: nth,
	}
}

func (m *MemoryProvider) CreateAccount(username string, password string, _ string, _ context.Context) error {
	err := prevalidate(username)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.nameToHash[username]
	if ok {
		return ErrUserExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInternal, err.Error())
	}
	m.nameToHash[username] = string(hash)
	file, err := os.OpenFile(".fileStore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInternal, err.Error())
	}
	defer file.Close()
	_, err = fmt.Fprintf(file, "%s,%s\n", username, string(hash))
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInternal, err.Error())
	}
	return nil
}

func (m *MemoryProvider) VerifyCredentials(username string, password string, _ context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	hash, ok := m.nameToHash[username]
	if !ok {
		return ErrUserDNE
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("%w: %s", ErrNotAuthorized, err.Error())
	}
	return nil
}

func (m *MemoryProvider) GenerateCode(_ string, _ context.Context) (code string, err error) {
	return "", nil
}

func (m *MemoryProvider) GeneratePublicCode(_ string, _ context.Context) (code string, err error) {
	return "", nil
}
