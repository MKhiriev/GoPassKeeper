package service

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type PrivateDataService interface {
	UploadPrivateData(ctx context.Context, data ...models.PrivateData) error

	DownloadPrivateData(ctx context.Context, downloadRequests ...models.DownloadRequest) ([]models.PrivateData, error)
	DownloadAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error)

	UpdatePrivateData(ctx context.Context, updateRequests ...models.UpdateRequest) error
	DeletePrivateData(ctx context.Context, deleteRequests ...models.DeleteRequest) error
}

type AuthService interface {
	RegisterUser(ctx context.Context, user models.User) (models.User, error)
	Login(ctx context.Context, user models.User) (models.User, error)
	CreateToken(ctx context.Context, user models.User) (models.Token, error)
	ParseToken(ctx context.Context, tokenString string) (models.Token, error)
}

// PrivateDataServiceWrapper defines middleware composition for PrivateDataService.
// Implementations wrap an existing PrivateDataService to add behavior such as
// logging or validating.
type PrivateDataServiceWrapper interface {
	Wrap(PrivateDataService) PrivateDataService // returns a decorated PrivateDataService applying additional behavior
}
