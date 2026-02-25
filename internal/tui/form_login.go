package tui

import (
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/textinput"
)

type formLoginModel struct {
	inputs     []textinput.Model
	focus      int
	editing    bool
	clientID   string
	submitting bool
}

func newFormLoginModel(item *models.DecipheredPayload) formLoginModel {
	inputs := make([]textinput.Model, 6)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Width = 40
	}
	inputs[2].EchoMode = textinput.EchoPassword
	inputs[2].EchoCharacter = '*'
	inputs[0].Focus()

	m := formLoginModel{inputs: inputs}
	if item == nil {
		return m
	}

	m.editing = true
	m.clientID = item.ClientSideID
	m.inputs[0].SetValue(item.Metadata.Name)
	if item.LoginData != nil {
		m.inputs[1].SetValue(item.LoginData.Username)
		m.inputs[2].SetValue(item.LoginData.Password)
		if len(item.LoginData.URIs) > 0 {
			m.inputs[3].SetValue(item.LoginData.URIs[0].URI)
		}
		if item.LoginData.TOTP != nil {
			m.inputs[4].SetValue(*item.LoginData.TOTP)
		}
	}
	if item.Notes != nil {
		m.inputs[5].SetValue(item.Notes.Notes)
	}
	return m
}

func (m formLoginModel) toPayload(userID int64) models.DecipheredPayload {
	uri := m.inputs[3].Value()
	totpValue := m.inputs[4].Value()
	var totp *string
	if totpValue != "" {
		totp = &totpValue
	}

	payload := models.DecipheredPayload{
		ClientSideID: m.clientID,
		UserID:       userID,
		Metadata:     models.Metadata{Name: m.inputs[0].Value()},
		Type:         models.LoginPassword,
		LoginData: &models.LoginData{
			Username: m.inputs[1].Value(),
			Password: m.inputs[2].Value(),
			URIs:     []models.LoginURI{},
			TOTP:     totp,
		},
		Notes: &models.Notes{Notes: m.inputs[5].Value()},
	}

	if uri != "" {
		payload.LoginData.URIs = []models.LoginURI{{URI: uri, Match: 0}}
	}

	return payload
}

func (m formLoginModel) View() string {
	title := "Новый логин"
	if m.editing {
		title = "Редактирование: " + m.inputs[0].Value()
	}

	out := title + "\n\n"
	out += "Название: [" + m.inputs[0].View() + "]\n"
	out += "Логин:    [" + m.inputs[1].View() + "]\n"
	out += "Пароль:   [" + m.inputs[2].View() + "]\n"
	out += "URI:      [" + m.inputs[3].View() + "]\n"
	out += "TOTP:     [" + m.inputs[4].View() + "]\n"
	out += "Заметки:  [" + m.inputs[5].View() + "]\n\n"
	out += "esc отмена  tab следующее поле  enter сохранить"
	return out
}
