package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type mainLoopModel struct {
	ctx      context.Context
	services *service.ClientServices
	userID   int64

	items   []models.DecipheredPayload
	idx     int
	loading bool
	syncing bool
	status  string
	errMsg  string
	detail  bool
	editing bool

	editInputs     []textinput.Model
	editFocus      int
	editSubmitting bool
	editPayload    models.DecipheredPayload

	logout bool
}

type listLoadedMsg struct {
	items []models.DecipheredPayload
	err   error
}

type syncDoneMsg struct {
	err error
}

type deleteDoneMsg struct {
	err error
}

type updateDoneMsg struct {
	err error
}

func newMainLoopModel(ctx context.Context, services *service.ClientServices, userID int64) mainLoopModel {
	return mainLoopModel{
		ctx:      ctx,
		services: services,
		userID:   userID,
		loading:  true,
	}
}

func (m mainLoopModel) Init() tea.Cmd {
	return m.cmdLoadItems()
}

func (m mainLoopModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case listLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.errMsg = ""
		m.items = msg.items
		if m.idx >= len(m.items) {
			m.idx = len(m.items) - 1
		}
		if m.idx < 0 {
			m.idx = 0
		}
		return m, nil
	case syncDoneMsg:
		m.syncing = false
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Ошибка синхронизации: %v", msg.err)
			return m, nil
		}
		m.status = "Синхронизация завершена"
		m.errMsg = ""
		m.loading = true
		return m, m.cmdLoadItems()
	case deleteDoneMsg:
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Ошибка удаления: %v", msg.err)
			return m, nil
		}
		m.status = "Запись удалена"
		m.errMsg = ""
		m.loading = true
		return m, m.cmdLoadItems()
	case updateDoneMsg:
		m.editSubmitting = false
		if msg.err != nil {
			m.errMsg = fmt.Sprintf("Ошибка изменения: %v", msg.err)
			return m, nil
		}
		m.editing = false
		m.status = "Запись обновлена"
		m.errMsg = ""
		m.loading = true
		return m, m.cmdLoadItems()
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}

	if m.editing {
		return m.updateEditing(msg)
	}

	if m.detail {
		switch keyMsg.String() {
		case "esc":
			m.detail = false
		}
		return m, nil
	}

	switch keyMsg.String() {
	case "up":
		if m.idx > 0 {
			m.idx--
		}
	case "down":
		if m.idx < len(m.items)-1 {
			m.idx++
		}
	case "s":
		if m.syncing {
			return m, nil
		}
		m.syncing = true
		m.status = "Синхронизация..."
		m.errMsg = ""
		return m, m.cmdSync()
	case "enter":
		if _, ok := m.current(); !ok {
			m.status = "Нет записей"
			return m, nil
		}
		m.detail = true
	case "e":
		item, ok := m.current()
		if !ok {
			m.status = "Нет записей"
			return m, nil
		}
		m.startEdit(item)
		return m, nil
	case "ctrl+d":
		item, ok := m.current()
		if !ok {
			m.status = "Нет записей"
			return m, nil
		}
		return m, m.cmdDelete(item.ClientSideID)
	case "l":
		m.logout = true
		return m, tea.Quit
	}

	return m, nil
}

func (m mainLoopModel) View() string {
	if m.editing {
		out := "Поле      │ Значение\n"
		out += "──────────┼──────────────────────────────────────────\n"
		out += "Название  │ [" + m.editInputs[0].View() + "]\n"
		out += "Папка     │ [" + m.editInputs[1].View() + "]\n"
		if m.editSubmitting {
			out += "Действие  │ [Сохранение...]\n"
		} else {
			out += "Действие  │ [Сохранить]\n"
		}
		if m.errMsg != "" {
			out += "Ошибка    │ " + m.errMsg + "\n"
		}
		return renderPage("ИЗМЕНЕНИЕ ЗАПИСИ", strings.TrimRight(out, "\n"), "esc: назад │ tab: след. поле │ enter: сохранить")
	}

	if m.detail {
		item, ok := m.current()
		if !ok {
			return renderPage("ПРОСМОТР ЗАПИСИ", "Запись не найдена", "esc: назад")
		}

		out := "Поле       │ Значение\n"
		out += "───────────┼─────────────────────────────────────────\n"
		out += "Название   │ " + item.Metadata.Name + "\n"
		out += "Тип        │ " + dataTypeLabel(item.Type) + "\n"
		out += "Папка      │ " + valueOrDash(item.Metadata.Folder) + "\n"
		return renderPage("ПРОСМОТР ЗАПИСИ", strings.TrimRight(out, "\n"), "esc: назад")
	}

	out := ""

	if m.loading {
		out += "Загрузка списка...\n"
		return renderPage("ГЛАВНАЯ СТРАНИЦА", strings.TrimRight(out, "\n"), "s: синхр. │ enter: открыть │ e: изм. │ ctrl+d: уд. │ ↑/↓: нав.")
	}

	if m.errMsg != "" {
		out += "Ошибка: " + m.errMsg + "\n"
	}

	if m.status != "" {
		out += "Статус: " + m.status + "\n"
	}

	if len(m.items) == 0 {
		if out != "" {
			out += "\n"
		}
		out += "Записей нет\n"
	} else {
		if out != "" {
			out += "\n"
		}
		out += "ID   │ Наименование             │ Тип             │ Папка\n"
		out += "─────┼──────────────────────────┼─────────────────┼────────────────\n"
		for i, item := range m.items {
			cursor := "  "
			if i == m.idx {
				cursor = "> "
			}

			out += fmt.Sprintf(
				"%s%-3d │ %-24s │ %-15s │ %s\n",
				cursor,
				i+1,
				fitText(item.Metadata.Name, 24),
				fitText(dataTypeLabel(item.Type), 15),
				valueOrDash(item.Metadata.Folder),
			)
		}
	}

	return renderPage(
		"ГЛАВНАЯ СТРАНИЦА",
		strings.TrimRight(out, "\n"),
		"s: синхр. │ enter: открыть │ e: изм. │ ctrl+d: уд. │ ↑/↓: нав.",
	)
}

func (m mainLoopModel) current() (models.DecipheredPayload, bool) {
	if len(m.items) == 0 || m.idx < 0 || m.idx >= len(m.items) {
		return models.DecipheredPayload{}, false
	}
	return m.items[m.idx], true
}

func (m mainLoopModel) cmdLoadItems() tea.Cmd {
	ctx := m.ctx
	svc := m.services.PrivateDataService
	userID := m.userID

	return func() tea.Msg {
		items, err := svc.GetAll(ctx, userID)
		return listLoadedMsg{items: items, err: err}
	}
}

func (m mainLoopModel) cmdSync() tea.Cmd {
	ctx := m.ctx
	svc := m.services.SyncService
	userID := m.userID

	return func() tea.Msg {
		err := svc.FullSync(ctx, userID)
		return syncDoneMsg{err: err}
	}
}

func (m mainLoopModel) cmdDelete(clientSideID string) tea.Cmd {
	ctx := m.ctx
	svc := m.services.PrivateDataService
	userID := m.userID

	return func() tea.Msg {
		err := svc.Delete(ctx, clientSideID, userID)
		return deleteDoneMsg{err: err}
	}
}

func (m mainLoopModel) cmdUpdate(payload models.DecipheredPayload) tea.Cmd {
	ctx := m.ctx
	svc := m.services.PrivateDataService

	return func() tea.Msg {
		err := svc.Update(ctx, payload)
		return updateDoneMsg{err: err}
	}
}

func (m *mainLoopModel) startEdit(item models.DecipheredPayload) {
	name := textinput.New()
	name.Placeholder = "name"
	name.SetValue(item.Metadata.Name)
	name.Width = 40
	name.Focus()

	folder := textinput.New()
	folder.Placeholder = "folder"
	if item.Metadata.Folder != nil {
		folder.SetValue(*item.Metadata.Folder)
	}
	folder.Width = 40

	m.editInputs = []textinput.Model{name, folder}
	m.editFocus = 0
	m.editSubmitting = false
	m.editPayload = item
	m.editing = true
	m.errMsg = ""
}

func (m mainLoopModel) updateEditing(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch keyMsg.String() {
		case "esc":
			m.editing = false
			m.editSubmitting = false
			m.errMsg = ""
			return m, nil
		case "tab":
			m.editInputs[m.editFocus].Blur()
			m.editFocus = (m.editFocus + 1) % len(m.editInputs)
			m.editInputs[m.editFocus].Focus()
			return m, nil
		case "shift+tab":
			m.editInputs[m.editFocus].Blur()
			m.editFocus = (m.editFocus - 1 + len(m.editInputs)) % len(m.editInputs)
			m.editInputs[m.editFocus].Focus()
			return m, nil
		case "enter":
			if m.editSubmitting {
				return m, nil
			}

			name := strings.TrimSpace(m.editInputs[0].Value())
			folder := strings.TrimSpace(m.editInputs[1].Value())
			if name == "" {
				m.errMsg = "Название обязательно"
				return m, nil
			}

			payload := m.editPayload
			payload.Metadata.Name = name
			if folder == "" {
				payload.Metadata.Folder = nil
			} else {
				f := folder
				payload.Metadata.Folder = &f
			}

			m.errMsg = ""
			m.editSubmitting = true
			return m, m.cmdUpdate(payload)
		}
	}

	var cmd tea.Cmd
	m.editInputs[m.editFocus], cmd = m.editInputs[m.editFocus].Update(msg)
	return m, cmd
}

func dataTypeLabel(t models.DataType) string {
	switch t {
	case models.LoginPassword:
		return "Логин/пароль"
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
