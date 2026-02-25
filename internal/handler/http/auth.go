package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/internal/store"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromRequest(r)

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Err(err).Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	registeredUser, err := h.services.AuthService.RegisterUser(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDataProvided):
			log.Err(err).Msg("invalid data provided")
			http.Error(w, "invalid data provided", http.StatusBadRequest)
			return
		case errors.Is(err, store.ErrLoginAlreadyExists):
			log.Err(err).Msg("login already exists")
			http.Error(w, "login already exists", http.StatusConflict)
			return
		default:
			log.Err(err).Msg("unexpected error occurred during user registration")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	token, err := h.services.AuthService.CreateToken(ctx, registeredUser)
	if err != nil {
		log.Err(err).Msg("creation of token failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token.SignedString))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromRequest(r)

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Err(err).Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	log.Debug().Any("received user info", user).Send()

	foundUser, err := h.services.AuthService.Login(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDataProvided):
			log.Err(err).Msg("invalid data provided")
			http.Error(w, "invalid data provided", http.StatusBadRequest)
			return
		case errors.Is(err, store.ErrNoUserWasFound) || errors.Is(err, service.ErrWrongPassword):
			log.Err(err).Msg("no user was found/wrong password")
			http.Error(w, "invalid login/password", http.StatusUnauthorized)
			return
		default:
			log.Err(err).Msg("unexpected error occurred during user login")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	log.Debug().Int64("id", foundUser.UserID).Any("found user", foundUser).Msg("user successfully logged in")

	token, err := h.services.AuthService.CreateToken(ctx, foundUser)
	if err != nil {
		log.Err(err).Msg("creation of token failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token.SignedString))
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) params(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromRequest(r)

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Err(err).Msg("Invalid JSON was passed")
		http.Error(w, "Invalid JSON was passed", http.StatusBadRequest)
		return
	}

	log.Debug().Any("received user info", user).Send()

	foundUser, err := h.services.AuthService.Params(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDataProvided):
			log.Err(err).Msg("invalid data provided")
			http.Error(w, "invalid data provided", http.StatusBadRequest)
			return
		case errors.Is(err, store.ErrNoUserWasFound) || errors.Is(err, service.ErrWrongPassword):
			log.Err(err).Msg("no user was found/wrong password")
			http.Error(w, "invalid login/password", http.StatusUnauthorized)
			return
		default:
			log.Err(err).Msg("unexpected error occurred during user login")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	log.Debug().Int64("id", foundUser.UserID).Any("found user", foundUser).Msg("user successfully logged in")

	userParam := models.User{
		Login:          foundUser.Login,
		EncryptionSalt: foundUser.EncryptionSalt,
	}

	utils.WriteJSON(w, userParam, http.StatusOK)
}
