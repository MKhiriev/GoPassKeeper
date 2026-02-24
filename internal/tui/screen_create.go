package tui

import (
	"context"
	"strconv"

	"github.com/MKhiriev/go-pass-keeper/models"
)

func (t *TUI) showCreateScreen(ctx context.Context, userID int64) error {
	payload, err := t.promptPayload(models.PrivateDataPayload{})
	if err != nil {
		return err
	}

	if err = t.dataSvc.Create(ctx, userID, payload); err != nil {
		return err
	}

	t.printLn("created")
	return nil
}

func (t *TUI) promptPayload(base models.PrivateDataPayload) (models.PrivateDataPayload, error) {
	metaPrompt := "Metadata: "
	if base.Metadata != "" {
		metaPrompt = "Metadata (empty keep): "
	}
	metadata, err := t.readLine(metaPrompt)
	if err != nil {
		return models.PrivateDataPayload{}, err
	}
	if metadata == "" {
		metadata = string(base.Metadata)
	}

	dataPrompt := "Data: "
	if base.Data != "" {
		dataPrompt = "Data (empty keep): "
	}
	data, err := t.readLine(dataPrompt)
	if err != nil {
		return models.PrivateDataPayload{}, err
	}
	if data == "" {
		data = string(base.Data)
	}

	typePrompt := "Type (1=login,2=text,3=binary,4=card): "
	if base.Type != 0 {
		typePrompt = "Type (empty keep): "
	}
	typeValue, err := t.readLine(typePrompt)
	if err != nil {
		return models.PrivateDataPayload{}, err
	}
	dt := base.Type
	if typeValue != "" {
		parsed, convErr := strconv.Atoi(typeValue)
		if convErr != nil {
			return models.PrivateDataPayload{}, convErr
		}
		dt = models.DataType(parsed)
	}
	if dt == 0 {
		dt = models.Text
	}

	notesValue, err := t.readLine("Notes (empty skip): ")
	if err != nil {
		return models.PrivateDataPayload{}, err
	}
	fieldsValue, err := t.readLine("Additional fields (json/text, empty skip): ")
	if err != nil {
		return models.PrivateDataPayload{}, err
	}

	payload := models.PrivateDataPayload{Metadata: models.CipheredMetadata(metadata), Type: dt, Data: models.CipheredData(data)}
	if notesValue != "" {
		n := models.CipheredNotes(notesValue)
		payload.Notes = &n
	} else if base.Notes != nil {
		payload.Notes = base.Notes
	}
	if fieldsValue != "" {
		f := models.CipheredCustomFields(fieldsValue)
		payload.AdditionalFields = &f
	} else if base.AdditionalFields != nil {
		payload.AdditionalFields = base.AdditionalFields
	}

	return payload, nil
}
