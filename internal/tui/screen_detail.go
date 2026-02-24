package tui

import "context"

func (t *TUI) showDetailScreen(ctx context.Context) error {
	id, err := t.readLine("ClientSideID: ")
	if err != nil {
		return err
	}

	item, err := t.dataSvc.Get(ctx, id)
	if err != nil {
		return err
	}

	t.printLn("id=%s", item.ClientSideID)
	t.printLn("type=%d", item.Payload.Type)
	t.printLn("version=%d", item.Version)
	t.printLn("metadata=%s", string(item.Payload.Metadata))
	t.printLn("data=%s", string(item.Payload.Data))
	if item.Payload.Notes != nil {
		t.printLn("notes=%s", string(*item.Payload.Notes))
	}
	if item.Payload.AdditionalFields != nil {
		t.printLn("additional_fields=%s", string(*item.Payload.AdditionalFields))
	}

	return nil
}
