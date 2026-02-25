package tui

import (
	"os"

	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/textinput"
)

type formBinaryModel struct {
	inputs     []textinput.Model
	focus      int
	editing    bool
	clientID   string
	submitting bool
}

func newFormBinaryModel(item *models.DecipheredPayload) formBinaryModel {
	inputs := make([]textinput.Model, 4)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Width = 50
	}
	inputs[0].Focus()

	m := formBinaryModel{inputs: inputs}
	if item == nil {
		return m
	}

	m.editing = true
	m.clientID = item.ClientSideID
	m.inputs[0].SetValue(item.Metadata.Name)
	if item.BinaryData != nil {
		m.inputs[1].SetValue(item.BinaryData.FileName)
	}
	if item.Notes != nil {
		m.inputs[3].SetValue(item.Notes.Notes)
	}
	return m
}

func (m formBinaryModel) toPayload(userID int64) models.DecipheredPayload {
	size := int64(0)
	if st, err := os.Stat(m.inputs[2].Value()); err == nil {
		size = st.Size()
	}

	return models.DecipheredPayload{
		ClientSideID: m.clientID,
		UserID:       userID,
		Metadata:     models.Metadata{Name: m.inputs[0].Value()},
		Type:         models.Binary,
		BinaryData: &models.BinaryData{
			FileName: m.inputs[1].Value(),
			Size:     size,
		},
		Notes: &models.Notes{Notes: m.inputs[3].Value()},
	}
}

func (m formBinaryModel) View() string {
	title := "Новый файл"
	if m.editing {
		title = "Редактирование: " + m.inputs[0].Value()
	}

	out := title + "\n\n"
	out += "Название:      [" + m.inputs[0].View() + "]\n"
	out += "Имя файла:     [" + m.inputs[1].View() + "]\n"
	out += "Путь к файлу:  [" + m.inputs[2].View() + "]\n"
	out += "Заметки:       [" + m.inputs[3].View() + "]\n\n"
	out += "esc отмена  tab следующее поле  enter сохранить"
	return out
}
