package tui

import "github.com/charmbracelet/bubbles/spinner"

type syncModel struct {
	spinner spinner.Model
	running bool
}

func newSyncModel() syncModel {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return syncModel{spinner: s}
}

func (m syncModel) View() string {
	return m.spinner.View() + " Синхронизация..."
}
