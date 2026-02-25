package tui

import (
	"context"
	"errors"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var ErrUserQuit = errors.New("вышел из программы")

type TUI struct {
	services *service.ClientServices
}

func New(services *service.ClientServices, _ *logger.Logger) (*TUI, error) {
	return &TUI{services: services}, nil
}

func (t *TUI) LoginFlow(ctx context.Context) (userID int64, encryptionKey []byte, err error) {
	m := newLoginFlowModel(ctx, t.services)
	finalModel, runErr := tea.NewProgram(m).Run()
	if runErr != nil {
		return 0, nil, runErr
	}

	result, ok := finalModel.(loginFlowModel)
	if !ok {
		return 0, nil, tea.ErrProgramKilled
	}
	if result.err != nil {
		return 0, nil, result.err
	}

	return result.resultUserID, result.resultKey, nil
}

func (t *TUI) MainLoop(_ context.Context, _ int64) (logout bool, err error) {
	m := mainLoopModel{}
	finalModel, runErr := tea.NewProgram(m).Run()
	if runErr != nil {
		return false, runErr
	}

	result, ok := finalModel.(mainLoopModel)
	if !ok {
		return false, tea.ErrProgramKilled
	}
	return result.logout, result.err
}

type screen int

const (
	screenWelcome screen = iota
	screenLogin
	screenRegister
)

type loginFlowModel struct {
	ctx      context.Context
	services *service.ClientServices

	screen   screen
	welcome  welcomeModel
	login    loginModel
	register registerModel

	err          error
	resultUserID int64
	resultKey    []byte
}

func newLoginFlowModel(ctx context.Context, services *service.ClientServices) loginFlowModel {
	return loginFlowModel{
		ctx:      ctx,
		services: services,
		screen:   screenWelcome,
		welcome:  newWelcomeModel(),
		login:    newLoginModel(),
		register: newRegisterModel(),
	}
}

func (m loginFlowModel) Init() tea.Cmd {
	return nil
}

func (m loginFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case authDoneMsg:
		m.resultUserID = msg.userID
		m.resultKey = msg.key
		return m, tea.Quit
	case authErrMsg:
		switch m.screen {
		case screenLogin:
			m.login.errMsg = msg.err.Error()
			m.login.submitting = false
		case screenRegister:
			m.register.errMsg = msg.err.Error()
			m.register.submitting = false
		}
		return m, nil
	}

	switch m.screen {
	case screenWelcome:
		return m.updateWelcome(msg)
	case screenLogin:
		return m.updateLogin(msg)
	case screenRegister:
		return m.updateRegister(msg)
	default:
		return m, nil
	}
}

func (m loginFlowModel) View() string {
	switch m.screen {
	case screenWelcome:
		return m.welcome.View()
	case screenLogin:
		return m.login.View()
	case screenRegister:
		return m.register.View()
	default:
		return ""
	}
}

func (m loginFlowModel) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "up", "k":
		if m.welcome.idx > 0 {
			m.welcome.idx--
		}
		m.welcome.confirmed = false
	case "down", "j":
		if m.welcome.idx < len(m.welcome.items)-1 {
			m.welcome.idx++
		}
		m.welcome.confirmed = false
	case "enter":
		m.welcome.confirmed = true
		if m.welcome.idx == 0 {
			m.screen = screenLogin
		} else {
			m.screen = screenRegister
		}
	case "ctrl+c", "q":
		m.welcome.confirmed = false
		m.err = ErrUserQuit
		return m, tea.Quit
	}

	return m, nil
}

func (m loginFlowModel) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch keyMsg.String() {
		case "ctrl+c", "q":
			m.err = ErrUserQuit
			return m, tea.Quit
		case "esc":
			m.welcome.confirmed = false
			m.screen = screenWelcome
			m.login.errMsg = ""
			m.login.submitting = false
			return m, nil
		case "tab":
			m.login = focusNextLogin(m.login)
			return m, nil
		case "shift+tab":
			m.login = focusPrevLogin(m.login)
			return m, nil
		case "enter":
			login := strings.TrimSpace(m.login.inputs[0].Value())
			pass := m.login.inputs[1].Value()
			if login == "" || pass == "" {
				m.login.errMsg = "Логин и пароль обязательны"
				return m, nil
			}
			m.login.submitting = true
			m.login.errMsg = ""
			return m, m.cmdLogin(models.User{Login: login, MasterPassword: pass})
		}
	}

	var cmd tea.Cmd
	m.login.inputs[m.login.focus], cmd = m.login.inputs[m.login.focus].Update(msg)
	return m, cmd
}

func (m loginFlowModel) updateRegister(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if ok {
		switch keyMsg.String() {
		case "ctrl+c", "q":
			m.err = ErrUserQuit
			return m, tea.Quit
		case "esc":
			m.welcome.confirmed = false
			m.screen = screenWelcome
			m.register.errMsg = ""
			m.register.submitting = false
			return m, nil
		case "tab":
			m.register = focusNextRegister(m.register)
			return m, nil
		case "shift+tab":
			m.register = focusPrevRegister(m.register)
			return m, nil
		case "enter":
			name := strings.TrimSpace(m.register.inputs[0].Value())
			login := strings.TrimSpace(m.register.inputs[1].Value())
			pass := m.register.inputs[2].Value()
			repeat := m.register.inputs[3].Value()
			hint := m.register.inputs[4].Value()
			if name == "" || login == "" || pass == "" {
				m.register.errMsg = "Имя, логин и пароль обязательны"
				return m, nil
			}
			if pass != repeat {
				m.register.errMsg = "Пароли не совпадают"
				return m, nil
			}
			m.register.submitting = true
			m.register.errMsg = ""
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

func (m loginFlowModel) cmdLogin(user models.User) tea.Cmd {
	auth := m.services.AuthService
	ctx := m.ctx

	return func() tea.Msg {
		userID, key, err := auth.Login(ctx, user)
		if err != nil {
			return authErrMsg{err: err}
		}
		return authDoneMsg{userID: userID, key: key}
	}
}

func (m loginFlowModel) cmdRegisterAndLogin(user models.User) tea.Cmd {
	auth := m.services.AuthService
	ctx := m.ctx

	return func() tea.Msg {
		if err := auth.Register(ctx, user); err != nil {
			return authErrMsg{err: err}
		}

		userID, key, err := auth.Login(ctx, models.User{Login: user.Login, MasterPassword: user.MasterPassword})
		if err != nil {
			return authErrMsg{err: err}
		}
		return authDoneMsg{userID: userID, key: key}
	}
}

type welcomeModel struct {
	items     []string
	idx       int
	confirmed bool
}

func newWelcomeModel() welcomeModel {
	return welcomeModel{items: []string{"Войти", "Зарегистрироваться"}}
}

func (m welcomeModel) View() string {
	out := "GoPassKeeper\n\nВыберите действие:\n\n"
	for i, item := range m.items {
		prefix := "  "
		if i == m.idx {
			prefix = "> "
		}
		out += prefix + item + "\n"
	}
	out += "\nHot keys:\nCtrl + C: Выход из программы"
	return out
}

type loginModel struct {
	inputs     []textinput.Model
	focus      int
	submitting bool
	errMsg     string
}

func newLoginModel() loginModel {
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

	return loginModel{inputs: []textinput.Model{loginInput, passwordInput}}
}

func (m loginModel) View() string {
	out := "Вход\n\n"
	out += "Логин:  [" + m.inputs[0].View() + "]\n"
	out += "Пароль: [" + m.inputs[1].View() + "]\n\n"
	if m.submitting {
		out += "[Войти...]\n"
	} else {
		out += "[Войти]\n"
	}
	if m.errMsg != "" {
		out += "\nОшибка: " + m.errMsg + "\n"
	}
	out += "\nesc назад  tab следующее поле  enter подтвердить  ctrl+c выход"
	return out
}

func focusNextLogin(m loginModel) loginModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus + 1) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

func focusPrevLogin(m loginModel) loginModel {
	m.inputs[m.focus].Blur()
	m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
	m.inputs[m.focus].Focus()
	return m
}

type registerModel struct {
	inputs     []textinput.Model
	focus      int
	submitting bool
	errMsg     string
}

func newRegisterModel() registerModel {
	fields := make([]textinput.Model, 5)

	fields[0] = textinput.New()
	fields[0].Placeholder = "name"
	fields[0].Width = 40
	fields[0].Focus()

	fields[1] = textinput.New()
	fields[1].Placeholder = "login"
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

	return registerModel{inputs: fields}
}

func (m registerModel) View() string {
	out := "Регистрация\n\n"
	out += "Имя:           [" + m.inputs[0].View() + "]\n"
	out += "Логин:         [" + m.inputs[1].View() + "]\n"
	out += "Пароль:        [" + m.inputs[2].View() + "]\n"
	out += "Повтор пароля: [" + m.inputs[3].View() + "]\n"
	out += "Подсказка:     [" + m.inputs[4].View() + "]\n\n"
	if m.submitting {
		out += "[Зарегистрироваться...]\n"
	} else {
		out += "[Зарегистрироваться]\n"
	}
	if m.errMsg != "" {
		out += "\nОшибка: " + m.errMsg + "\n"
	}
	out += "\nesc назад  tab следующее поле  enter подтвердить  ctrl+c выход"
	return out
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

type authDoneMsg struct {
	userID int64
	key    []byte
}

type authErrMsg struct {
	err error
}

type mainLoopModel struct {
	logout bool
	err    error
}

func (m mainLoopModel) Init() tea.Cmd {
	return nil
}

func (m mainLoopModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "l":
		m.logout = true
		return m, tea.Quit
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m mainLoopModel) View() string {
	return "Главный экран\n\nl logout\nCtrl + C: Выход из программы"
}
