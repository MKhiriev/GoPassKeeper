// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package validators

import (
	"context"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func ptrMeta(s string) *models.CipheredMetadata       { v := models.CipheredMetadata(s); return &v }
func ptrData(s string) *models.CipheredData           { v := models.CipheredData(s); return &v }
func ptrNotes(s string) *models.CipheredNotes         { v := models.CipheredNotes(s); return &v }
func ptrFields(s string) *models.CipheredCustomFields { v := models.CipheredCustomFields(s); return &v }

func validPrivateData() models.PrivateData {
	return models.PrivateData{
		ClientSideID: "cid-1",
		UserID:       1,
		Payload: models.PrivateDataPayload{
			Metadata: models.CipheredMetadata("meta"),
			Type:     models.LoginPassword,
			Data:     models.CipheredData("data"),
		},
		Hash:    "hash",
		Version: 0,
	}
}

func validPrivateDataUpdate() models.PrivateDataUpdate {
	return models.PrivateDataUpdate{
		ClientSideID:      "cid-1",
		Version:           1,
		UpdatedRecordHash: "new-hash",
		FieldsUpdate: models.FieldsUpdate{
			Metadata: ptrMeta("meta"),
		},
	}
}

// ---------------------------------------------------------------------------
// TestNewPrivateDataValidator
// ---------------------------------------------------------------------------

func TestNewPrivateDataValidator(t *testing.T) {
	v := NewPrivateDataValidator()
	require.NotNil(t, v)
}

// ---------------------------------------------------------------------------
// TestValidate_Dispatch
// ---------------------------------------------------------------------------

func TestValidate_Dispatch(t *testing.T) {
	v := NewPrivateDataValidator()
	ctx := context.Background()

	t.Run("unsupported type", func(t *testing.T) {
		err := v.Validate(ctx, "a string")
		require.ErrorIs(t, err, ErrUnsupportedType)
	})

	t.Run("PrivateData value", func(t *testing.T) {
		d := validPrivateData()
		err := v.Validate(ctx, d)
		require.NoError(t, err)
	})

	t.Run("PrivateData pointer", func(t *testing.T) {
		d := validPrivateData()
		err := v.Validate(ctx, &d)
		require.NoError(t, err)
	})

	t.Run("DownloadRequest value", func(t *testing.T) {
		r := models.DownloadRequest{UserID: 1}
		err := v.Validate(ctx, r)
		require.NoError(t, err)
	})

	t.Run("DownloadRequest pointer", func(t *testing.T) {
		r := models.DownloadRequest{UserID: 1}
		err := v.Validate(ctx, &r)
		require.NoError(t, err)
	})
}

// ---------------------------------------------------------------------------
// TestValidatePrivateData
// ---------------------------------------------------------------------------

func TestValidatePrivateData(t *testing.T) {
	v := NewPrivateDataValidator()
	ctx := context.Background()

	t.Run("valid with defaults", func(t *testing.T) {
		d := validPrivateData()
		require.NoError(t, v.Validate(ctx, d))
	})

	t.Run("empty client_side_id", func(t *testing.T) {
		d := validPrivateData()
		d.ClientSideID = ""
		require.ErrorIs(t, v.Validate(ctx, d, FieldClientSideID), ErrInvalidClientSideID)
	})

	t.Run("zero user_id", func(t *testing.T) {
		d := validPrivateData()
		d.UserID = 0
		require.ErrorIs(t, v.Validate(ctx, d, FieldUserID), ErrInvalidUserID)
	})

	t.Run("negative user_id", func(t *testing.T) {
		d := validPrivateData()
		d.UserID = -1
		require.ErrorIs(t, v.Validate(ctx, d, FieldUserID), ErrInvalidUserID)
	})

	t.Run("empty metadata", func(t *testing.T) {
		d := validPrivateData()
		d.Payload.Metadata = ""
		require.ErrorIs(t, v.Validate(ctx, d, FieldMetadata), ErrEmptyMetadata)
	})

	t.Run("invalid type", func(t *testing.T) {
		d := validPrivateData()
		d.Payload.Type = models.DataType(999)
		require.ErrorIs(t, v.Validate(ctx, d, FieldType), ErrInvalidType)
	})

	t.Run("empty data", func(t *testing.T) {
		d := validPrivateData()
		d.Payload.Data = ""
		require.ErrorIs(t, v.Validate(ctx, d, FieldData), ErrEmptyData)
	})

	t.Run("empty hash", func(t *testing.T) {
		d := validPrivateData()
		d.Hash = ""
		require.ErrorIs(t, v.Validate(ctx, d, FieldHash), ErrInvalidHash)
	})

	t.Run("negative version", func(t *testing.T) {
		d := validPrivateData()
		d.Version = -1
		require.ErrorIs(t, v.Validate(ctx, d, FieldVersion), ErrInvalidVersion)
	})

	t.Run("version zero is valid for FieldVersion", func(t *testing.T) {
		d := validPrivateData()
		d.Version = 0
		require.NoError(t, v.Validate(ctx, d, FieldVersion))
	})

	t.Run("version non-zero fails FieldPrivateDataVersionForDataUpload", func(t *testing.T) {
		d := validPrivateData()
		d.Version = 5
		require.ErrorIs(t, v.Validate(ctx, d, FieldPrivateDataVersionForDataUpload), ErrInvalidVersion)
	})

	t.Run("version zero passes FieldPrivateDataVersionForDataUpload", func(t *testing.T) {
		d := validPrivateData()
		d.Version = 0
		require.NoError(t, v.Validate(ctx, d, FieldPrivateDataVersionForDataUpload))
	})

	t.Run("unknown field", func(t *testing.T) {
		d := validPrivateData()
		require.ErrorIs(t, v.Validate(ctx, d, "nonexistent"), ErrUnknownField)
	})

	t.Run("all data types accepted", func(t *testing.T) {
		for _, dt := range allowedDataTypes {
			d := validPrivateData()
			d.Payload.Type = dt
			require.NoError(t, v.Validate(ctx, d, FieldType), "DataType %d should be valid", dt)
		}
	})
}

// ---------------------------------------------------------------------------
// TestValidateUploadRequest
// ---------------------------------------------------------------------------

func TestValidateUploadRequest(t *testing.T) {
	v := NewPrivateDataValidator()
	ctx := context.Background()

	validItem := func() *models.PrivateData {
		d := validPrivateData()
		return &d
	}

	t.Run("valid with defaults", func(t *testing.T) {
		r := models.UploadRequest{
			UserID:          1,
			PrivateDataList: []*models.PrivateData{validItem()},
		}
		require.NoError(t, v.Validate(ctx, r))
	})

	t.Run("invalid user_id", func(t *testing.T) {
		r := models.UploadRequest{
			UserID:          0,
			PrivateDataList: []*models.PrivateData{validItem()},
		}
		require.ErrorIs(t, v.Validate(ctx, r, FieldUserID), ErrInvalidUserID)
	})

	t.Run("empty private data list", func(t *testing.T) {
		r := models.UploadRequest{
			UserID:          1,
			PrivateDataList: nil,
		}
		require.ErrorIs(t, v.Validate(ctx, r, FieldPrivateData), ErrEmptyPrivateData)
	})

	t.Run("invalid item in list returns indexed error", func(t *testing.T) {
		bad := validItem()
		bad.ClientSideID = ""
		r := models.UploadRequest{
			UserID:          1,
			PrivateDataList: []*models.PrivateData{validItem(), bad},
		}
		err := v.Validate(ctx, r, FieldPrivateData)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "index 1")
		assert.ErrorIs(t, err, ErrInvalidClientSideID)
	})

	t.Run("item with non-zero version fails upload", func(t *testing.T) {
		bad := validItem()
		bad.Version = 3
		r := models.UploadRequest{
			UserID:          1,
			PrivateDataList: []*models.PrivateData{bad},
		}
		err := v.Validate(ctx, r, FieldPrivateData)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidVersion)
	})

	t.Run("unknown field", func(t *testing.T) {
		r := models.UploadRequest{UserID: 1}
		require.ErrorIs(t, v.Validate(ctx, r, "bad_field"), ErrUnknownField)
	})
}

// ---------------------------------------------------------------------------
// TestValidateUpdateDataRequest
// ---------------------------------------------------------------------------

func TestValidateUpdateDataRequest(t *testing.T) {
	v := NewPrivateDataValidator()
	ctx := context.Background()

	t.Run("valid with defaults", func(t *testing.T) {
		r := models.UpdateRequest{
			UserID:             1,
			PrivateDataUpdates: []models.PrivateDataUpdate{validPrivateDataUpdate()},
		}
		require.NoError(t, v.Validate(ctx, r))
	})

	t.Run("invalid user_id", func(t *testing.T) {
		r := models.UpdateRequest{
			UserID:             0,
			PrivateDataUpdates: []models.PrivateDataUpdate{validPrivateDataUpdate()},
		}
		require.ErrorIs(t, v.Validate(ctx, r, FieldUserID), ErrInvalidUserID)
	})

	t.Run("empty updates list", func(t *testing.T) {
		r := models.UpdateRequest{
			UserID:             1,
			PrivateDataUpdates: nil,
		}
		require.ErrorIs(t, v.Validate(ctx, r, FieldPrivateDataUpdates), ErrEmptyUpdates)
	})

	t.Run("invalid update in list returns indexed error", func(t *testing.T) {
		bad := validPrivateDataUpdate()
		bad.ClientSideID = ""
		r := models.UpdateRequest{
			UserID:             1,
			PrivateDataUpdates: []models.PrivateDataUpdate{validPrivateDataUpdate(), bad},
		}
		err := v.Validate(ctx, r, FieldPrivateDataUpdates)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "index 1")
		assert.ErrorIs(t, err, ErrInvalidClientSideID)
	})

	t.Run("unknown field", func(t *testing.T) {
		r := models.UpdateRequest{UserID: 1}
		require.ErrorIs(t, v.Validate(ctx, r, "bad_field"), ErrUnknownField)
	})
}

// ---------------------------------------------------------------------------
// TestValidatePrivateDataUpdate
// ---------------------------------------------------------------------------

func TestValidatePrivateDataUpdate(t *testing.T) {
	v := NewPrivateDataValidator()
	ctx := context.Background()

	t.Run("valid with defaults", func(t *testing.T) {
		u := validPrivateDataUpdate()
		require.NoError(t, v.Validate(ctx, u))
	})

	t.Run("empty client_side_id", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.ClientSideID = ""
		require.ErrorIs(t, v.Validate(ctx, u, FieldClientSideID), ErrInvalidClientSideID)
	})

	t.Run("empty metadata pointer", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.FieldsUpdate.Metadata = ptrMeta("")
		require.ErrorIs(t, v.Validate(ctx, u, FieldMetadata), ErrEmptyMetadata)
	})

	t.Run("nil metadata is OK", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.FieldsUpdate.Metadata = nil
		u.FieldsUpdate.Data = ptrData("d")
		require.NoError(t, v.Validate(ctx, u, FieldMetadata))
	})

	t.Run("empty data pointer", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.FieldsUpdate.Data = ptrData("")
		require.ErrorIs(t, v.Validate(ctx, u, FieldData), ErrEmptyData)
	})

	t.Run("nil data is OK", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.FieldsUpdate.Data = nil
		require.NoError(t, v.Validate(ctx, u, FieldData))
	})

	t.Run("empty updated_record_hash", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.UpdatedRecordHash = ""
		require.ErrorIs(t, v.Validate(ctx, u, FieldUpdatedRecordHash), ErrInvalidUpdatedRecordHash)
	})

	t.Run("zero version", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.Version = 0
		require.ErrorIs(t, v.Validate(ctx, u, FieldVersion), ErrInvalidUpdateVersion)
	})

	t.Run("negative version", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.Version = -1
		require.ErrorIs(t, v.Validate(ctx, u, FieldVersion), ErrInvalidUpdateVersion)
	})

	t.Run("no fields to update", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.FieldsUpdate = models.FieldsUpdate{}
		require.ErrorIs(t, v.Validate(ctx, u), ErrNoFieldsToUpdate)
	})

	t.Run("only notes is enough", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.FieldsUpdate = models.FieldsUpdate{Notes: ptrNotes("n")}
		require.NoError(t, v.Validate(ctx, u))
	})

	t.Run("only additional_fields is enough", func(t *testing.T) {
		u := validPrivateDataUpdate()
		u.FieldsUpdate = models.FieldsUpdate{AdditionalFields: ptrFields("f")}
		require.NoError(t, v.Validate(ctx, u))
	})

	t.Run("unknown field", func(t *testing.T) {
		u := validPrivateDataUpdate()
		require.ErrorIs(t, v.Validate(ctx, u, "bad_field"), ErrUnknownField)
	})

	t.Run("pointer receiver dispatches correctly", func(t *testing.T) {
		u := validPrivateDataUpdate()
		require.NoError(t, v.Validate(ctx, &u))
	})
}

// ---------------------------------------------------------------------------
// TestValidateDeleteDataRequest
// ---------------------------------------------------------------------------

func TestValidateDeleteDataRequest(t *testing.T) {
	v := NewPrivateDataValidator()
	ctx := context.Background()

	t.Run("valid with defaults (no delete_entries field checked)", func(t *testing.T) {
		r := models.DeleteRequest{UserID: 1}
		require.ErrorIs(t, v.Validate(ctx, r), ErrNoDeleteEntries)
	})

	t.Run("invalid user_id", func(t *testing.T) {
		r := models.DeleteRequest{UserID: 0}
		require.ErrorIs(t, v.Validate(ctx, r, FieldUserID), ErrInvalidUserID)
	})

	t.Run("empty delete entries", func(t *testing.T) {
		r := models.DeleteRequest{UserID: 1}
		require.ErrorIs(t, v.Validate(ctx, r, FieldDeleteEntries), ErrNoDeleteEntries)
	})

	t.Run("unknown field", func(t *testing.T) {
		r := models.DeleteRequest{UserID: 1}
		require.ErrorIs(t, v.Validate(ctx, r, "bad_field"), ErrUnknownField)
	})

	t.Run("pointer receiver", func(t *testing.T) {
		r := models.DeleteRequest{UserID: 1}
		require.ErrorIs(t, v.Validate(ctx, &r), ErrNoDeleteEntries)
	})
}

// ---------------------------------------------------------------------------
// TestValidateDownloadDataRequest
// ---------------------------------------------------------------------------

func TestValidateDownloadDataRequest(t *testing.T) {
	v := NewPrivateDataValidator()
	ctx := context.Background()

	t.Run("valid with defaults", func(t *testing.T) {
		r := models.DownloadRequest{UserID: 1}
		require.NoError(t, v.Validate(ctx, r))
	})

	t.Run("invalid user_id", func(t *testing.T) {
		r := models.DownloadRequest{UserID: 0}
		require.ErrorIs(t, v.Validate(ctx, r, FieldUserID), ErrInvalidUserID)
	})

	t.Run("unknown field", func(t *testing.T) {
		r := models.DownloadRequest{UserID: 1}
		require.ErrorIs(t, v.Validate(ctx, r, "bad_field"), ErrUnknownField)
	})

	t.Run("pointer receiver", func(t *testing.T) {
		r := models.DownloadRequest{UserID: 1}
		require.NoError(t, v.Validate(ctx, &r))
	})
}

// ---------------------------------------------------------------------------
// TestIsValidDataType
// ---------------------------------------------------------------------------

func TestIsValidDataType(t *testing.T) {
	for _, dt := range allowedDataTypes {
		assert.True(t, isValidDataType(dt), "expected %d to be valid", dt)
	}
	assert.False(t, isValidDataType(models.DataType(0)))
	assert.False(t, isValidDataType(models.DataType(999)))
	assert.False(t, isValidDataType(models.DataType(-1)))
}
