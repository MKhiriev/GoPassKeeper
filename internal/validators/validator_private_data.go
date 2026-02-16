package validators

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/models"
)

const (
	FieldID                 = "id"
	FieldUserID             = "user_id"
	FieldMetadata           = "metadata"
	FieldType               = "type"
	FieldData               = "data"
	FieldIDs                = "ids"
	FieldTypes              = "types"
	FieldPrivateData        = "private_data"
	FieldPrivateDataUpdates = "private_data_updates"
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

func (v *PrivateDataValidator) validatePrivateData(ctx context.Context, data models.PrivateData, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldUserID, FieldMetadata, FieldType, FieldData}
	}

	for _, f := range fields {
		switch f {
		case FieldUserID:
			if data.UserID <= 0 {
				return ErrInvalidUserID
			}
		case FieldMetadata:
			if len(data.Metadata) == 0 {
				return ErrEmptyMetadata
			}
		case FieldType:
			if !isValidDataType(data.Type) {
				return ErrInvalidType
			}
		case FieldData:
			if len(data.Data) == 0 {
				return ErrEmptyData
			}
		default:
			return ErrUnknownField
		}
	}

	return nil
}

func (v *PrivateDataValidator) validateUploadRequest(ctx context.Context, request models.UploadRequest, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldPrivateData}
	}

	for _, f := range fields {
		switch f {
		case FieldPrivateData:
			if len(request.PrivateData) == 0 {
				return ErrEmptyPrivateData
			}
			for i, data := range request.PrivateData {
				if err := v.validatePrivateData(ctx, *data); err != nil {
					return fmt.Errorf("validation error at index %d: %w", i, err)
				}
			}
		default:
			return ErrUnknownField
		}
	}

	return nil
}

func (v *PrivateDataValidator) validateUpdateDataRequest(ctx context.Context, request models.UpdateRequest, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldPrivateDataUpdates}
	}

	for _, f := range fields {
		switch f {
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

func (v *PrivateDataValidator) validatePrivateDataUpdate(ctx context.Context, update models.PrivateDataUpdate, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldID, FieldUserID, FieldMetadata, FieldType, FieldData}
	}

	for _, f := range fields {
		switch f {
		case FieldID:
			if update.ID <= 0 {
				return ErrInvalidID
			}
		case FieldUserID:
			if update.UserID <= 0 {
				return ErrInvalidUserID
			}
		case FieldMetadata:
			if update.Metadata != nil && len(*update.Metadata) == 0 {
				return ErrEmptyMetadata
			}
		case FieldType:
			if update.Type != nil && !isValidDataType(*update.Type) {
				return ErrInvalidType
			}
		case FieldData:
			if update.Data != nil && len(*update.Data) == 0 {
				return ErrEmptyData
			}
		default:
			return ErrUnknownField
		}
	}

	if update.Metadata == nil && update.Type == nil && update.Data == nil && update.Notes == nil && update.AdditionalFields == nil {
		return ErrNoFieldsToUpdate
	}

	return nil
}

func (v *PrivateDataValidator) validateDeleteDataRequest(ctx context.Context, request models.DeleteRequest, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldUserID, FieldIDs}
	}

	for _, f := range fields {
		switch f {
		case FieldUserID:
			if request.UserID <= 0 {
				return ErrInvalidUserID
			}
		case FieldIDs:
			if len(request.IDs) == 0 {
				return ErrEmptyIDs
			}
			for _, id := range request.IDs {
				if id <= 0 {
					return ErrInvalidID
				}
			}
		default:
			return ErrUnknownField
		}
	}

	return nil
}

func (v *PrivateDataValidator) validateDownloadDataRequest(ctx context.Context, request models.DownloadRequest, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldUserID}
	}

	for _, f := range fields {
		switch f {
		case FieldUserID:
			if request.UserID <= 0 {
				return ErrInvalidUserID
			}
		case FieldIDs:
			for _, id := range request.IDs {
				if id <= 0 {
					return ErrInvalidID
				}
			}
		case FieldTypes:
			for _, dataType := range request.Types {
				if !isValidDataType(dataType) {
					return ErrInvalidType
				}
			}
		default:
			return ErrUnknownField
		}
	}

	return nil
}
