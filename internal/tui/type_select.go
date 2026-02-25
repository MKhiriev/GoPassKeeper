package tui

type typeSelectModel struct {
	items []string
	idx   int
}

func newTypeSelectModel() typeSelectModel {
	return typeSelectModel{items: []string{"Логин/Пароль", "Текстовая заметка", "Банковская карта", "Файл"}}
}

func (m typeSelectModel) View() string {
	out := "Выберите тип записи:\n\n"
	for i, item := range m.items {
		cursor := "  "
		if i == m.idx {
			cursor = "> "
		}
		out += cursor + item + "\n"
	}
	out += "\nesc назад  enter выбрать"
	return out
}
