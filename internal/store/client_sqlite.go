package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/MKhiriev/go-pass-keeper/models"
)

var ErrLocalSessionNotFound = errors.New("local session not found")

type localSQLiteStorage struct {
	path     string
	inMemory bool

	mu      sync.RWMutex
	nextID  int64
	items   map[string]models.PrivateData
	session *localSession
}

type localSession struct {
	UserID int64     `json:"user_id"`
	Token  string    `json:"token"`
	At     time.Time `json:"at"`
}

type localPersistedState struct {
	NextID  int64                         `json:"next_id"`
	Items   map[string]models.PrivateData `json:"items"`
	Session *localSession                 `json:"session,omitempty"`
}

func NewLocalStorage(dbPath string) (LocalStorage, error) {
	if dbPath == "" {
		dbPath = ":memory:"
	}

	inMemory := dbPath == ":memory:" || dbPath == "memory"
	s := &localSQLiteStorage{
		path:     dbPath,
		inMemory: inMemory,
		items:    make(map[string]models.PrivateData),
		nextID:   1,
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *localSQLiteStorage) load() error {
	if s.inMemory {
		return nil
	}

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read local storage file: %w", err)
	}

	var st localPersistedState
	if err = json.Unmarshal(data, &st); err != nil {
		return fmt.Errorf("decode local storage file: %w", err)
	}

	if st.NextID <= 0 {
		st.NextID = 1
	}
	if st.Items == nil {
		st.Items = make(map[string]models.PrivateData)
	}

	s.nextID = st.NextID
	s.items = st.Items
	s.session = st.Session

	return nil
}

func (s *localSQLiteStorage) persist() error {
	if s.inMemory {
		return nil
	}

	dir := filepath.Dir(s.path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create local storage dir: %w", err)
		}
	}

	state := localPersistedState{NextID: s.nextID, Items: s.items, Session: s.session}
	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode local storage: %w", err)
	}

	if err = os.WriteFile(s.path, payload, 0o600); err != nil {
		return fmt.Errorf("write local storage file: %w", err)
	}

	return nil
}
