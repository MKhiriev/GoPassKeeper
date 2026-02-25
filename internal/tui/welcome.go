package tui

type welcomeModel struct {
	items []string
	idx   int
}

func newWelcomeModel() welcomeModel {
	return welcomeModel{items: []string{"Войти", "Зарегистрироваться"}}
}

func (m welcomeModel) View() string {
	out := "GoPassKeeper\n\nВыберите действие:\n\n"
	for i, item := range m.items {
		cursor := "  "
		if i == m.idx {
			cursor = "> "
		}
		out += cursor + item + "\n"
	}
	out += "\nq выход"
	return out
}
