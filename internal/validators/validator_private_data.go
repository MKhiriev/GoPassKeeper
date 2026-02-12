package validators

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type PrivateDataValidator struct {
}

func NewPrivateDataValidator() Validator {
	return &PrivateDataValidator{}
}

func (v *PrivateDataValidator) Validate(ctx context.Context, obj any, fields ...string) error {
	// TODO implement me!
	panic("implement me!")
}

func (v *PrivateDataValidator) validatePrivateData(ctx context.Context, data models.PrivateData) error {
	// TODO implement me!
	panic("implement me!")
}

func (v *PrivateDataValidator) validateUpdateDataRequest(ctx context.Context, request models.UpdateRequest) error {
	// TODO implement me!
	panic("implement me!")
}

func (v *PrivateDataValidator) validateDeleteDataRequest(ctx context.Context, request models.DeleteRequest) error {
	// TODO implement me!
	panic("implement me!")
}

func (v *PrivateDataValidator) validateDownloadDataRequest(ctx context.Context, request models.DownloadRequest) error {
	// TODO implement me!
	panic("implement me!")
}
