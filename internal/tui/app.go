// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package tui

import (
	"github.com/MKhiriev/go-pass-keeper/models"
	tea "github.com/charmbracelet/bubbletea"
)

// RootModel is the top-level TUI router used during the login/registration flow.
// It is responsible for:
//  1. Keeping track of the currently active page model.
//  2. Handling the global Ctrl+C quit hotkey on every page.
//  3. Intercepting [NavigateTo] messages and switching the active page accordingly.
//  4. Delegating all other messages to the currently active page.
//  5. Collecting the [LoginResult] that signals a successful authentication and
//     terminating the Bubble Tea program with the authenticated user's data.
type RootModel struct {
	pages   map[string]tea.Model
	current tea.Model

	quitByUser bool
	resultID   int64
	resultKey  []byte
	buildInfo  models.AppBuildInfo

	showBuildInfo bool
}

// NewRootModel creates a [RootModel] with the provided page map and sets startPage as
// the initially active page. buildInfo is stored for optional display via the 'v' hotkey.
func NewRootModel(pages map[string]tea.Model, startPage string, buildInfo models.AppBuildInfo) RootModel {
	return RootModel{
		pages:     pages,
		current:   pages[startPage],
		buildInfo: buildInfo,
	}
}

// Init implements [tea.Model]. It delegates to the Init method of the currently active page.
func (r RootModel) Init() tea.Cmd {
	if r.current == nil {
		return nil
	}
	return r.current.Init()
}

// Update implements [tea.Model]. It processes the following messages at the router level:
//   - [tea.KeyMsg] "ctrl+c" — sets quitByUser and terminates the program.
//   - [tea.KeyMsg] "v"      — toggles the build-info overlay when on the menu page.
//   - [tea.KeyMsg] "esc"    — closes the build-info overlay.
//   - [NavigateTo]          — switches the active page; optionally dispatches a payload.
//   - [LoginResult]         — captures user ID and encryption key, then quits.
//
// All other messages are forwarded to the active page's Update method.
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

// View implements [tea.Model]. It renders the build-info overlay when it is active,
// otherwise delegates rendering to the current page's View method.
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
