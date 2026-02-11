package validators

import "context"

type PrivateDataValidator struct {
}

func NewPrivateDataValidator() Validator {
	return &PrivateDataValidator{}
}

func (v *PrivateDataValidator) Validate(ctx context.Context, obj any, fields ...string) error {
	// TODO implement me!
	panic("implement me!")
}
