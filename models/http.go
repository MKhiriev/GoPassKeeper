package models

type DownloadRequest struct {
	ID     int64    `json:"id"`
	UserID int64    `json:"user_id"`
	Type   DataType `json:"type"`
}

type UpdateRequest struct {
	ID           int64                 `json:"id"`
	UserID       int64                 `json:"user_id"`
	Type         DataType              `json:"type"`
	Data         *CipheredData         `json:"data"`
	Metadata     *CipheredMetadata     `json:"metadata"`
	Notes        *CipheredNotes        `json:"notes"`
	CustomFields *CipheredCustomFields `json:"custom_fields"`
}

type DeleteRequest struct {
	ID     int64    `json:"id"`
	UserID int64    `json:"user_id"`
	Type   DataType `json:"type"`
}
