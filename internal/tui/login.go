// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package tui

import (
	"context"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// LoginModel is the Bubble Tea model for the login screen. It renders two text inputs
// (username and password) and dispatches an async login command on form submission.
// On success a [LoginResult] message is produced and handled by [RootModel] to finish
// the authentication flow.
type LoginModel struct {
	ctx  context.Context
	auth service.ClientAuthService

	inputs     []textinput.Model
	focus      int
	submitting bool
	errMsg     string
}

// NewLoginModel creates a [LoginModel] with pre-configured username and password inputs.
// The username field receives focus immediately; the password field uses masked echo.
func NewLoginModel(ctx context.Context, auth service.ClientAuthService) *LoginModel {
	loginInput := textinput.New()
	loginInput.Placeholder = "login"
	loginInput.CharLimit = 20
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

// Init implements [tea.Model]. Starts the cursor-blink animation for the active input.
func (m *LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements [tea.Model]. Handled messages:
//   - [LoginResult]  — clears submitting state; on error, populates errMsg.
//   - esc            — cancels and navigates back to the menu.
//   - tab            — moves focus to the next input.
//   - shift+tab      — moves focus to the previous input.
//   - enter          — validates inputs and dispatches the async login command.
//
// All other key events are forwarded to the focused input widget.
func (m *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if result, ok := msg.(LoginResult); ok {
		m.submitting = false
		if result.Err != nil {
			m.errMsg = humanizeServerUnavailableError(result.Err)
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

// View implements [tea.Model]. Renders the login form as a two-column table with
// username and password inputs, a submission indicator, and an optional error message.
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
