package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	idColWidth := lipgloss.Width("ID")
	itemsCountWidth := lipgloss.Width(fmt.Sprintf("%d", len(m.items)))
	if itemsCountWidth > idColWidth {
		idColWidth = itemsCountWidth
	}
	idColWidth += 2 // reserve space for selection marker and space ("<marker> <id>")

	actionColWidth := lipgloss.Width("Действие")
	for _, item := range m.items {
		if w := lipgloss.Width(item); w > actionColWidth {
			actionColWidth = w
		}
	}

	if m.status != "" {
		b.WriteString("OK: ")
		b.WriteString(m.status)
		b.WriteString("\n\n")
	}

	b.WriteString(fmt.Sprintf("%-*s │ %-*s\n", idColWidth, "ID", actionColWidth, "Действие"))
	b.WriteString(strings.Repeat("─", idColWidth))
	b.WriteString("─┼─")
	b.WriteString(strings.Repeat("─", actionColWidth))
	b.WriteString("\n")

	for i, item := range m.items {
		cursor := " "
		if i == m.idx {
			cursor = ">"
		}
		idCell := fmt.Sprintf("%s %d", cursor, i+1)
		b.WriteString(fmt.Sprintf("%-*s │ %-*s\n", idColWidth, idCell, actionColWidth, item))
	}

	return renderPage("ГЛАВНОЕ МЕНЮ", strings.TrimRight(b.String(), "\n"), "enter: выбрать │ ↑/↓: навигация │ v: версия")
}
