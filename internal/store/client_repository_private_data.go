package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/MKhiriev/go-pass-keeper/models"
)

func (s *localSQLiteStorage) Save(ctx context.Context, data ...*models.PrivateData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range data {
		if err := ctx.Err(); err != nil {
			return err
		}
		if item == nil {
			continue
		}
		if item.ID == 0 {
			item.ID = s.nextID
			s.nextID++
		}
		if item.CreatedAt == nil {
			n := time.Now().UTC()
			item.CreatedAt = &n
		}
		if item.UpdatedAt == nil {
			n := time.Now().UTC()
			item.UpdatedAt = &n
		}
		s.items[item.ClientSideID] = *item
	}

	if err := s.persist(); err != nil {
		return fmt.Errorf("persist saved local data: %w", err)
	}
	return nil
}

func (s *localSQLiteStorage) GetAll(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]models.PrivateData, 0, len(s.items))
	for _, item := range s.items {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if item.UserID == userID && !item.Deleted {
			out = append(out, item)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt == nil || out[j].UpdatedAt == nil {
			return out[i].ID > out[j].ID
		}
		return out[i].UpdatedAt.After(*out[j].UpdatedAt)
	})

	return out, nil
}

func (s *localSQLiteStorage) Get(ctx context.Context, clientSideID string) (models.PrivateData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return models.PrivateData{}, err
	}

	item, ok := s.items[clientSideID]
	if !ok {
		return models.PrivateData{}, ErrPrivateDataNotFound
	}
	return item, nil
}

func (s *localSQLiteStorage) GetAllStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	states := make([]models.PrivateDataState, 0, len(s.items))
	for _, item := range s.items {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if item.UserID != userID {
			continue
		}
		states = append(states, models.PrivateDataState{
			ClientSideID: item.ClientSideID,
			Hash:         item.Hash,
			Version:      item.Version,
			Deleted:      item.Deleted,
			UpdatedAt:    item.UpdatedAt,
		})
	}

	return states, nil
}

func (s *localSQLiteStorage) Update(ctx context.Context, data models.PrivateData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	current, ok := s.items[data.ClientSideID]
	if !ok || current.UserID != data.UserID {
		return ErrPrivateDataNotFound
	}

	if data.CreatedAt == nil {
		data.CreatedAt = current.CreatedAt
	}
	if data.UpdatedAt == nil {
		n := time.Now().UTC()
		data.UpdatedAt = &n
	}
	s.items[data.ClientSideID] = data

	if err := s.persist(); err != nil {
		return fmt.Errorf("persist local update: %w", err)
	}
	return nil
}

func (s *localSQLiteStorage) SoftDelete(ctx context.Context, clientSideID string, version int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	item, ok := s.items[clientSideID]
	if !ok {
		return ErrPrivateDataNotFound
	}
	item.Deleted = true
	item.Version = version
	n := time.Now().UTC()
	item.UpdatedAt = &n
	s.items[clientSideID] = item

	if err := s.persist(); err != nil {
		return fmt.Errorf("persist local soft delete: %w", err)
	}
	return nil
}

func (s *localSQLiteStorage) SaveSession(ctx context.Context, userID int64, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	s.session = &localSession{UserID: userID, Token: token, At: time.Now().UTC()}
	if err := s.persist(); err != nil {
		return fmt.Errorf("persist local session: %w", err)
	}
	return nil
}

func (s *localSQLiteStorage) LoadSession(ctx context.Context) (int64, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return 0, "", err
	}

	if s.session == nil || s.session.Token == "" {
		return 0, "", ErrLocalSessionNotFound
	}

	return s.session.UserID, s.session.Token, nil
}

func (s *localSQLiteStorage) ClearSession(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	s.session = nil
	if err := s.persist(); err != nil {
		return fmt.Errorf("persist clear local session: %w", err)
	}
	return nil
}
