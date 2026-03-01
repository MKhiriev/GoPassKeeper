// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

// Package tui implements the terminal user interface (TUI) for the GoPassKeeper client.
//
// The package is built on top of the Bubble Tea framework (github.com/charmbracelet/bubbletea)
// and follows the Elm architecture: each screen is represented by a model with Init, Update,
// and View methods. Navigation between screens is performed via the [NavigateTo] message
// intercepted by the root model [RootModel].
//
// The entry point is the [TUI] type, created via [New]. The application lifecycle consists
// of two stages:
//   - [TUI.LoginFlow] — login and registration screens; terminates when the user
//     successfully authenticates or explicitly quits (Ctrl+C).
//   - [TUI.MainLoop] — the main record-management screen; terminates on quit (q / Ctrl+C)
//     or logout (l).
package tui

import (
	"context"
	"errors"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/models"
	tea "github.com/charmbracelet/bubbletea"
)

// ErrUserQuit is returned by [TUI.LoginFlow] when the user terminates the program
// with Ctrl+C before completing authentication.
var ErrUserQuit = errors.New("вышел из программы")

// ErrUserIDMissing is returned by [TUI.LoginFlow] when the login succeeded but
// the user ID could not be extracted from the final model state.
var ErrUserIDMissing = errors.New("не удалось получить user id после входа")

// TUI is the facade of the package. It holds a reference to client-side services
// and exposes methods for running each lifecycle stage of the application.
type TUI struct {
	services *service.ClientServices
}

// New creates and returns a new [TUI] instance.
// The logger parameter is reserved for future use and is currently ignored.
func New(services *service.ClientServices, _ *logger.Logger) (*TUI, error) {
	return &TUI{services: services}, nil
}

// LoginFlow launches the interactive login/registration TUI in alternate-screen mode
// (full-screen terminal mode).
//
// The method blocks until the user authenticates successfully or quits the program.
// On success it returns the authenticated user's ID and the encryption key received
// from the authentication service.
//
// Possible errors:
//   - [ErrUserQuit]    — the user pressed Ctrl+C without logging in.
//   - [ErrUserIDMissing] — login succeeded but the resulting user ID is zero.
//   - any other error  — failure inside the Bubble Tea program runtime.
func (t *TUI) LoginFlow(ctx context.Context, buildInfo models.AppBuildInfo) (userID int64, encryptionKey []byte, err error) {
	clearSessionUserID()

	pages := map[string]tea.Model{
		"menu":     NewMenuModel(),
		"login":    NewLoginModel(ctx, t.services.AuthService),
		"register": NewRegisterModel(ctx, t.services.AuthService),
	}

	root := NewRootModel(pages, "menu", buildInfo)
	finalModel, runErr := tea.NewProgram(root, tea.WithAltScreen()).Run()
	if runErr != nil {
		return 0, nil, runErr
	}

	result, ok := finalModel.(RootModel)
	if !ok {
		return 0, nil, tea.ErrProgramKilled
	}
	if result.quitByUser {
		clearSessionUserID()
		return 0, nil, ErrUserQuit
	}
	if result.resultID <= 0 {
		clearSessionUserID()
		return 0, nil, ErrUserIDMissing
	}
	setSessionUserID(result.resultID)

	return result.resultID, result.resultKey, nil
}

// MainLoop launches the primary record-management TUI in alternate-screen mode.
//
// If userID is greater than zero the session user ID is initialised from it; otherwise
// the value stored by a previous [TUI.LoginFlow] call is used.
// The method blocks until the user quits (q / Ctrl+C) or requests a logout (l).
//
// Returns logout=true when the user explicitly chose to log out so that the caller
// can re-run [TUI.LoginFlow] for a new session.
func (t *TUI) MainLoop(ctx context.Context, userID int64, buildInfo models.AppBuildInfo) (logout bool, err error) {
	if userID > 0 {
		setSessionUserID(userID)
	}

	model := newMainLoopModel(ctx, t.services, userID, buildInfo)
	finalModel, runErr := tea.NewProgram(model, tea.WithAltScreen()).Run()
	if runErr != nil {
		return false, runErr
	}

	result, ok := finalModel.(mainLoopModel)
	if !ok {
		return false, tea.ErrProgramKilled
	}
	return result.logout, nil
}
