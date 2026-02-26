package tui

import tea "github.com/charmbracelet/bubbletea"

// RootModel is a TUI router:
// 1) keeps active page
// 2) handles global Ctrl+C quit
// 3) handles NavigateTo messages
// 4) delegates all other messages to the active page
type RootModel struct {
	pages   map[string]tea.Model
	current tea.Model

	quitByUser bool
	resultID   int64
	resultKey  []byte
}

// NewRootModel registers all pages and opens startPage.
func NewRootModel(pages map[string]tea.Model, startPage string) RootModel {
	return RootModel{
		pages:   pages,
		current: pages[startPage],
	}
}

func (r RootModel) Init() tea.Cmd {
	if r.current == nil {
		return nil
	}
	return r.current.Init()
}

func (r RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global hotkey for every page.
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "ctrl+c" {
		r.quitByUser = true
		return r, tea.Quit
	}

	// Cross-page navigation.
	if nav, ok := msg.(NavigateTo); ok {
		next, exists := r.pages[nav.Page]
		if !exists {
			return r, nil
		}

		r.current = next

		if nav.Payload != nil {
			return r, func() tea.Msg { return nav.Payload }
		}
		return r, r.current.Init()
	}

	// Finalize login/register flow on success.
	switch result := msg.(type) {
	case LoginResult:
		if result.Err == nil {
			r.resultID = result.UserID
			r.resultKey = result.EncryptionKey
			return r, tea.Quit
		}
	}

	if r.current == nil {
		return r, nil
	}

	updated, cmd := r.current.Update(msg)
	r.current = updated
	return r, cmd
}

func (r RootModel) View() string {
	if r.current == nil {
		return renderPage("TUI", "", "")
	}
	return r.current.View()
}
