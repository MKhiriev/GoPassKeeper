package tui

import (
	"github.com/MKhiriev/go-pass-keeper/models"
	tea "github.com/charmbracelet/bubbletea"
)

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
	buildInfo  models.AppBuildInfo

	showBuildInfo bool
}

// NewRootModel registers all pages and opens startPage.
func NewRootModel(pages map[string]tea.Model, startPage string, buildInfo models.AppBuildInfo) RootModel {
	return RootModel{
		pages:     pages,
		current:   pages[startPage],
		buildInfo: buildInfo,
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
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c":
			r.quitByUser = true
			return r, tea.Quit
		case "v":
			if r.isMenuPage() {
				r.showBuildInfo = !r.showBuildInfo
				return r, nil
			}
		case "esc":
			if r.showBuildInfo {
				r.showBuildInfo = false
				return r, nil
			}
		}

		if r.showBuildInfo {
			return r, nil
		}
	}

	// Cross-page navigation.
	if nav, ok := msg.(NavigateTo); ok {
		next, exists := r.pages[nav.Page]
		if !exists {
			return r, nil
		}

		r.showBuildInfo = false
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
			setSessionUserID(result.UserID)
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
	if r.showBuildInfo {
		return renderBuildInfoWindow(r.buildInfo)
	}
	if r.current == nil {
		return renderPage("TUI", "", "")
	}
	return r.current.View()
}

func (r RootModel) isMenuPage() bool {
	_, ok := r.current.(*MenuModel)
	return ok
}
