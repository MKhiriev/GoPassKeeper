package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
)

type TUI struct {
	reader *bufio.Reader
	out    io.Writer

	authSvc service.ClientAuthService
	dataSvc service.ClientPrivateDataService
	syncSvc service.ClientSyncService
}

func New(services *service.ClientServices, logger *logger.Logger) (*TUI, error) {
	return &TUI{
		reader:  bufio.NewReader(os.Stdin),
		out:     os.Stdout,
		authSvc: services.AuthService,
		dataSvc: services.PrivateDataService,
		syncSvc: services.SyncService,
	}, nil
}

func (t *TUI) readLine(prompt string) (string, error) {
	if prompt != "" {
		fmt.Fprint(t.out, prompt)
	}
	line, err := t.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (t *TUI) printLn(format string, args ...any) {
	fmt.Fprintf(t.out, format+"\n", args...)
}

func (t *TUI) LoginFlow(ctx context.Context) (int64, []byte, error) {
	return t.showLoginScreen(ctx)
}

// TODO change
func (t *TUI) UnlockWithMasterPassword(userID int64, derive func(string, int64) []byte) ([]byte, error) {
	password, err := t.readLine("Master password: ")
	if err != nil {
		return nil, err
	}
	return derive(password, userID), nil
}

func (t *TUI) MainLoop(ctx context.Context, userID int64) (bool, error) {
	for {
		t.printLn("\n=== GoPassKeeper Client ===")
		t.printLn("1) List")
		t.printLn("2) Create")
		t.printLn("3) Detail")
		t.printLn("4) Edit")
		t.printLn("5) Delete")
		t.printLn("6) Sync now")
		t.printLn("7) Logout")
		t.printLn("8) Exit")

		choice, err := t.readLine("> ")
		if err != nil {
			return false, err
		}

		switch choice {
		case "1":
			if err = t.showListScreen(ctx, userID); err != nil {
				t.printLn("list error: %v", err)
			}
		case "2":
			if err = t.showCreateScreen(ctx, userID); err != nil {
				t.printLn("create error: %v", err)
			}
		case "3":
			if err = t.showDetailScreen(ctx); err != nil {
				t.printLn("detail error: %v", err)
			}
		case "4":
			if err = t.showEditScreen(ctx); err != nil {
				t.printLn("edit error: %v", err)
			}
		case "5":
			if err = t.showDeleteScreen(ctx); err != nil {
				t.printLn("delete error: %v", err)
			}
		case "6":
			if err = t.syncSvc.FullSync(ctx, userID); err != nil {
				t.printLn("sync error: %v", err)
			} else {
				t.printLn("sync complete")
			}
		case "7":
			if err = t.authSvc.Logout(ctx); err != nil {
				return false, err
			}
			return true, nil
		case "8":
			return false, nil
		default:
			t.printLn("unknown option")
		}
	}
}
