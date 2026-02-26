package tui

import (
	"context"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type LoginModel struct {
	ctx  context.Context
	auth service.ClientAuthService

	inputs     []textinput.Model
	focus      int
	submitting bool
	errMsg     string
}

func NewLoginModel(ctx context.Context, auth service.ClientAuthService) *LoginModel {
	loginInput := textinput.New()
	loginInput.Placeholder = "login"
	loginInput.CharLimit = 256
	loginInput.Width = 40
	loginInput.Focus()

	passwordInput := textinput.New()
	passwordInput.Placeholder = "password"
	passwordInput.CharLimit = 256
	passwordInput.Width = 40
	passwordInput.EchoMode = textinput.EchoPassword
	passwordInput.EchoCharacter = '*'

	return &LoginModel{
		ctx:    ctx,
		auth:   auth,
		inputs: []textinput.Model{loginInput, passwordInput},
	}
}

func (m *LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if result, ok := msg.(LoginResult); ok {
		m.submitting = false
		if result.Err != nil {
			m.errMsg = result.Err.Error()
		}
		return m, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch keyMsg.String() {
		case "esc":
			m.submitting = false
			m.errMsg = ""
			return m, func() tea.Msg { return NavigateTo{Page: "menu"} }
		case "tab":
			m.focusNext()
			return m, nil
		case "shift+tab":
			m.focusPrev()
			return m, nil
		case "enter":
			if m.submitting {
				return m, nil
			}

			login := strings.TrimSpace(m.inputs[0].Value())
			pass := m.inputs[1].Value()
			if login == "" || pass == "" {
				m.errMsg = "Логин и пароль обязательны"
				return m, nil
			}

			m.errMsg = ""
			m.submitting = true
			return m, m.cmdLogin(login, pass)
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
	return m, cmd
}

func (m *LoginModel) View() string {
	var b strings.Builder
	b.WriteString("Поле    │ Значение\n")
	b.WriteString("────────┼────────────────────────────────────────────\n")
	b.WriteString("Логин   │ [")
	b.WriteString(m.inputs[0].View())
	b.WriteString("]\n")
	b.WriteString("Пароль  │ [")
	b.WriteString(m.inputs[1].View())
	b.WriteString("]\n")

	if m.submitting {
		b.WriteString("\n[Войти...]\n")
	} else {
		b.WriteString("\n[Войти]\n")
	}

	if m.errMsg != "" {
		b.WriteString("\nОшибка: ")
		b.WriteString(m.errMsg)
		b.WriteString("\n")
	}

	return renderPage("ВХОД", strings.TrimRight(b.String(), "\n"), "esc: назад │ tab: след. поле │ enter: подтвердить")
}

func (m *LoginModel) cmdLogin(login, pass string) tea.Cmd {
	ctx := m.ctx
	auth := m.auth

	return func() tea.Msg {
		userID, key, err := auth.Login(ctx, models.User{
			Login:          login,
			MasterPassword: pass,
		})

		return LoginResult{
			Err:           err,
			Username:      login,
			UserID:        userID,
			EncryptionKey: key,
		}
	}
}

func (m *LoginModel) focusNext() {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + 1) % len(m.inputs)
	m.inputs[m.focus].Focus()
}

func (m *LoginModel) focusPrev() {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
}
