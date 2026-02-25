package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

var ErrUserQuit = errors.New("user quit")

type screen int

const (
	screenWelcome screen = iota
	screenLogin
	screenRegister
	screenSync
	screenList
	screenDetail
	screenTypeSelect
	screenFormLogin
	screenFormText
	screenFormCard
	screenFormBinary
)

type appMode int

const (
	modeLogin appMode = iota
	modeMain
)

type appModel struct {
	ctx           context.Context
	services      *service.ClientServices
	mode          appMode
	currentScreen screen

	welcome    welcomeModel
	login      loginModel
	register   registerModel
	syncScreen syncModel
	list       listModel
	detail     detailModel
	typeSelect typeSelectModel
	formLogin  formLoginModel
	formText   formTextModel
	formCard   formCardModel
	formBinary formBinaryModel

	userID        int64
	encryptionKey []byte
	err           error
	showError     bool
	errorOverlay  errorOverlayModel
	showConfirm   bool
	confirm       confirmModel
	pendingDelete string
	logout        bool
	resultUserID  int64
	resultKey     []byte
}

func newLoginAppModel(ctx context.Context, services *service.ClientServices) appModel {
	return appModel{
		ctx:           ctx,
		services:      services,
		mode:          modeLogin,
		currentScreen: screenWelcome,
		welcome:       newWelcomeModel(),
		login:         newLoginModel(),
		register:      newRegisterModel(),
		syncScreen:    newSyncModel(),
		list:          newListModel(),
		typeSelect:    newTypeSelectModel(),
	}
}

func newMainAppModel(ctx context.Context, services *service.ClientServices, userID int64) appModel {
	m := newLoginAppModel(ctx, services)
	m.mode = modeMain
	m.userID = userID
	m.currentScreen = screenList
	m.list.userID = userID
	m.list.loading = true
	return m
}

func (m appModel) Init() tea.Cmd {
	if m.mode == modeMain {
		return m.cmdLoadList()
	}
	return nil
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showError {
			if key.Matches(msg, keys.enter) || key.Matches(msg, keys.esc) {
				m.showError = false
				m.errorOverlay.message = ""
			}
			return m, nil
		}
		if m.showConfirm {
			if key.Matches(msg, keys.yes) {
				m.showConfirm = false
				if m.pendingDelete == "" {
					return m, nil
				}
				return m, m.cmdDeleteItem(m.pendingDelete)
			}
			if key.Matches(msg, keys.no) || key.Matches(msg, keys.esc) {
				m.showConfirm = false
				m.pendingDelete = ""
			}
			return m, nil
		}
	case authDoneMsg:
		m.resultUserID = msg.userID
		m.resultKey = msg.key
		return m, tea.Quit
	case listLoadedMsg:
		m.list.loading = false
		if msg.err != nil {
			m.showErrorf(msg.err.Error())
			return m, nil
		}
		m.list.items = msg.items
		if m.list.idx >= len(m.list.items) {
			m.list.idx = len(m.list.items) - 1
		}
		if m.list.idx < 0 {
			m.list.idx = 0
		}
		return m, nil
	case syncDoneMsg:
		m.list.syncing = false
		if msg.err != nil {
			m.showErrorf("Сервер недоступен, синхронизация будет выполнена позже")
		}
		return m, m.cmdLoadList()
	case itemSavedMsg:
		if msg.err != nil {
			m.showErrorf(msg.err.Error())
			m.setSubmitting(false)
			return m, nil
		}
		m.setSubmitting(false)
		m.currentScreen = screenList
		return m, m.cmdLoadList()
	case itemDeletedMsg:
		if msg.err != nil {
			m.showErrorf(msg.err.Error())
			return m, nil
		}
		m.pendingDelete = ""
		m.currentScreen = screenList
		return m, m.cmdLoadList()
	case copiedMsg:
		if m.currentScreen == screenDetail {
			m.detail.status = "Скопировано!"
		}
		m.list.status = "Скопировано!"
		return m, cmdClearStatus()
	case clearStatusMsg:
		m.detail.status = ""
		m.list.status = ""
		return m, nil
	case tea.WindowSizeMsg:
		return m, nil
	}

	switch m.currentScreen {
	case screenWelcome:
		return m.updateWelcome(msg)
	case screenLogin:
		return m.updateLogin(msg)
	case screenRegister:
		return m.updateRegister(msg)
	case screenSync:
		return m.updateSync(msg)
	case screenList:
		return m.updateList(msg)
	case screenDetail:
		return m.updateDetail(msg)
	case screenTypeSelect:
		return m.updateTypeSelect(msg)
	case screenFormLogin:
		return m.updateFormLogin(msg)
	case screenFormText:
		return m.updateFormText(msg)
	case screenFormCard:
		return m.updateFormCard(msg)
	case screenFormBinary:
		return m.updateFormBinary(msg)
	}

	return m, nil
}

func (m appModel) View() string {
	var body string
	switch m.currentScreen {
	case screenWelcome:
		body = m.welcome.View()
	case screenLogin:
		body = m.login.View()
	case screenRegister:
		body = m.register.View()
	case screenSync:
		body = m.syncScreen.View()
	case screenList:
		body = m.list.View()
	case screenDetail:
		body = m.detail.View()
	case screenTypeSelect:
		body = m.typeSelect.View()
	case screenFormLogin:
		body = m.formLogin.View()
	case screenFormText:
		body = m.formText.View()
	case screenFormCard:
		body = m.formCard.View()
	case screenFormBinary:
		body = m.formBinary.View()
	}

	if m.showConfirm {
		body += "\n\n" + m.confirm.View()
	}
	if m.showError {
		body += "\n\n" + m.errorOverlay.View()
	}

	return appStyle.Render(body)
}

func (m *appModel) showErrorf(message string) {
	m.showError = true
	m.errorOverlay.message = message
}

func (m *appModel) setSubmitting(v bool) {
	m.login.submitting = v
	m.register.submitting = v
	m.formLogin.submitting = v
	m.formText.submitting = v
	m.formCard.submitting = v
	m.formBinary.submitting = v
}

func (m appModel) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch {
	case key.Matches(keyMsg, keys.up):
		if m.welcome.idx > 0 {
			m.welcome.idx--
		}
	case key.Matches(keyMsg, keys.down):
		if m.welcome.idx < len(m.welcome.items)-1 {
			m.welcome.idx++
		}
	case key.Matches(keyMsg, keys.enter):
		if m.welcome.idx == 0 {
			m.currentScreen = screenLogin
		} else {
			m.currentScreen = screenRegister
		}
	case key.Matches(keyMsg, keys.quit):
		m.err = ErrUserQuit
		return m, tea.Quit
	}
	return m, nil
}

func (m appModel) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, keys.quit):
			m.err = ErrUserQuit
			return m, tea.Quit
		case key.Matches(keyMsg, keys.esc):
			m.currentScreen = screenWelcome
			return m, nil
		case key.Matches(keyMsg, keys.tab):
			m.login = focusNext(m.login)
			return m, nil
		case key.Matches(keyMsg, keys.backtab):
			m.login = focusPrev(m.login)
			return m, nil
		case key.Matches(keyMsg, keys.enter):
			login := strings.TrimSpace(m.login.inputs[0].Value())
			pass := m.login.inputs[1].Value()
			if login == "" || pass == "" {
				m.showErrorf("Логин и пароль обязательны")
				return m, nil
			}
			m.login.submitting = true
			return m, m.cmdLogin(models.User{Login: login, MasterPassword: pass})
		}
	}

	var cmd tea.Cmd
	m.login.inputs[m.login.focus], cmd = m.login.inputs[m.login.focus].Update(msg)
	return m, cmd
}

func (m appModel) updateRegister(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, keys.quit):
			m.err = ErrUserQuit
			return m, tea.Quit
		case key.Matches(keyMsg, keys.esc):
			m.currentScreen = screenWelcome
			return m, nil
		case key.Matches(keyMsg, keys.tab):
			m.register = focusNextRegister(m.register)
			return m, nil
		case key.Matches(keyMsg, keys.backtab):
			m.register = focusPrevRegister(m.register)
			return m, nil
		case key.Matches(keyMsg, keys.enter):
			name := strings.TrimSpace(m.register.inputs[0].Value())
			login := strings.TrimSpace(m.register.inputs[1].Value())
			pass := m.register.inputs[2].Value()
			repeat := m.register.inputs[3].Value()
			hint := m.register.inputs[4].Value()
			if name == "" || login == "" || pass == "" {
				m.showErrorf("Имя, логин и пароль обязательны")
				return m, nil
			}
			if pass != repeat {
				m.showErrorf("Пароли не совпадают")
				return m, nil
			}
			m.register.submitting = true
			return m, m.cmdRegisterAndLogin(models.User{
				Name:               name,
				Login:              login,
				MasterPassword:     pass,
				MasterPasswordHint: hint,
			})
		}
	}

	var cmd tea.Cmd
	m.register.inputs[m.register.focus], cmd = m.register.inputs[m.register.focus].Update(msg)
	return m, cmd
}

func (m appModel) updateSync(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, nil
	}

	spinnerModel, cmd := m.syncScreen.spinner.Update(msg)
	m.syncScreen.spinner = spinnerModel
	return m, cmd
}

func (m appModel) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.up):
			if m.list.idx > 0 {
				m.list.idx--
			}
		case key.Matches(msg, keys.down):
			if m.list.idx < len(m.list.items)-1 {
				m.list.idx++
			}
		case key.Matches(msg, keys.enter):
			item, ok := m.list.current()
			if !ok {
				return m, nil
			}
			m.detail.item = item
			m.currentScreen = screenDetail
		case key.Matches(msg, keys.newItem):
			m.currentScreen = screenTypeSelect
		case key.Matches(msg, keys.sync):
			if m.list.syncing {
				return m, nil
			}
			m.list.syncing = true
			return m, tea.Batch(m.list.spinner.Tick, m.cmdSync())
		case key.Matches(msg, keys.quit):
			return m, tea.Quit
		case key.Matches(msg, keys.logout):
			m.logout = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		if m.list.syncing {
			var cmd tea.Cmd
			m.list.spinner, cmd = m.list.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m appModel) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch {
	case key.Matches(keyMsg, keys.esc):
		m.currentScreen = screenList
		return m, nil
	case key.Matches(keyMsg, keys.edit):
		item := m.detail.item
		switch item.Type {
		case models.LoginPassword:
			m.formLogin = newFormLoginModel(&item)
			m.currentScreen = screenFormLogin
		case models.Text:
			m.formText = newFormTextModel(&item)
			m.currentScreen = screenFormText
		case models.BankCard:
			m.formCard = newFormCardModel(&item)
			m.currentScreen = screenFormCard
		case models.Binary:
			m.formBinary = newFormBinaryModel(&item)
			m.currentScreen = screenFormBinary
		}
		return m, nil
	case key.Matches(keyMsg, keys.delete):
		m.showConfirm = true
		m.confirm.message = m.detail.item.Metadata.Name
		m.pendingDelete = m.detail.item.ClientSideID
		return m, nil
	case key.Matches(keyMsg, keys.copy):
		text, ok := copyValue(m.detail.item, false)
		if !ok {
			return m, nil
		}
		return m, cmdCopyToClipboard(text)
	case key.Matches(keyMsg, keys.copyUser):
		text, ok := copyValue(m.detail.item, true)
		if !ok {
			return m, nil
		}
		return m, cmdCopyToClipboard(text)
	}

	return m, nil
}

func (m appModel) updateTypeSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch {
	case key.Matches(keyMsg, keys.esc):
		m.currentScreen = screenList
	case key.Matches(keyMsg, keys.up):
		if m.typeSelect.idx > 0 {
			m.typeSelect.idx--
		}
	case key.Matches(keyMsg, keys.down):
		if m.typeSelect.idx < len(m.typeSelect.items)-1 {
			m.typeSelect.idx++
		}
	case key.Matches(keyMsg, keys.enter):
		switch m.typeSelect.idx {
		case 0:
			m.formLogin = newFormLoginModel(nil)
			m.currentScreen = screenFormLogin
		case 1:
			m.formText = newFormTextModel(nil)
			m.currentScreen = screenFormText
		case 2:
			m.formCard = newFormCardModel(nil)
			m.currentScreen = screenFormCard
		case 3:
			m.formBinary = newFormBinaryModel(nil)
			m.currentScreen = screenFormBinary
		}
	}

	return m, nil
}

func (m appModel) updateFormLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, keys.esc):
			m.currentScreen = backFromForm(m.formLogin.editing)
			return m, nil
		case key.Matches(keyMsg, keys.tab):
			m.formLogin = focusNextFormLogin(m.formLogin)
			return m, nil
		case key.Matches(keyMsg, keys.backtab):
			m.formLogin = focusPrevFormLogin(m.formLogin)
			return m, nil
		case key.Matches(keyMsg, keys.enter):
			if strings.TrimSpace(m.formLogin.inputs[0].Value()) == "" || strings.TrimSpace(m.formLogin.inputs[1].Value()) == "" || m.formLogin.inputs[2].Value() == "" {
				m.showErrorf("Название, логин и пароль обязательны")
				return m, nil
			}
			m.formLogin.submitting = true
			payload := m.formLogin.toPayload(m.userID)
			return m, m.savePayload(payload, m.formLogin.editing)
		}
	}

	var cmd tea.Cmd
	m.formLogin.inputs[m.formLogin.focus], cmd = m.formLogin.inputs[m.formLogin.focus].Update(msg)
	return m, cmd
}

func (m appModel) updateFormText(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, keys.esc):
			m.currentScreen = backFromForm(m.formText.editing)
			return m, nil
		case key.Matches(keyMsg, keys.tab):
			m.formText = focusNextFormText(m.formText)
			return m, nil
		case key.Matches(keyMsg, keys.backtab):
			m.formText = focusPrevFormText(m.formText)
			return m, nil
		case key.Matches(keyMsg, keys.enter):
			if strings.TrimSpace(m.formText.inputs[0].Value()) == "" || strings.TrimSpace(m.formText.inputs[1].Value()) == "" {
				m.showErrorf("Название и текст обязательны")
				return m, nil
			}
			m.formText.submitting = true
			payload := m.formText.toPayload(m.userID)
			return m, m.savePayload(payload, m.formText.editing)
		}
	}

	var cmd tea.Cmd
	m.formText.inputs[m.formText.focus], cmd = m.formText.inputs[m.formText.focus].Update(msg)
	return m, cmd
}

func (m appModel) updateFormCard(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, keys.esc):
			m.currentScreen = backFromForm(m.formCard.editing)
			return m, nil
		case key.Matches(keyMsg, keys.tab):
			m.formCard = focusNextFormCard(m.formCard)
			return m, nil
		case key.Matches(keyMsg, keys.backtab):
			m.formCard = focusPrevFormCard(m.formCard)
			return m, nil
		case key.Matches(keyMsg, keys.enter):
			if strings.TrimSpace(m.formCard.inputs[0].Value()) == "" || strings.TrimSpace(m.formCard.inputs[2].Value()) == "" {
				m.showErrorf("Название и номер карты обязательны")
				return m, nil
			}
			m.formCard.submitting = true
			payload := m.formCard.toPayload(m.userID)
			return m, m.savePayload(payload, m.formCard.editing)
		}
	}

	var cmd tea.Cmd
	m.formCard.inputs[m.formCard.focus], cmd = m.formCard.inputs[m.formCard.focus].Update(msg)
	return m, cmd
}

func (m appModel) updateFormBinary(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch {
		case key.Matches(keyMsg, keys.esc):
			m.currentScreen = backFromForm(m.formBinary.editing)
			return m, nil
		case key.Matches(keyMsg, keys.tab):
			m.formBinary = focusNextFormBinary(m.formBinary)
			return m, nil
		case key.Matches(keyMsg, keys.backtab):
			m.formBinary = focusPrevFormBinary(m.formBinary)
			return m, nil
		case key.Matches(keyMsg, keys.enter):
			if strings.TrimSpace(m.formBinary.inputs[0].Value()) == "" || strings.TrimSpace(m.formBinary.inputs[1].Value()) == "" {
				m.showErrorf("Название и имя файла обязательны")
				return m, nil
			}
			m.formBinary.submitting = true
			payload := m.formBinary.toPayload(m.userID)
			return m, m.savePayload(payload, m.formBinary.editing)
		}
	}

	var cmd tea.Cmd
	m.formBinary.inputs[m.formBinary.focus], cmd = m.formBinary.inputs[m.formBinary.focus].Update(msg)
	return m, cmd
}

func (m appModel) savePayload(payload models.DecipheredPayload, editing bool) tea.Cmd {
	if editing {
		return m.cmdUpdateItem(payload)
	}
	return m.cmdCreateItem(payload)
}

func (m appModel) cmdLogin(user models.User) tea.Cmd {
	ctx := m.ctx
	auth := m.services.AuthService
	return func() tea.Msg {
		userID, key, err := auth.Login(ctx, user)
		if err != nil {
			return itemSavedMsg{err: err}
		}
		return authDoneMsg{userID: userID, key: key}
	}
}

func (m appModel) cmdRegisterAndLogin(user models.User) tea.Cmd {
	ctx := m.ctx
	auth := m.services.AuthService
	return func() tea.Msg {
		if err := auth.Register(ctx, user); err != nil {
			return itemSavedMsg{err: err}
		}
		userID, key, err := auth.Login(ctx, models.User{Login: user.Login, MasterPassword: user.MasterPassword})
		if err != nil {
			return itemSavedMsg{err: err}
		}
		return authDoneMsg{userID: userID, key: key}
	}
}

func (m appModel) cmdLoadList() tea.Cmd {
	ctx := m.ctx
	svc := m.services.PrivateDataService
	userID := m.userID
	return func() tea.Msg {
		items, err := svc.GetAll(ctx, userID)
		return listLoadedMsg{items: items, err: err}
	}
}

func (m appModel) cmdSync() tea.Cmd {
	ctx := m.ctx
	svc := m.services.SyncService
	userID := m.userID
	return func() tea.Msg {
		err := svc.FullSync(ctx, userID)
		return syncDoneMsg{err: err}
	}
}

func (m appModel) cmdCreateItem(payload models.DecipheredPayload) tea.Cmd {
	ctx := m.ctx
	svc := m.services.PrivateDataService
	userID := m.userID
	return func() tea.Msg {
		err := svc.Create(ctx, userID, payload)
		return itemSavedMsg{err: err}
	}
}

func (m appModel) cmdUpdateItem(payload models.DecipheredPayload) tea.Cmd {
	ctx := m.ctx
	svc := m.services.PrivateDataService
	return func() tea.Msg {
		err := svc.Update(ctx, payload)
		return itemSavedMsg{err: err}
	}
}

func (m appModel) cmdDeleteItem(clientSideID string) tea.Cmd {
	ctx := m.ctx
	svc := m.services.PrivateDataService
	userID := m.userID
	return func() tea.Msg {
		err := svc.Delete(ctx, clientSideID, userID)
		return itemDeletedMsg{err: err}
	}
}

func cmdCopyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		if err := clipboard.WriteAll(text); err != nil {
			return itemSavedMsg{err: fmt.Errorf("copy to clipboard: %w", err)}
		}
		return copiedMsg{}
	}
}

func cmdClearStatus() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func copyValue(item models.DecipheredPayload, copyUsername bool) (string, bool) {
	switch item.Type {
	case models.LoginPassword:
		if item.LoginData == nil {
			return "", false
		}
		if copyUsername {
			return item.LoginData.Username, item.LoginData.Username != ""
		}
		return item.LoginData.Password, item.LoginData.Password != ""
	case models.BankCard:
		if item.BankCardData == nil {
			return "", false
		}
		return item.BankCardData.Number, item.BankCardData.Number != ""
	case models.Text:
		if item.TextData == nil {
			return "", false
		}
		return item.TextData.Text, item.TextData.Text != ""
	default:
		return "", false
	}
}

func backFromForm(editing bool) screen {
	if editing {
		return screenDetail
	}
	return screenList
}

func focusNext(m loginModel) loginModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + 1) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusPrev(m loginModel) loginModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusNextRegister(m registerModel) registerModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + 1) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusPrevRegister(m registerModel) registerModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusNextFormLogin(m formLoginModel) formLoginModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + 1) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusPrevFormLogin(m formLoginModel) formLoginModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusNextFormText(m formTextModel) formTextModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + 1) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusPrevFormText(m formTextModel) formTextModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusNextFormCard(m formCardModel) formCardModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + 1) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusPrevFormCard(m formCardModel) formCardModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusNextFormBinary(m formBinaryModel) formBinaryModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + 1) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusPrevFormBinary(m formBinaryModel) formBinaryModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}
