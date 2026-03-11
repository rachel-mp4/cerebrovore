package db

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type MemoryIDStore struct {
	nameToHash map[string]string
	mu         sync.RWMutex
}

func NewMemoryIDStore() *MemoryIDStore {
	file, err := os.Open(".fileStore")
	if err != nil {
		return &MemoryIDStore{
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
	return &MemoryIDStore{
		nameToHash: nth,
	}
}

var stripper = regexp.MustCompile(`\s`)

func (m *MemoryIDStore) CreateAccount(username string, password string, _ string) error {
	stripped := stripper.ReplaceAllString(username, "")
	if stripped != username || strings.Contains(username, ",") {
		return errors.New("username contains illegal character")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.nameToHash[username]
	if ok {
		return errors.New("Username already exists!")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	m.nameToHash[username] = string(hash)
	file, err := os.OpenFile(".fileStore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Fprintf(file, "%s,%s\n", username, string(hash))
	return nil
}

func (m *MemoryIDStore) VerifyCredentials(username string, password string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	hash, ok := m.nameToHash[username]
	if !ok {
		return errors.New("Account does not exist!")
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
