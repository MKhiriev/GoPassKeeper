package models

import "time"

type PrivateDataState struct {
	ClientSideID string     `json:"client_side_id"`
	Hash         string     `json:"hash"`
	Version      int64      `json:"version"`
	Deleted      bool       `json:"deleted"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}
