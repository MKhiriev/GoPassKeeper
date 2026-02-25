package tui

import (
	"github.com/MKhiriev/go-pass-keeper/models"
)

type authDoneMsg struct {
	userID int64
	key    []byte
}

type syncDoneMsg struct {
	err error
}

type listLoadedMsg struct {
	items []models.DecipheredPayload
	err   error
}

type itemSavedMsg struct {
	err error
}

type itemDeletedMsg struct {
	err error
}

type copiedMsg struct{}

type clearStatusMsg struct{}
