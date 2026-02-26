package tui

import (
	"context"
	"errors"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	tea "github.com/charmbracelet/bubbletea"
)

var ErrUserQuit = errors.New("вышел из программы")

type TUI struct {
	services *service.ClientServices
}

func New(services *service.ClientServices, _ *logger.Logger) (*TUI, error) {
	return &TUI{services: services}, nil
}

func (t *TUI) LoginFlow(ctx context.Context) (userID int64, encryptionKey []byte, err error) {
	pages := map[string]tea.Model{
		"menu":     NewMenuModel(),
		"login":    NewLoginModel(ctx, t.services.AuthService),
		"register": NewRegisterModel(ctx, t.services.AuthService),
	}

	root := NewRootModel(pages, "menu")
	finalModel, runErr := tea.NewProgram(root, tea.WithAltScreen()).Run()
	if runErr != nil {
		return 0, nil, runErr
	}

	result, ok := finalModel.(RootModel)
	if !ok {
		return 0, nil, tea.ErrProgramKilled
	}
	if result.quitByUser {
		return 0, nil, ErrUserQuit
	}

	return result.resultID, result.resultKey, nil
}

func (t *TUI) MainLoop(ctx context.Context, userID int64) (logout bool, err error) {
	model := newMainLoopModel(ctx, t.services, userID)
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
