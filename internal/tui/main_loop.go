package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type addStage int

const (
	addStageNone addStage = iota
	addStageType
	addStageMeta
	addStageData
	addStageNotes
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

	addStage       addStage
	addTypeOptions []models.DataType
	addTypeIdx     int
	addErr         string
	addPayload     models.DecipheredPayload
	addMetaInputs  []textinput.Model
	addMetaFocus   int
	addDataInputs  []textinput.Model
	addDataFocus   int
	addTextArea    textarea.Model
	addNotesArea   textarea.Model
	addSaving      bool

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

type createDoneMsg struct {
	err error
}

func newMainLoopModel(ctx context.Context, services *service.ClientServices, userID int64) mainLoopModel {
	return mainLoopModel{
		ctx:      ctx,
		services: services,
		userID:   userID,
		loading:  true,
		addTypeOptions: []models.DataType{
			models.LoginPassword,
			models.Text,
			models.Binary,
			models.BankCard,
		},
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
	case createDoneMsg:
		m.addSaving = false
		if msg.err != nil {
			m.status = "Возникла ошибка"
			m.errMsg = msg.err.Error()
			m.resetAddFlow()
			return m, nil
		}
		m.status = "Запись добавлена!"
		m.errMsg = ""
		m.resetAddFlow()
		m.loading = true
		return m, m.cmdLoadItems()
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		if m.addStage != addStageNone {
			return m.updateAddFlow(msg)
		}
		if m.editing {
			return m.updateEditing(msg)
		}
		return m, nil
	}

	switch keyMsg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	}

	if m.addStage != addStageNone {
		return m.updateAddFlow(msg)
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
	case "a":
		m.startAddFlow()
		return m, nil
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

func (m mainLoopModel) updateAddFlow(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.addStage {
	case addStageType:
		return m.updateAddType(msg)
	case addStageMeta:
		return m.updateAddMeta(msg)
	case addStageData:
		return m.updateAddData(msg)
	case addStageNotes:
		return m.updateAddNotes(msg)
	default:
		return m, nil
	}
}

func (m mainLoopModel) updateAddType(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		m.resetAddFlow()
		return m, nil
	case "up":
		if m.addTypeIdx > 0 {
			m.addTypeIdx--
		}
	case "down":
		if m.addTypeIdx < len(m.addTypeOptions)-1 {
			m.addTypeIdx++
		}
	case "1", "2", "3", "4":
		m.addTypeIdx = int(keyMsg.String()[0] - '1')
		m.selectAddType()
		return m, nil
	case "enter":
		m.selectAddType()
		return m, nil
	}

	return m, nil
}

func (m *mainLoopModel) selectAddType() {
	m.addPayload = models.DecipheredPayload{UserID: m.userID, Type: m.addTypeOptions[m.addTypeIdx]}
	m.addErr = ""
	m.addStage = addStageMeta
	m.initAddMetaInputs()
}

func (m *mainLoopModel) initAddMetaInputs() {
	name := textinput.New()
	name.Placeholder = "Название"
	name.Width = 40
	name.Focus()

	folder := textinput.New()
	folder.Placeholder = "Папка (можно пусто)"
	folder.Width = 40

	m.addMetaInputs = []textinput.Model{name, folder}
	m.addMetaFocus = 0
}

func (m mainLoopModel) updateAddMeta(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch keyMsg.String() {
		case "esc":
			m.resetAddFlow()
			return m, nil
		case "tab":
			m.addMetaInputs[m.addMetaFocus].Blur()
			m.addMetaFocus = (m.addMetaFocus + 1) % len(m.addMetaInputs)
			m.addMetaInputs[m.addMetaFocus].Focus()
			return m, nil
		case "shift+tab":
			m.addMetaInputs[m.addMetaFocus].Blur()
			m.addMetaFocus = (m.addMetaFocus - 1 + len(m.addMetaInputs)) % len(m.addMetaInputs)
			m.addMetaInputs[m.addMetaFocus].Focus()
			return m, nil
		case "enter":
			name := strings.TrimSpace(m.addMetaInputs[0].Value())
			folder := strings.TrimSpace(m.addMetaInputs[1].Value())
			if name == "" {
				m.addErr = "нужно название."
				return m, nil
			}

			m.addPayload.Metadata.Name = name
			if folder == "" {
				m.addPayload.Metadata.Folder = nil
			} else {
				f := folder
				m.addPayload.Metadata.Folder = &f
			}

			m.addErr = ""
			m.addStage = addStageData
			m.initAddDataInputs()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.addMetaInputs[m.addMetaFocus], cmd = m.addMetaInputs[m.addMetaFocus].Update(msg)
	return m, cmd
}

func (m *mainLoopModel) initAddDataInputs() {
	m.addDataInputs = nil
	m.addDataFocus = 0

	switch m.addPayload.Type {
	case models.LoginPassword:
		login := textinput.New()
		login.Placeholder = "Логин"
		login.Width = 40
		login.Focus()

		pass := textinput.New()
		pass.Placeholder = "Пароль"
		pass.Width = 40
		pass.EchoMode = textinput.EchoPassword
		pass.EchoCharacter = '*'

		uri := textinput.New()
		uri.Placeholder = "URI"
		uri.Width = 40

		totp := textinput.New()
		totp.Placeholder = "TOTP (необязательно)"
		totp.Width = 40

		m.addDataInputs = []textinput.Model{login, pass, uri, totp}

	case models.Text:
		ta := textarea.New()
		ta.Placeholder = "Введите текст"
		ta.SetWidth(54)
		ta.SetHeight(6)
		ta.Focus()
		m.addTextArea = ta

	case models.Binary:
		path := textinput.New()
		path.Placeholder = "/path/to/file"
		path.Width = 54
		path.Focus()
		m.addDataInputs = []textinput.Model{path}

	case models.BankCard:
		holder := textinput.New()
		holder.Placeholder = "Держатель"
		holder.Width = 40
		holder.Focus()

		number := textinput.New()
		number.Placeholder = "Номер"
		number.Width = 40

		brand := textinput.New()
		brand.Placeholder = "Сеть"
		brand.Width = 40

		month := textinput.New()
		month.Placeholder = "Месяц (мм)"
		month.Width = 40

		year := textinput.New()
		year.Placeholder = "Год (гг)"
		year.Width = 40

		cvv := textinput.New()
		cvv.Placeholder = "CVV"
		cvv.Width = 40
		cvv.EchoMode = textinput.EchoPassword
		cvv.EchoCharacter = '*'

		m.addDataInputs = []textinput.Model{holder, number, brand, month, year, cvv}
	}
}

func (m mainLoopModel) updateAddData(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.addPayload.Type {
	case models.Text:
		return m.updateAddDataText(msg)
	default:
		return m.updateAddDataInputs(msg)
	}
}

func (m mainLoopModel) updateAddDataText(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch keyMsg.String() {
		case "esc":
			m.resetAddFlow()
			return m, nil
		case "ctrl+s":
			text := strings.TrimSpace(m.addTextArea.Value())
			if text == "" {
				m.addErr = "нужно заполнить текст"
				return m, nil
			}
			m.addPayload.TextData = &models.TextData{Text: text}
			m.addErr = ""
			m.startAddNotes()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.addTextArea, cmd = m.addTextArea.Update(msg)
	return m, cmd
}

func (m mainLoopModel) updateAddDataInputs(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch keyMsg.String() {
		case "esc":
			m.resetAddFlow()
			return m, nil
		case "tab":
			m.addDataInputs[m.addDataFocus].Blur()
			m.addDataFocus = (m.addDataFocus + 1) % len(m.addDataInputs)
			m.addDataInputs[m.addDataFocus].Focus()
			return m, nil
		case "shift+tab":
			m.addDataInputs[m.addDataFocus].Blur()
			m.addDataFocus = (m.addDataFocus - 1 + len(m.addDataInputs)) % len(m.addDataInputs)
			m.addDataInputs[m.addDataFocus].Focus()
			return m, nil
		case "enter":
			if err := m.collectAddTypedData(); err != nil {
				m.addErr = err.Error()
				return m, nil
			}
			m.addErr = ""
			m.startAddNotes()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.addDataInputs[m.addDataFocus], cmd = m.addDataInputs[m.addDataFocus].Update(msg)
	return m, cmd
}

func (m *mainLoopModel) collectAddTypedData() error {
	switch m.addPayload.Type {
	case models.LoginPassword:
		login := strings.TrimSpace(m.addDataInputs[0].Value())
		pass := strings.TrimSpace(m.addDataInputs[1].Value())
		uri := strings.TrimSpace(m.addDataInputs[2].Value())
		totpRaw := strings.TrimSpace(m.addDataInputs[3].Value())

		if login == "" || pass == "" {
			return fmt.Errorf("логин и пароль обязательны")
		}

		data := &models.LoginData{Username: login, Password: pass}
		if uri != "" {
			data.URIs = []models.LoginURI{{URI: uri, Match: 0}}
		}
		if totpRaw != "" {
			totp := totpRaw
			data.TOTP = &totp
		}
		m.addPayload.LoginData = data
		return nil

	case models.Binary:
		path := strings.TrimSpace(m.addDataInputs[0].Value())
		if path == "" {
			return fmt.Errorf("нужно указать путь к файлу")
		}

		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("файл не найден")
		}
		if info.IsDir() {
			return fmt.Errorf("укажите путь к файлу, а не к папке")
		}

		m.addPayload.BinaryData = &models.BinaryData{
			ID:       fmt.Sprintf("bin-%d", time.Now().UnixNano()),
			FileName: filepath.Base(path),
			Size:     info.Size(),
			Key:      "",
		}
		return nil

	case models.BankCard:
		holder := strings.TrimSpace(m.addDataInputs[0].Value())
		number := strings.TrimSpace(m.addDataInputs[1].Value())
		brand := strings.TrimSpace(m.addDataInputs[2].Value())
		month := strings.TrimSpace(m.addDataInputs[3].Value())
		year := strings.TrimSpace(m.addDataInputs[4].Value())
		cvv := strings.TrimSpace(m.addDataInputs[5].Value())

		if number == "" || cvv == "" {
			return fmt.Errorf("номер карты и CVV обязательны")
		}

		m.addPayload.BankCardData = &models.BankCardData{
			CardholderName: holder,
			Number:         number,
			Brand:          brand,
			ExpMonth:       month,
			ExpYear:        year,
			Code:           cvv,
		}
		return nil
	}

	return nil
}

func (m *mainLoopModel) startAddNotes() {
	ta := textarea.New()
	ta.Placeholder = "Введите заметки (опционально)"
	ta.SetWidth(54)
	ta.SetHeight(4)
	ta.Focus()

	m.addNotesArea = ta
	m.addStage = addStageNotes
}

func (m mainLoopModel) updateAddNotes(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch keyMsg.String() {
		case "esc":
			m.resetAddFlow()
			return m, nil
		case "ctrl+s":
			if m.addSaving {
				return m, nil
			}

			notesText := strings.TrimSpace(m.addNotesArea.Value())
			payload := m.addPayload
			if notesText != "" {
				payload.Notes = &models.Notes{Notes: notesText}
			}

			m.addErr = ""
			m.addSaving = true
			return m, m.cmdCreate(payload)
		}
	}

	var cmd tea.Cmd
	m.addNotesArea, cmd = m.addNotesArea.Update(msg)
	return m, cmd
}

func (m *mainLoopModel) startAddFlow() {
	m.addStage = addStageType
	m.addTypeIdx = 0
	m.addErr = ""
	m.addSaving = false
	m.addPayload = models.DecipheredPayload{}
	m.addMetaInputs = nil
	m.addDataInputs = nil
	m.addMetaFocus = 0
	m.addDataFocus = 0
}

func (m *mainLoopModel) resetAddFlow() {
	m.addStage = addStageNone
	m.addErr = ""
	m.addSaving = false
	m.addPayload = models.DecipheredPayload{}
	m.addMetaInputs = nil
	m.addDataInputs = nil
	m.addMetaFocus = 0
	m.addDataFocus = 0
}

func (m mainLoopModel) View() string {
	switch m.addStage {
	case addStageType:
		return m.viewAddType()
	case addStageMeta:
		return m.viewAddMeta()
	case addStageData:
		return m.viewAddData()
	case addStageNotes:
		return m.viewAddNotes()
	}

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
		return renderPage("ГЛАВНАЯ СТРАНИЦА", strings.TrimRight(out, "\n"), "a: добавить │ s: синхр. │ enter: открыть │ e: изм. │ ctrl+d: уд. │ ↑/↓: нав.")
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
			cursor := " "
			if i == m.idx {
				cursor = ">"
			}

			out += fmt.Sprintf(
				"%s %-3d│ %-24s │ %-15s │ %s\n",
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
		"a: добавить │ s: синхр. │ enter: открыть │ e: изм. │ ctrl+d: уд. │ ↑/↓: нав.",
	)
}

func (m mainLoopModel) viewAddType() string {
	out := ""
	for i, t := range m.addTypeOptions {
		cursor := " "
		if i == m.addTypeIdx {
			cursor = ">"
		}
		out += fmt.Sprintf("%s %d. %s\n", cursor, i+1, dataTypeLabel(t))
	}
	if m.addErr != "" {
		out += "\nОшибка: " + m.addErr + "\n"
	}

	return renderPage("ДОБАВИТЬ: ВЫБОР ТИПА", strings.TrimRight(out, "\n"), "1-4/enter: выбрать │ ↑/↓: навигация │ esc: отмена")
}

func (m mainLoopModel) viewAddMeta() string {
	out := "[ ОСНОВНОЕ ]\n"
	out += "Название  : [ " + m.addMetaInputs[0].View() + " ]\n"
	out += "Папка     : [ " + m.addMetaInputs[1].View() + " ]\n"
	if m.addErr != "" {
		out += "\nОшибка: " + m.addErr + "\n"
	}

	return renderPage("ДОБАВИТЬ: МЕТАДАННЫЕ", strings.TrimRight(out, "\n"), "tab: след. поле │ shift+tab: пред. поле │ enter: далее │ esc: отмена")
}

func (m mainLoopModel) viewAddData() string {
	meta := "[ ОСНОВНОЕ ]\n"
	meta += "Название  : " + m.addPayload.Metadata.Name + "\n"
	meta += "Папка     : " + valueOrDash(m.addPayload.Metadata.Folder) + "\n\n"

	switch m.addPayload.Type {
	case models.LoginPassword:
		out := meta
		out += "Логин     : [ " + m.addDataInputs[0].View() + " ]\n"
		out += "Пароль    : [ " + m.addDataInputs[1].View() + " ]\n"
		out += "URI       : [ " + m.addDataInputs[2].View() + " ]\n"
		out += "TOTP      : [ " + m.addDataInputs[3].View() + " ]\n"
		if m.addErr != "" {
			out += "\nОшибка: " + m.addErr + "\n"
		}
		return renderPage("НОВАЯ ЗАПИСЬ: Логин/Пароль", strings.TrimRight(out, "\n"), "tab: след. поле │ shift+tab: пред. поле │ enter: сохранить │ esc: отмена")

	case models.Text:
		out := meta
		out += "Текст:\n"
		out += m.addTextArea.View()
		if m.addErr != "" {
			out += "\nОшибка: " + m.addErr + "\n"
		}
		return renderPage("НОВАЯ ЗАПИСЬ: Текстовые данные", strings.TrimRight(out, "\n"), "enter: новая строка │ ctrl+s: сохранить │ esc: отмена")

	case models.Binary:
		out := meta
		path := strings.TrimSpace(m.addDataInputs[0].Value())
		out += "Путь      : [ " + m.addDataInputs[0].View() + " ]\n\n"
		out += "Файл      : " + binaryPreview(path) + "\n"
		if m.addErr != "" {
			out += "\nОшибка: " + m.addErr + "\n"
		}
		return renderPage("НОВАЯ ЗАПИСЬ: Файл", strings.TrimRight(out, "\n"), "tab: след. поле │ enter: сохранить │ esc: отмена")

	case models.BankCard:
		out := meta
		out += "Держатель : [ " + m.addDataInputs[0].View() + " ]\n"
		out += "Номер     : [ " + m.addDataInputs[1].View() + " ]\n"
		out += "Сеть      : [ " + m.addDataInputs[2].View() + " ]\n"
		out += "Срок (мм) : [ " + m.addDataInputs[3].View() + " ]\n"
		out += "Срок (гг) : [ " + m.addDataInputs[4].View() + " ]\n"
		out += "CVV       : [ " + m.addDataInputs[5].View() + " ]\n"
		if m.addErr != "" {
			out += "\nОшибка: " + m.addErr + "\n"
		}
		return renderPage("НОВАЯ ЗАПИСЬ: Банковская карта", strings.TrimRight(out, "\n"), "tab: след. поле │ shift+tab: пред. поле │ enter: сохранить │ esc: отмена")
	}

	return renderPage("НОВАЯ ЗАПИСЬ", "Неизвестный тип", "esc: отмена")
}

func (m mainLoopModel) viewAddNotes() string {
	out := "[ ЗАМЕТКИ ]\n"
	out += m.addNotesArea.View()
	if m.addErr != "" {
		out += "\nОшибка: " + m.addErr + "\n"
	}
	if m.addSaving {
		out += "\nСохранение...\n"
	}

	return renderPage("ЗАМЕТКИ", strings.TrimRight(out, "\n"), "enter: новая строка │ ctrl+s: сохранить │ esc: отмена")
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

func (m mainLoopModel) cmdCreate(payload models.DecipheredPayload) tea.Cmd {
	ctx := m.ctx
	svc := m.services.PrivateDataService
	userID := m.userID

	return func() tea.Msg {
		err := svc.Create(ctx, userID, payload)
		return createDoneMsg{err: err}
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
		return "Текстовые данные"
	case models.Binary:
		return "Бинарные"
	case models.BankCard:
		return "Банковская карта"
	default:
		return "Неизвестно"
	}
}

func binaryPreview(path string) string {
	if path == "" {
		return "(не выбран)"
	}

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return "не найден"
	}

	return fmt.Sprintf("%s (%s) ✓ готов к загрузке", filepath.Base(path), formatSize(info.Size()))
}

func formatSize(size int64) string {
	const mb = 1024 * 1024
	const kb = 1024

	if size >= mb {
		return fmt.Sprintf("%.1f MB", float64(size)/mb)
	}
	if size >= kb {
		return fmt.Sprintf("%.1f KB", float64(size)/kb)
	}
	return fmt.Sprintf("%d B", size)
}
