package validators

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/models"
)

// Field name constants used to specify which fields should be validated.
// These constants are passed to Validate or internal validation methods
// to restrict validation to a subset of fields (field-level scoping).
const (
	// FieldClientSideID targets the client-generated unique identifier of a vault item.
	FieldClientSideID = "client_side_id"

	// FieldUserID targets the owner identifier of a vault item or request.
	FieldUserID = "user_id"

	// FieldMetadata targets the encrypted metadata field of a vault item payload.
	FieldMetadata = "metadata"

	// FieldType targets the semantic data type field (e.g. login, text, binary, bank card).
	FieldType = "type"

	// FieldData targets the encrypted data payload field of a vault item.
	FieldData = "data"

	// FieldHash targets the integrity checksum field of a vault item.
	FieldHash = "hash"

	// FieldVersion targets the optimistic concurrency version field of a vault item.
	FieldVersion = "version"

	// FieldClientSideIDs targets the array of client-side identifiers in bulk requests.
	FieldClientSideIDs = "client_side_ids"

	// FieldDeleteEntries targets the list of entries to be soft-deleted.
	FieldDeleteEntries = "delete_entries"

	// FieldPrivateData targets the list of vault items in an upload request.
	FieldPrivateData = "private_data"

	// FieldPrivateDataVersionForDataUpload enforces that version must be zero
	// for newly uploaded vault items (initial creation).
	FieldPrivateDataVersionForDataUpload = "version for data upload"

	// FieldPrivateDataUpdates targets the list of update descriptors in a batch update request.
	FieldPrivateDataUpdates = "private_data_updates"

	// FieldUpdatedRecordHash targets the post-update integrity hash
	// that the client computes from the merged record state.
	FieldUpdatedRecordHash = "updated_record_hash"
)

// allowedDataTypes is the exhaustive set of DataType values accepted by the validator.
// Any DataType not present in this slice is considered invalid.
var allowedDataTypes = []models.DataType{
	models.LoginPassword,
	models.Text,
	models.Binary,
	models.BankCard,
}

// PrivateDataValidator implements the Validator interface for all
// private-data-related domain models: PrivateData, UploadRequest,
// UpdateRequest, PrivateDataUpdate, DeleteRequest, and DownloadRequest.
//
// It supports both value and pointer receivers for every model type
// and allows optional field-level scoping via variadic field name arguments.
type PrivateDataValidator struct {
}

// NewPrivateDataValidator constructs a new PrivateDataValidator
// and returns it as the Validator interface.
func NewPrivateDataValidator() Validator {
	return &PrivateDataValidator{}
}

// Validate dispatches validation to the appropriate type-specific method
// based on the dynamic type of obj. Both value and pointer forms of each
// supported model are accepted.
//
// Supported types:
//   - models.PrivateData / *models.PrivateData
//   - models.UploadRequest / *models.UploadRequest
//   - models.UpdateRequest / *models.UpdateRequest
//   - models.PrivateDataUpdate / *models.PrivateDataUpdate
//   - models.DeleteRequest / *models.DeleteRequest
//   - models.DownloadRequest / *models.DownloadRequest
//
// Returns ErrUnsupportedType if obj does not match any known model.
// Optional fields restrict validation to the named subset; when omitted,
// a sensible default set of fields is validated.
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

// isValidDataType reports whether dt is one of the recognized DataType values
// defined in allowedDataTypes.
func isValidDataType(dt models.DataType) bool {
	for _, t := range allowedDataTypes {
		if dt == t {
			return true
		}
	}
	return false
}

// validatePrivateData validates a single PrivateData model.
//
// Default validated fields (when none specified):
// ClientSideID, UserID, Metadata, Type, Data, Hash, Version.
//
// Special field FieldPrivateDataVersionForDataUpload enforces Version == 0
// for newly created records.
//
// Returns the first encountered validation error or nil.
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

// validateUploadRequest validates an UploadRequest, which contains a batch
// of new vault items to be persisted.
//
// Default validated fields: UserID, PrivateData.
//
// When FieldPrivateData is validated, each item in PrivateDataList is
// individually checked with validatePrivateData using the upload-specific
// field set (including FieldPrivateDataVersionForDataUpload to ensure
// version is zero for new records).
//
// Returns a wrapped error indicating the index of the first invalid item.
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

// validateUpdateDataRequest validates an UpdateRequest, which contains
// a batch of partial updates to existing vault items.
//
// Default validated fields: UserID, PrivateDataUpdates.
//
// When FieldPrivateDataUpdates is validated, each PrivateDataUpdate
// is individually checked with validatePrivateDataUpdate.
//
// Returns a wrapped error indicating the index of the first invalid update.
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

// validatePrivateDataUpdate validates a single PrivateDataUpdate descriptor.
//
// Default validated fields: ClientSideID, Metadata, Data, Version, UpdatedRecordHash.
//
// Field-level checks for Metadata and Data only trigger when the corresponding
// pointer is non-nil (partial update semantics: nil means "do not touch").
//
// After field-level checks, an additional structural rule is enforced:
// at least one payload field (Metadata, Data, Notes, or AdditionalFields)
// must be non-nil. Returns ErrNoFieldsToUpdate otherwise.
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
			if update.Version < 0 {
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

// validateDeleteDataRequest validates a DeleteRequest, which contains
// a list of vault items to be soft-deleted.
//
// Default validated fields: UserID, ClientSideIDs.
//
// When FieldDeleteEntries is validated, each entry is checked for
// a non-empty ClientSideID and a non-negative Version.
func (v *PrivateDataValidator) validateDeleteDataRequest(ctx context.Context, request models.DeleteRequest, fields ...string) error {
	if len(fields) == 0 {
		fields = []string{FieldUserID, FieldDeleteEntries}
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
				if deleteEntry.ClientSideID == "" {
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

// validateDownloadDataRequest validates a DownloadRequest, which specifies
// search criteria for querying vault items by owner and optional client-side IDs.
//
// Default validated fields: UserID, ClientSideIDs.
//
// When FieldClientSideIDs is validated, each entry in the list is checked
// for a non-empty value.
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
				if clientSideID == "" {
					return ErrInvalidClientSideID
				}
			}
		default:
			return ErrUnknownField
		}
	}

	return nil
}
