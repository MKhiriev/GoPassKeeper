package tui

import (
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/textinput"
)

type formCardModel struct {
	inputs     []textinput.Model
	focus      int
	editing    bool
	clientID   string
	submitting bool
}

func newFormCardModel(item *models.DecipheredPayload) formCardModel {
	inputs := make([]textinput.Model, 8)
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Width = 40
	}
	inputs[6].EchoMode = textinput.EchoPassword
	inputs[6].EchoCharacter = '*'
	inputs[0].Focus()

	m := formCardModel{inputs: inputs}
	if item == nil {
		return m
	}

	m.editing = true
	m.clientID = item.ClientSideID
	m.inputs[0].SetValue(item.Metadata.Name)
	if item.BankCardData != nil {
		m.inputs[1].SetValue(item.BankCardData.CardholderName)
		m.inputs[2].SetValue(item.BankCardData.Number)
		m.inputs[3].SetValue(item.BankCardData.Brand)
		m.inputs[4].SetValue(item.BankCardData.ExpMonth)
		m.inputs[5].SetValue(item.BankCardData.ExpYear)
		m.inputs[6].SetValue(item.BankCardData.Code)
	}
	if item.Notes != nil {
		m.inputs[7].SetValue(item.Notes.Notes)
	}
	return m
}

func (m formCardModel) toPayload(userID int64) models.DecipheredPayload {
	return models.DecipheredPayload{
		ClientSideID: m.clientID,
		UserID:       userID,
		Metadata:     models.Metadata{Name: m.inputs[0].Value()},
		Type:         models.BankCard,
		BankCardData: &models.BankCardData{
			CardholderName: m.inputs[1].Value(),
			Number:         m.inputs[2].Value(),
			Brand:          m.inputs[3].Value(),
			ExpMonth:       m.inputs[4].Value(),
			ExpYear:        m.inputs[5].Value(),
			Code:           m.inputs[6].Value(),
		},
		Notes: &models.Notes{Notes: m.inputs[7].Value()},
	}
}

func (m formCardModel) View() string {
	title := "Новая карта"
	if m.editing {
		title = "Редактирование: " + m.inputs[0].Value()
	}

	out := title + "\n\n"
	out += "Название: [" + m.inputs[0].View() + "]\n"
	out += "Владелец: [" + m.inputs[1].View() + "]\n"
	out += "Номер:    [" + m.inputs[2].View() + "]\n"
	out += "Бренд:    [" + m.inputs[3].View() + "]\n"
	out += "Месяц:    [" + m.inputs[4].View() + "]\n"
	out += "Год:      [" + m.inputs[5].View() + "]\n"
	out += "CVV:      [" + m.inputs[6].View() + "]\n"
	out += "Заметки:  [" + m.inputs[7].View() + "]\n\n"
	out += "esc отмена  tab следующее поле  enter сохранить"
	return out
}
