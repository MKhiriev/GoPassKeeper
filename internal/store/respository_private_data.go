package store

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type privateDataRepository struct {
	*DB
}

func NewPrivateDataRepository(cfg config.DBConfig, log *logger.Logger) (PrivateDataRepository, error) {
	db, err := NewConnectPostgres(context.Background(), cfg, log)
	if err != nil {
		log.Err(err).Msg("connection to database failed")
		return nil, err
	}

	return &privateDataRepository{db}, nil
}

func (p *privateDataRepository) SavePrivateData(ctx context.Context, data ...models.PrivateData) error {
	log := logger.FromContext(ctx)

	saveErr := p.savePrivateData(ctx, data...)
	if saveErr != nil {
		log.Err(saveErr).Str("func", "privateDataRepository.SavePrivateData").Msg("error saving private data to a repository")
		return saveErr
	}

	return nil
}

func (p *privateDataRepository) GetPrivateData(ctx context.Context, downloadRequests ...models.DownloadRequest) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) GetAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) UpdatePrivateData(ctx context.Context, updateRequests ...models.UpdateRequest) error {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) DeletePrivateData(ctx context.Context, deleteRequests ...models.DeleteRequest) error {
	//TODO implement me
	panic("implement me")
}

func (p *privateDataRepository) savePrivateData(ctx context.Context, data ...models.PrivateData) error {
	log := logger.FromContext(ctx)

	if len(data) == 1 {
		dataToSave := data[0]
		result, err := p.DB.ExecContext(ctx, savePrivateData, dataToSave.ID, dataToSave.UserID, dataToSave.Metadata, dataToSave.Type, dataToSave.Data, dataToSave.Notes, dataToSave.AdditionalFields, dataToSave.CreatedAt)
		if err != nil {
			log.Err(err).Str("func", "*DB.savePrivateData").Msg("error executing query for saving private data")
			return err
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			log.Error().Str("func", "*DB.savePrivateData").Msg("provided private data was not saved")
			return ErrPrivateDataNotSaved
		}

		return nil
	}

	// begin transaction
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Err(err).Str("func", "*DB.savePrivateData").Msg("error during opening transaction")
		return fmt.Errorf("error during opening transaction: %w", err)
	}
	defer tx.Rollback()

	// prepare context
	stmt, err := tx.PrepareContext(ctx, savePrivateData)
	if err != nil {
		log.Err(err).Str("func", "*DB.savePrivateData").Msg("error during preparing context")
		return err
	}
	defer stmt.Close()

	// for each single private data object
	for idx, singleData := range data {
		log.Debug().Str("func", "*DB.savePrivateData").Int("iteration", idx).Any("private data id", singleData.ID).Msg("trying to save private data")

		result, statementExecutionError := stmt.ExecContext(ctx, singleData.ID, singleData.UserID, singleData.Metadata, singleData.Type, singleData.Data, singleData.Notes, singleData.AdditionalFields, singleData.CreatedAt)
		if statementExecutionError != nil {
			log.Err(statementExecutionError).Str("func", "*DB.savePrivateData").Msg("error executing prepared query for saving metric")
			return statementExecutionError
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			log.Err(err).Str("func", "*DB.savePrivateData").Msg("all provided private data was not saved")
			return ErrPrivateDataNotSaved
		}
	}

	return tx.Commit()
}
