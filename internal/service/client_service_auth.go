package service

import (
	"context"
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/adapter"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type clientAuthService struct {
	localStore store.LocalStorage
	adapter    adapter.ServerAdapter
	crypto     ClientCryptoService
}

func NewClientAuthService(localStore store.LocalStorage, serverAdapter adapter.ServerAdapter, crypto ClientCryptoService) ClientAuthService {
	return &clientAuthService{localStore: localStore, adapter: serverAdapter, crypto: crypto}
}

func (a *clientAuthService) Register(ctx context.Context, user models.User) error {
	registered, err := a.adapter.Register(ctx, user)
	if err != nil {
		return fmt.Errorf("register user on server: %w", err)
	}

	if err = a.localStore.SaveSession(ctx, registered.UserID, a.adapter.Token()); err != nil {
		return fmt.Errorf("save local session after register: %w", err)
	}
	return nil
}

func (a *clientAuthService) Login(ctx context.Context, user models.User) ([]byte, error) {
	token, err := a.adapter.Login(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("login on server: %w", err)
	}

	a.adapter.SetToken(token.SignedString)
	if err = a.localStore.SaveSession(ctx, token.UserID, token.SignedString); err != nil {
		return nil, fmt.Errorf("save local session after login: %w", err)
	}

	return a.crypto.DeriveKey(user.MasterPassword, token.UserID), nil
}

func (a *clientAuthService) RestoreSession(ctx context.Context) (int64, string, error) {
	userID, token, err := a.localStore.LoadSession(ctx)
	if err != nil {
		return 0, "", err
	}
	a.adapter.SetToken(token)
	return userID, token, nil
}

func (a *clientAuthService) Logout(ctx context.Context) error {
	a.adapter.SetToken("")
	if err := a.localStore.ClearSession(ctx); err != nil {
		return fmt.Errorf("clear local session: %w", err)
	}
	return nil
}
