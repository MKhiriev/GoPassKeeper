package tui

import (
	"context"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// RegisterModel is the Bubble Tea model for the registration screen. It renders five
// text inputs (display name, username, password, password confirmation, and password
// hint) and dispatches an async registration command on form submission.
// On success a [RegisterResult] message is produced; the model then resets the form
// and navigates back to the menu, passing a [RegisterSuccessNotice] payload.
type RegisterModel struct {
	ctx  context.Context
	auth service.ClientAuthService

	inputs     []textinput.Model
	focus      int
	submitting bool
	errMsg     string
}

// NewRegisterModel creates a [RegisterModel] with five pre-configured text inputs.
// The name field receives focus immediately; the password fields use masked echo.
func NewRegisterModel(ctx context.Context, auth service.ClientAuthService) *RegisterModel {
	fields := make([]textinput.Model, 5)

	fields[0] = textinput.New()
	fields[0].Placeholder = "name"
	fields[0].Width = 40
	fields[0].Focus()

	fields[1] = textinput.New()
	fields[1].Placeholder = "login"
	fields[1].CharLimit = 20
	fields[1].Width = 40

	fields[2] = textinput.New()
	fields[2].Placeholder = "password"
	fields[2].EchoMode = textinput.EchoPassword
	fields[2].EchoCharacter = '*'
	fields[2].Width = 40

	fields[3] = textinput.New()
	fields[3].Placeholder = "repeat password"
	fields[3].EchoMode = textinput.EchoPassword
	fields[3].EchoCharacter = '*'
	fields[3].Width = 40

	fields[4] = textinput.New()
	fields[4].Placeholder = "hint"
	fields[4].Width = 40

	return &RegisterModel{
		ctx:    ctx,
		auth:   auth,
		inputs: fields,
	}
}

// Init implements [tea.Model]. Starts the cursor-blink animation for the active input.
func (m *RegisterModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements [tea.Model]. Handled messages:
//   - [RegisterResult] — clears submitting state; on error, populates errMsg;
//     on success, resets the form and navigates to the menu.
//   - esc              — cancels and navigates back to the menu.
//   - tab              — moves focus to the next input.
//   - shift+tab        — moves focus to the previous input.
//   - enter            — validates inputs (all required; passwords must match) and
//     dispatches the async registration command.
//
// All other key events are forwarded to the focused input widget.
func (m *RegisterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if result, ok := msg.(RegisterResult); ok {
		m.submitting = false
		if result.Err != nil {
			m.errMsg = humanizeServerUnavailableError(result.Err)
			return m, nil
		}

		m.errMsg = ""
		m.resetForm()
		return m, func() tea.Msg {
			return NavigateTo{
				Page:    "menu",
				Payload: RegisterSuccessNotice{Username: result.Username},
			}
		}
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

			name := strings.TrimSpace(m.inputs[0].Value())
			login := strings.TrimSpace(m.inputs[1].Value())
			pass := strings.TrimSpace(m.inputs[2].Value())
			repeat := strings.TrimSpace(m.inputs[3].Value())
			hint := strings.TrimSpace(m.inputs[4].Value())

			if name == "" || login == "" || pass == "" || repeat == "" || hint == "" {
				m.errMsg = "Все поля обязательны"
				return m, nil
			}
			if pass != repeat {
				m.errMsg = "Пароли не совпадают"
				return m, nil
			}

			m.errMsg = ""
			m.submitting = true
			return m, m.cmdRegister(name, login, pass, hint)
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
	return m, cmd
}

// View implements [tea.Model]. Renders the registration form as a two-column table
// with all five input fields, a submission indicator, and an optional error message.
func (m *RegisterModel) View() string {
	var b strings.Builder
	b.WriteString("Поле           │ Значение\n")
	b.WriteString("───────────────┼────────────────────────────────────\n")
	b.WriteString("Имя            │ [")
	b.WriteString(m.inputs[0].View())
	b.WriteString("]\n")
	b.WriteString("Логин          │ [")
	b.WriteString(m.inputs[1].View())
	b.WriteString("]\n")
	b.WriteString("Пароль         │ [")
	b.WriteString(m.inputs[2].View())
	b.WriteString("]\n")
	b.WriteString("Повтор пароля  │ [")
	b.WriteString(m.inputs[3].View())
	b.WriteString("]\n")
	b.WriteString("Подсказка      │ [")
	b.WriteString(m.inputs[4].View())
	b.WriteString("]\n")

	if m.submitting {
		b.WriteString("\n[Зарегистрироваться...]\n")
	} else {
		b.WriteString("\n[Зарегистрироваться]\n")
	}

	if m.errMsg != "" {
		b.WriteString("\nОшибка: ")
		b.WriteString(m.errMsg)
		b.WriteString("\n")
	}

	return renderPage("РЕГИСТРАЦИЯ", strings.TrimRight(b.String(), "\n"), "esc: назад │ tab: след. поле │ enter: подтвердить")
}

func (m *RegisterModel) cmdRegister(name, login, pass, hint string) tea.Cmd {
	ctx := m.ctx
	auth := m.auth

	return func() tea.Msg {
		err := auth.Register(ctx, models.User{
			Name:               name,
			Login:              login,
			MasterPassword:     pass,
			MasterPasswordHint: hint,
		})
		return RegisterResult{
			Err:      err,
			Username: login,
		}
	}
}

func (m *RegisterModel) resetForm() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
		m.inputs[i].Blur()
	}
	m.focus = 0
	m.inputs[m.focus].Focus()
}

func (m *RegisterModel) focusNext() {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + 1) % len(m.inputs)
	m.inputs[m.focus].Focus()
}

func (m *RegisterModel) focusPrev() {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
}
