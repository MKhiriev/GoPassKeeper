package validators

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/models"
)

const (
	FieldClientSideID                    = "client_side_id"
	FieldUserID                          = "user_id"
	FieldMetadata                        = "metadata"
	FieldType                            = "type"
	FieldData                            = "data"
	FieldHash                            = "hash"
	FieldVersion                         = "version"
	FieldClientSideIDs                   = "client_side_ids"
	FieldDeleteEntries                   = "delete_entries"
	FieldPrivateData                     = "private_data"
	FieldPrivateDataVersionForDataUpload = "version for data upload"
	FieldPrivateDataUpdates              = "private_data_updates"
	FieldUpdatedRecordHash               = "updated_record_hash"
)

var allowedDataTypes = []models.DataType{
	models.LoginPassword,
	models.Text,
	models.Binary,
	models.BankCard,
}

type PrivateDataValidator struct {
}

func NewPrivateDataValidator() Validator {
	return &PrivateDataValidator{}
}

func (v *PrivateDataValidator) Validate(ctx context.Context, obj any, fields ...string) error {
	switch value := obj.(type) {
	case models.PrivateData:
		return v.validatePrivateData(ctx, value, fields...)
	case *models.PrivateData:
		return v.validatePrivateData(ctx, *value, fields...)

	case models.UploadRequest:
		return v.validateUploadRequest(ctx, value, fields...)
	case *models.UploadRequest:
		return v.validateUploadRequest(ctx, *value, fields...)

	case models.UpdateRequest:
		return v.validateUpdateDataRequest(ctx, value, fields...)
	case *models.UpdateRequest:
		return v.validateUpdateDataRequest(ctx, *value, fields...)

	case models.PrivateDataUpdate:
		return v.validatePrivateDataUpdate(ctx, value, fields...)
	case *models.PrivateDataUpdate:
		return v.validatePrivateDataUpdate(ctx, *value, fields...)

	case models.DeleteRequest:
		return v.validateDeleteDataRequest(ctx, value, fields...)
	case *models.DeleteRequest:
		return v.validateDeleteDataRequest(ctx, *value, fields...)

	case models.DownloadRequest:
		return v.validateDownloadDataRequest(ctx, value, fields...)
	case *models.DownloadRequest:
		return v.validateDownloadDataRequest(ctx, *value, fields...)

	default:
		return ErrUnsupportedType
	}
}

func isValidDataType(dt models.DataType) bool {
	for _, t := range allowedDataTypes {
		if dt == t {
			return true
		}
	}
	return false
}

// checked!
func (v *PrivateDataValidator) validatePrivateData(ctx context.Context, data models.PrivateData, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldClientSideID, FieldUserID, FieldMetadata, FieldType, FieldData, FieldHash, FieldVersion}
	}

	for _, f := range fields {
		switch f {
		case FieldClientSideID:
			if data.ClientSideID == "" {
				return ErrInvalidClientSideID
			}
		case FieldUserID:
			if data.UserID <= 0 {
				return ErrInvalidUserID
			}
		case FieldMetadata:
			if len(data.Payload.Metadata) == 0 {
				return ErrEmptyMetadata
			}
		case FieldType:
			if !isValidDataType(data.Payload.Type) {
				return ErrInvalidType
			}
		case FieldData:
			if len(data.Payload.Data) == 0 {
				return ErrEmptyData
			}
		case FieldHash:
			if data.Hash == "" {
				return ErrInvalidHash
			}
		case FieldVersion:
			if data.Version < 0 {
				return ErrInvalidVersion
			}
		case FieldPrivateDataVersionForDataUpload:
			if data.Version != 0 {
				return ErrInvalidVersion
			}
		default:
			return ErrUnknownField
		}
	}

	return nil
}

// checked!
func (v *PrivateDataValidator) validateUploadRequest(ctx context.Context, request models.UploadRequest, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldUserID, FieldPrivateData}
	}

	for _, f := range fields {
		switch f {
		case FieldUserID:
			if request.UserID <= 0 {
				return ErrInvalidUserID
			}
		case FieldPrivateData:
			if len(request.PrivateDataList) == 0 {
				return ErrEmptyPrivateData
			}
			for i, data := range request.PrivateDataList {
				if err := v.validatePrivateData(ctx, *data, FieldClientSideID, FieldUserID, FieldMetadata, FieldType, FieldData, FieldHash, FieldPrivateDataVersionForDataUpload); err != nil {
					return fmt.Errorf("validation error at index %d: %w", i, err)
				}
			}
		default:
			return ErrUnknownField
		}
	}

	return nil
}

// checked!
func (v *PrivateDataValidator) validateUpdateDataRequest(ctx context.Context, request models.UpdateRequest, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldUserID, FieldPrivateDataUpdates}
	}

	for _, f := range fields {
		switch f {
		case FieldUserID:
			if request.UserID <= 0 {
				return ErrInvalidUserID
			}
		case FieldPrivateDataUpdates:
			if len(request.PrivateDataUpdates) == 0 {
				return ErrEmptyUpdates
			}
			for i, update := range request.PrivateDataUpdates {
				if err := v.validatePrivateDataUpdate(ctx, update); err != nil {
					return fmt.Errorf("validation error at index %d: %w", i, err)
				}
			}
		default:
			return ErrUnknownField
		}
	}

	return nil
}

// checked!
func (v *PrivateDataValidator) validatePrivateDataUpdate(ctx context.Context, update models.PrivateDataUpdate, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldClientSideID, FieldMetadata, FieldData, FieldVersion, FieldVersion, FieldUpdatedRecordHash}
	}

	for _, f := range fields {
		switch f {
		case FieldClientSideID:
			if update.ClientSideID == "" {
				return ErrInvalidClientSideID
			}
		case FieldMetadata:
			if update.FieldsUpdate.Metadata != nil && len(*update.FieldsUpdate.Metadata) == 0 {
				return ErrEmptyMetadata
			}
		case FieldData:
			if update.FieldsUpdate.Data != nil && len(*update.FieldsUpdate.Data) == 0 {
				return ErrEmptyData
			}
		case FieldUpdatedRecordHash:
			if update.UpdatedRecordHash == "" {
				return ErrInvalidUpdatedRecordHash
			}
		case FieldVersion:
			if update.Version <= 0 {
				return ErrInvalidUpdateVersion
			}
		default:
			return ErrUnknownField
		}
	}

	if update.FieldsUpdate.Metadata == nil && update.FieldsUpdate.Data == nil && update.FieldsUpdate.Notes == nil && update.FieldsUpdate.AdditionalFields == nil {
		return ErrNoFieldsToUpdate
	}

	return nil
}

// checked!
func (v *PrivateDataValidator) validateDeleteDataRequest(ctx context.Context, request models.DeleteRequest, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldUserID, FieldClientSideIDs}
	}

	for _, f := range fields {
		switch f {
		case FieldUserID:
			if request.UserID <= 0 {
				return ErrInvalidUserID
			}
		case FieldDeleteEntries:
			if len(request.DeleteEntries) == 0 {
				return ErrNoDeleteEntries
			}
			for _, deleteEntry := range request.DeleteEntries {
				if deleteEntry.ClientSideID != "" {
					return ErrInvalidClientSideID
				}
				if deleteEntry.Version < 0 {
					return ErrInvalidVersion
				}
			}
		default:
			return ErrUnknownField
		}
	}

	return nil
}

// checked!
func (v *PrivateDataValidator) validateDownloadDataRequest(ctx context.Context, request models.DownloadRequest, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldUserID, FieldClientSideIDs}
	}

	for _, f := range fields {
		switch f {
		case FieldUserID:
			if request.UserID <= 0 {
				return ErrInvalidUserID
			}
		case FieldClientSideIDs:
			for _, clientSideID := range request.ClientSideIDs {
				if clientSideID != "" {
					return ErrInvalidClientSideID
				}
			}
		default:
			return ErrUnknownField
		}
	}

	return nil
}
