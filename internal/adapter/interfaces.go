package adapter

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

// ServerAdapter defines transport-agnostic communication with GoPassKeeper server.
type ServerAdapter interface {
	SetToken(token string)
	Token() string

	Register(ctx context.Context, user models.User) (models.User, error)
	Login(ctx context.Context, user models.User) (models.Token, error)

	Upload(ctx context.Context, req models.UploadRequest) error
	Download(ctx context.Context, req models.DownloadRequest) ([]models.PrivateData, error)
	Update(ctx context.Context, req models.UpdateRequest) error
	Delete(ctx context.Context, req models.DeleteRequest) error

	GetServerStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error)
}
