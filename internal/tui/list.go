package tui

import (
	"fmt"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/spinner"
)

type listModel struct {
	items   []models.DecipheredPayload
	idx     int
	loading bool
	syncing bool
	spinner spinner.Model
	status  string
	userID  int64
	lastErr error
}

func newListModel() listModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return listModel{spinner: s, loading: true}
}

func (m listModel) current() (models.DecipheredPayload, bool) {
	if len(m.items) == 0 || m.idx < 0 || m.idx >= len(m.items) {
		return models.DecipheredPayload{}, false
	}
	return m.items[m.idx], true
}

func listIcon(t models.DataType) string {
	switch t {
	case models.LoginPassword:
		return "[P]"
	case models.Text:
		return "[T]"
	case models.Binary:
		return "[B]"
	case models.BankCard:
		return "[C]"
	default:
		return "[?]"
	}
}

func (m listModel) View() string {
	header := "GoPassKeeper"
	if m.syncing {
		header += "  " + m.spinner.View()
	}
	out := header + "\n\n"

	if m.loading {
		out += "Загрузка...\n"
	} else if len(m.items) == 0 {
		out += "Нет записей\n"
	} else {
		for i, item := range m.items {
			cursor := "  "
			if i == m.idx {
				cursor = "> "
			}
			out += fmt.Sprintf("%s%s %s\n", cursor, listIcon(item.Type), item.Metadata.Name)
		}
	}

	if m.status != "" {
		out += "\n" + m.status + "\n"
	}
	if m.lastErr != nil {
		out += "\nОшибка: " + m.lastErr.Error() + "\n"
	}

	out += "\n" + strings.TrimSpace("n новая  s синхр.  l перелогин  q выход  enter открыть")
	return out
}
