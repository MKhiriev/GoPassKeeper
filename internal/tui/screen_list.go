package tui

import "context"

func (t *TUI) showListScreen(ctx context.Context, userID int64) error {
	items, err := t.dataSvc.GetAll(ctx, userID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		t.printLn("empty vault")
		return nil
	}
	for _, item := range items {
		t.printLn("- id=%s type=%d version=%d deleted=%v", item.ClientSideID, item.Payload.Type, item.Version, item.Deleted)
	}
	return nil
}
