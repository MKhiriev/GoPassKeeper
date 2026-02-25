package tui

import (
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/textinput"
)

type formTextModel struct {
	inputs     []textinput.Model
	focus      int
	editing    bool
	clientID   string
	submitting bool
}

func newFormTextModel(item *models.DecipheredPayload) formTextModel {
	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Width = 50
	}
	inputs[0].Focus()

	m := formTextModel{inputs: inputs}
	if item == nil {
		return m
	}

	m.editing = true
	m.clientID = item.ClientSideID
	m.inputs[0].SetValue(item.Metadata.Name)
	if item.TextData != nil {
		m.inputs[1].SetValue(item.TextData.Text)
	}
	if item.Notes != nil {
		m.inputs[2].SetValue(item.Notes.Notes)
	}
	return m
}

func (m formTextModel) toPayload(userID int64) models.DecipheredPayload {
	return models.DecipheredPayload{
		ClientSideID: m.clientID,
		UserID:       userID,
		Metadata:     models.Metadata{Name: m.inputs[0].Value()},
		Type:         models.Text,
		TextData:     &models.TextData{Text: m.inputs[1].Value()},
		Notes:        &models.Notes{Notes: m.inputs[2].Value()},
	}
}

func (m formTextModel) View() string {
	title := "Новая заметка"
	if m.editing {
		title = "Редактирование: " + m.inputs[0].Value()
	}

	out := title + "\n\n"
	out += "Название: [" + m.inputs[0].View() + "]\n"
	out += "Текст:    [" + m.inputs[1].View() + "]\n"
	out += "Заметки:  [" + m.inputs[2].View() + "]\n\n"
	out += "esc отмена  tab следующее поле  enter сохранить"
	return out
}
