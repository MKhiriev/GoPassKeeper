package tui

type errorOverlayModel struct {
	message string
}

func (m errorOverlayModel) View() string {
	content := "Ошибка\n\n" + m.message + "\n\nenter / esc закрыть"
	return overlayBoxStyle.Render(content)
}
