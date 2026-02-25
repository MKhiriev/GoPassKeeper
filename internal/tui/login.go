package tui

import "github.com/charmbracelet/bubbles/textinput"

type loginModel struct {
	inputs     []textinput.Model
	focus      int
	submitting bool
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
		out += "[Войти...]\n\n"
	} else {
		out += "[Войти]\n\n"
	}
	out += "esc назад  tab следующее поле  enter подтвердить  q выход"
	return out
}
