package tui

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/models"
)

func (t *TUI) showLoginScreen(ctx context.Context) (int64, []byte, error) {
	for {
		t.printLn("\n1) Login")
		t.printLn("2) Register")
		choice, err := t.readLine("> ")
		if err != nil {
			return 0, nil, err
		}

		switch choice {
		case "1":
			return t.handleLogin(ctx)
		case "2":
			return t.handleRegister(ctx)
		default:
			t.printLn("unknown option")
		}
	}
}

func (t *TUI) handleLogin(ctx context.Context) (int64, []byte, error) {
	login, err := t.readLine("Login: ")
	if err != nil {
		return 0, nil, err
	}
	password, err := t.readLine("Master password: ")
	if err != nil {
		return 0, nil, err
	}

	key, err := t.authSvc.Login(ctx, models.User{Login: login, MasterPassword: password})
	if err != nil {
		return 0, nil, fmt.Errorf("login failed: %w", err)
	}

	userID, _, err := t.authSvc.RestoreSession(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("load session after login: %w", err)
	}

	return userID, key, nil
}

func (t *TUI) handleRegister(ctx context.Context) (int64, []byte, error) {
	name, err := t.readLine("Name: ")
	if err != nil {
		return 0, nil, err
	}
	login, err := t.readLine("Login: ")
	if err != nil {
		return 0, nil, err
	}
	password, err := t.readLine("Master password: ")
	if err != nil {
		return 0, nil, err
	}

	user := models.User{Name: name, Login: login, MasterPassword: password}
	if err = t.authSvc.Register(ctx, user); err != nil {
		return 0, nil, fmt.Errorf("register failed: %w", err)
	}

	key, err := t.authSvc.Login(ctx, user)
	if err != nil {
		return 0, nil, fmt.Errorf("login after register failed: %w", err)
	}

	userID, _, err := t.authSvc.RestoreSession(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("load session after register: %w", err)
	}

	return userID, key, nil
}
