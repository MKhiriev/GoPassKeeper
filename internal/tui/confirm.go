package tui

type confirmModel struct {
	message string
}

func (m confirmModel) View() string {
	content := "Удалить \"" + m.message + "\"?\n\n"
	content += "y да    n нет"
	return overlayBoxStyle.Render(content)
}
