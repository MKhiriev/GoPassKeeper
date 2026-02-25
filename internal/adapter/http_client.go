package adapter

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MKhiriev/go-pass-keeper/internal/config"
	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/go-resty/resty/v2"
)

type httpServerAdapter struct {
	client *utils.HTTPClient

	hashKey string
	token   string

	logger *logger.Logger
}

func NewHTTPServerAdapter(adapterCfg config.ClientAdapter, appCfg config.ClientApp, logger *logger.Logger) (ServerAdapter, error) {
	client := utils.NewHTTPClient()

	client.
		SetBaseURL(adapterCfg.HTTPAddress).
		SetTimeout(adapterCfg.RequestTimeout)

	utils.InitHasherPool(appCfg.HashKey)

	return &httpServerAdapter{client: client, hashKey: appCfg.HashKey, logger: logger}, nil
}

func (h *httpServerAdapter) SetToken(token string) {
	h.token = strings.TrimSpace(token)
}

func (h *httpServerAdapter) Token() string {
	return h.token
}

func (h *httpServerAdapter) Register(ctx context.Context, user models.User) (models.User, error) {
	resp, err := h.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(user).
		Post("/api/auth/register")
	if err != nil {
		return models.User{}, fmt.Errorf("register request: %w", err)
	}
	if err = mapHTTPError(resp); err != nil {
		return models.User{}, err
	}

	token, err := utils.ParseBearerToken(resp.Header().Get("Authorization"))
	if err != nil {
		return models.User{}, fmt.Errorf("register parse bearer token: %w", err)
	}
	userID, err := utils.ParseUserIDFromJWT(token)
	if err != nil {
		return models.User{}, fmt.Errorf("register parse user id: %w", err)
	}

	h.SetToken(token)
	return models.User{UserID: userID, Login: user.Login, Name: user.Name}, nil
}

func (h *httpServerAdapter) RequestSalt(ctx context.Context, user models.User) (models.User, error) {
	var foundUser models.User // only login and encryption salt

	resp, err := h.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(user).
		SetResult(foundUser).
		Post("/api/auth/params")

	if err != nil {
		return user, fmt.Errorf("request request: %w", err)
	}
	if err = mapHTTPError(resp); err != nil {
		return user, err
	}

	return foundUser, nil
}

func (h *httpServerAdapter) Login(ctx context.Context, user models.User) (models.User, error) {
	var foundUser models.User

	resp, err := h.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(user).
		SetResult(foundUser).
		Post("/api/auth/login")

	if err != nil {
		return user, fmt.Errorf("login request: %w", err)
	}
	if err = mapHTTPError(resp); err != nil {
		return user, err
	}

	token, err := utils.ParseBearerToken(resp.Header().Get("Authorization"))
	if err != nil {
		return user, fmt.Errorf("login parse bearer token: %w", err)
	}

	h.SetToken(token)
	return foundUser, nil
}

func (h *httpServerAdapter) Upload(ctx context.Context, req models.UploadRequest) error {
	req.Hash = computeTransportHash(req.PrivateDataList)
	req.Length = len(req.PrivateDataList)

	resp, err := h.authedRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/data/")
	if err != nil {
		return fmt.Errorf("upload request: %w", err)
	}

	return mapHTTPError(resp)
}

func (h *httpServerAdapter) Download(ctx context.Context, req models.DownloadRequest) ([]models.PrivateData, error) {
	req.Length = len(req.ClientSideIDs)

	resp, err := h.authedRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/api/data/download")
	if err != nil {
		return nil, fmt.Errorf("download request: %w", err)
	}
	if err = mapHTTPError(resp); err != nil {
		return nil, err
	}

	var items []models.PrivateData
	if err = json.Unmarshal(resp.Body(), &items); err != nil {
		return nil, fmt.Errorf("decode download response: %w", err)
	}

	return items, nil
}

func (h *httpServerAdapter) Update(ctx context.Context, req models.UpdateRequest) error {
	req.Hash = computeTransportHash(req.PrivateDataUpdates)
	req.Length = len(req.PrivateDataUpdates)

	resp, err := h.authedRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Put("/api/data/update")
	if err != nil {
		return fmt.Errorf("update request: %w", err)
	}

	return mapHTTPError(resp)
}

func (h *httpServerAdapter) Delete(ctx context.Context, req models.DeleteRequest) error {
	req.Length = len(req.DeleteEntries)

	resp, err := h.authedRequest(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Delete("/api/data/delete")
	if err != nil {
		return fmt.Errorf("delete request: %w", err)
	}

	return mapHTTPError(resp)
}

func (h *httpServerAdapter) GetServerStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	resp, err := h.authedRequest(ctx).Get("/api/sync/")
	if err != nil {
		return nil, fmt.Errorf("get server states request: %w", err)
	}
	if err = mapHTTPError(resp); err != nil {
		return nil, err
	}

	var sr models.SyncResponse
	if err = json.Unmarshal(resp.Body(), &sr); err != nil {
		return nil, fmt.Errorf("decode server sync response: %w", err)
	}
	return sr.PrivateDataStates, nil
}

func (h *httpServerAdapter) authedRequest(ctx context.Context) *resty.Request {
	req := h.client.R().SetContext(ctx)
	if token := h.Token(); token != "" {
		req.SetHeader("Authorization", "Bearer "+token)
	}
	return req
}

func mapHTTPError(resp *resty.Response) error {
	if resp.StatusCode() >= http.StatusOK && resp.StatusCode() < http.StatusMultipleChoices {
		return nil
	}

	body := strings.TrimSpace(string(resp.Body()))
	bodyLower := strings.ToLower(body)

	if resp.StatusCode() == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if resp.StatusCode() == http.StatusConflict || strings.Contains(bodyLower, "version conflict") {
		return ErrVersionConflict
	}
	if body == "" {
		body = http.StatusText(resp.StatusCode())
	}
	return fmt.Errorf("http %d: %s", resp.StatusCode(), body)
}

func computeTransportHash(v any) string {
	payload, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	return hex.EncodeToString(utils.Hash(payload))
}
