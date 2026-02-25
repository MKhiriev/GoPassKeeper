package tui

import (
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/models"
)

type detailModel struct {
	item   models.DecipheredPayload
	status string
}

func dataTypeName(t models.DataType) string {
	switch t {
	case models.LoginPassword:
		return "Логин/Пароль"
	case models.Text:
		return "Текст"
	case models.Binary:
		return "Файл"
	case models.BankCard:
		return "Банковская карта"
	default:
		return "Неизвестно"
	}
}

func safeNotes(n *models.Notes) string {
	if n == nil || n.Notes == "" {
		return "—"
	}
	return n.Notes
}

func (m detailModel) View() string {
	out := fmt.Sprintf("%s  [%s]\n\n", m.item.Metadata.Name, dataTypeName(m.item.Type))

	switch m.item.Type {
	case models.LoginPassword:
		login := m.item.LoginData
		if login != nil {
			uri := "—"
			if len(login.URIs) > 0 {
				uri = login.URIs[0].URI
			}
			totp := "—"
			if login.TOTP != nil && *login.TOTP != "" {
				totp = *login.TOTP
			}
			out += fmt.Sprintf("Логин:    %s\n", login.Username)
			out += "Пароль:   ••••••••\n"
			out += fmt.Sprintf("URI:      %s\n", uri)
			out += fmt.Sprintf("TOTP:     %s\n", totp)
			out += fmt.Sprintf("Заметки:  %s\n", safeNotes(m.item.Notes))
		}
		out += "\n"
		out += "e редакт.  d удалить  c копир. пароль  u копир. логин  esc назад"
	case models.BankCard:
		card := m.item.BankCardData
		if card != nil {
			out += fmt.Sprintf("Владелец: %s\n", card.CardholderName)
			out += fmt.Sprintf("Номер:    %s\n", card.Number)
			out += fmt.Sprintf("Бренд:    %s\n", card.Brand)
			out += fmt.Sprintf("Срок:     %s/%s\n", card.ExpMonth, card.ExpYear)
			out += fmt.Sprintf("CVV:      %s\n", card.Code)
			out += fmt.Sprintf("Заметки:  %s\n", safeNotes(m.item.Notes))
		}
		out += "\n"
		out += "e редакт.  d удалить  c копир. номер  esc назад"
	case models.Text:
		text := m.item.TextData
		if text != nil {
			out += text.Text + "\n\n"
			out += fmt.Sprintf("Заметки:  %s\n", safeNotes(m.item.Notes))
		}
		out += "\n"
		out += "e редакт.  d удалить  c копир. текст  esc назад"
	case models.Binary:
		bin := m.item.BinaryData
		if bin != nil {
			out += fmt.Sprintf("Имя файла: %s\n", bin.FileName)
			out += fmt.Sprintf("Размер:    %d bytes\n", bin.Size)
			out += fmt.Sprintf("Заметки:   %s\n", safeNotes(m.item.Notes))
		}
		out += "\n"
		out += "e редакт.  d удалить  esc назад"
	}

	if m.status != "" {
		out += "\n\n" + m.status
	}

	return out
}
