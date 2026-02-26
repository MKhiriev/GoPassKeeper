package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type MenuModel struct {
	items     []string
	idx       int
	confirmed bool
	status    string
}

func NewMenuModel() *MenuModel {
	return &MenuModel{
		items: []string{"Войти", "Зарегистрироваться"},
	}
}

func (m *MenuModel) Init() tea.Cmd {
	return nil
}

func (m *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if notice, ok := msg.(RegisterSuccessNotice); ok {
		if notice.Username != "" {
			m.status = "Пользователь " + notice.Username + " успешно зарегистрирован"
		} else {
			m.status = "Регистрация прошла успешно"
		}
		return m, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k":
		if m.idx > 0 {
			m.idx--
		}
		m.confirmed = false
	case "down", "j":
		if m.idx < len(m.items)-1 {
			m.idx++
		}
		m.confirmed = false
	case "enter":
		m.confirmed = true
		if m.idx == 0 {
			return m, func() tea.Msg { return NavigateTo{Page: "login"} }
		}
		return m, func() tea.Msg { return NavigateTo{Page: "register"} }
	}

	return m, nil
}

func (m *MenuModel) View() string {
	var b strings.Builder
	b.WriteString("ID   │ Действие\n")
	b.WriteString("─────┼────────────────────\n")

	if m.status != "" {
		b.WriteString("     │ ")
		b.WriteString("OK: ")
		b.WriteString(m.status)
		b.WriteString("\n")
	}

	for i, item := range m.items {
		cursor := "  "
		if i == m.idx {
			cursor = "> "
		}
		b.WriteString(cursor)
		b.WriteString(fmt.Sprintf("%-3d │ ", i+1))
		b.WriteString(item)
		b.WriteString("\n")
	}

	return renderPage("ГЛАВНОЕ МЕНЮ", strings.TrimRight(b.String(), "\n"), "enter: выбрать │ ↑/↓: навигация")
}
