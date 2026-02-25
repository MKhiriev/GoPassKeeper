package tui

import "github.com/charmbracelet/bubbles/textinput"

type registerModel struct {
	inputs     []textinput.Model
	focus      int
	submitting bool
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
		out += "[Зарегистрироваться...]\n\n"
	} else {
		out += "[Зарегистрироваться]\n\n"
	}
	out += "esc назад  tab следующее поле  enter подтвердить  q выход"
	return out
}
