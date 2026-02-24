package adapter

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/go-resty/resty/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUnauthorized    = errors.New("client unauthorized")
	ErrVersionConflict = errors.New("version conflict")
)

type HTTPClientConfig struct {
	BaseURL string
	HashKey string
	Timeout time.Duration
}

type httpServerAdapter struct {
	client  *resty.Client
	hashKey string

	mu    sync.RWMutex
	token string
}

func NewHTTPServerAdapter(cfg HTTPClientConfig) ServerAdapter {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}

	cli := resty.New().
		SetBaseURL(strings.TrimRight(cfg.BaseURL, "/")).
		SetTimeout(cfg.Timeout)

	return &httpServerAdapter{client: cli, hashKey: cfg.HashKey}
}

func (h *httpServerAdapter) SetToken(token string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.token = strings.TrimSpace(token)
}

func (h *httpServerAdapter) Token() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
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

	token, err := parseBearerToken(resp.Header().Get("Authorization"))
	if err != nil {
		return models.User{}, fmt.Errorf("register parse bearer token: %w", err)
	}
	userID, err := parseUserIDFromJWT(token)
	if err != nil {
		return models.User{}, fmt.Errorf("register parse user id: %w", err)
	}

	h.SetToken(token)
	return models.User{UserID: userID, Login: user.Login, Name: user.Name}, nil
}

func (h *httpServerAdapter) Login(ctx context.Context, user models.User) (models.Token, error) {
	resp, err := h.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(user).
		Post("/api/auth/login")
	if err != nil {
		return models.Token{}, fmt.Errorf("login request: %w", err)
	}
	if err = mapHTTPError(resp); err != nil {
		return models.Token{}, err
	}

	token, err := parseBearerToken(resp.Header().Get("Authorization"))
	if err != nil {
		return models.Token{}, fmt.Errorf("login parse bearer token: %w", err)
	}
	userID, err := parseUserIDFromJWT(token)
	if err != nil {
		return models.Token{}, fmt.Errorf("login parse user id: %w", err)
	}

	h.SetToken(token)
	return models.Token{SignedString: token, UserID: userID}, nil
}

func (h *httpServerAdapter) Upload(ctx context.Context, req models.UploadRequest) error {
	req.Hash = computeTransportHash(req.PrivateDataList, h.hashKey)
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
	req.Hash = computeTransportHash(req.PrivateDataUpdates, h.hashKey)
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

func parseBearerToken(value string) (string, error) {
	parts := strings.Split(strings.TrimSpace(value), " ")
	if len(parts) != 2 || parts[1] == "" {
		return "", errors.New("invalid authorization header")
	}
	return parts[1], nil
}

func parseUserIDFromJWT(tokenString string) (int64, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return 0, err
	}

	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func computeTransportHash(v any, key string) string {
	if key == "" {
		return ""
	}
	payload, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}
