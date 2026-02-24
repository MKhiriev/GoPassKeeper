package tui

import "context"

func (t *TUI) showEditScreen(ctx context.Context) error {
	id, err := t.readLine("ClientSideID: ")
	if err != nil {
		return err
	}

	item, err := t.dataSvc.Get(ctx, id)
	if err != nil {
		return err
	}

	payload, err := t.promptPayload(item.Payload)
	if err != nil {
		return err
	}

	item.Payload = payload
	if err = t.dataSvc.Update(ctx, item); err != nil {
		return err
	}

	t.printLn("updated")
	return nil
}

func (t *TUI) showDeleteScreen(ctx context.Context) error {
	id, err := t.readLine("ClientSideID: ")
	if err != nil {
		return err
	}

	item, err := t.dataSvc.Get(ctx, id)
	if err != nil {
		return err
	}

	if err = t.dataSvc.Delete(ctx, id, item.Version); err != nil {
		return err
	}

	t.printLn("deleted")
	return nil
}
