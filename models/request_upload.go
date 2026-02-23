package models

// UploadRequest represents a batch upload request for storing vault items.
// Used to insert multiple encrypted records in a single operation.
type UploadRequest struct {
	// UserID filters records by owner.
	UserID int64 `json:"user_id"`

	// PrivateDataList contains one or more vault items to be stored.
	PrivateDataList []*PrivateData `json:"private_data_list"`

	// Hash of serialized PrivateDataList — transport integrity check.
	Hash string `json:"hash"`

	// Length is the total number of entries in PrivateDataList.
	Length int `json:"length"`
}
