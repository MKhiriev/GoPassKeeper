package tui

import (
	"context"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	tea "github.com/charmbracelet/bubbletea"
)

type TUI struct {
	services *service.ClientServices
}

func New(services *service.ClientServices, _ *logger.Logger) (*TUI, error) {
	return &TUI{services: services}, nil
}

func (t *TUI) LoginFlow(ctx context.Context) (userID int64, encryptionKey []byte, err error) {
	model := newLoginAppModel(ctx, t.services)
	p := tea.NewProgram(model)
	finalModel, runErr := p.Run()
	if runErr != nil {
		return 0, nil, runErr
	}

	result, ok := finalModel.(appModel)
	if !ok {
		return 0, nil, tea.ErrProgramKilled
	}
	if result.err != nil {
		return 0, nil, result.err
	}

	return result.resultUserID, result.resultKey, nil
}

func (t *TUI) MainLoop(ctx context.Context, userID int64) (logout bool, err error) {
	model := newMainAppModel(ctx, t.services, userID)
	p := tea.NewProgram(model)
	finalModel, runErr := p.Run()
	if runErr != nil {
		return false, runErr
	}

	result, ok := finalModel.(appModel)
	if !ok {
		return false, tea.ErrProgramKilled
	}
	if result.err != nil {
		return false, result.err
	}
	return result.logout, nil
}
