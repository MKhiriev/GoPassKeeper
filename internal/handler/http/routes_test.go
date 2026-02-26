package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/internal/logger"
	"github.com/MKhiriev/go-pass-keeper/internal/service"
	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
)

// ---- Mock: AuthService ----

type mockAuthSvc struct{}

func (m *mockAuthSvc) RegisterUser(_ context.Context, u models.User) (models.User, error) {
	return u, nil
}
func (m *mockAuthSvc) Login(_ context.Context, u models.User) (models.User, error) {
	return u, nil
}
func (m *mockAuthSvc) CreateToken(_ context.Context, _ models.User) (models.Token, error) {
	return models.Token{}, nil
}
func (m *mockAuthSvc) ParseToken(_ context.Context, _ string) (models.Token, error) {
	return models.Token{UserID: 1}, nil
}
func (m *mockAuthSvc) Params(ctx context.Context, user models.User) (models.User, error) {
	return models.User{}, nil
}

// ---- Mock: AppInfoService ----

type mockAppInfoSvc struct{}

func (m *mockAppInfoSvc) GetAppVersion(_ context.Context) string {
	return "test-version"
}

// ---- Mock: PrivateDataService ----

type mockPrivateDataSvc struct {
	uploadFn      func(ctx context.Context, req models.UploadRequest) error
	downloadFn    func(ctx context.Context, req models.DownloadRequest) ([]models.PrivateData, error)
	downloadAllFn func(ctx context.Context, userID int64) ([]models.PrivateData, error)
	updateFn      func(ctx context.Context, req models.UpdateRequest) error
	deleteFn      func(ctx context.Context, req models.DeleteRequest) error
}

func (m *mockPrivateDataSvc) UploadPrivateData(ctx context.Context, req models.UploadRequest) error {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, req)
	}
	return nil
}
func (m *mockPrivateDataSvc) DownloadPrivateData(ctx context.Context, req models.DownloadRequest) ([]models.PrivateData, error) {
	if m.downloadFn != nil {
		return m.downloadFn(ctx, req)
	}
	return nil, nil
}
func (m *mockPrivateDataSvc) DownloadAllPrivateData(ctx context.Context, userID int64) ([]models.PrivateData, error) {
	if m.downloadAllFn != nil {
		return m.downloadAllFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockPrivateDataSvc) UpdatePrivateData(ctx context.Context, req models.UpdateRequest) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, req)
	}
	return nil
}
func (m *mockPrivateDataSvc) DeletePrivateData(ctx context.Context, req models.DeleteRequest) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, req)
	}
	return nil
}
func (m *mockPrivateDataSvc) DownloadUserPrivateDataStates(ctx context.Context, userID int64) ([]models.PrivateDataState, error) {
	return nil, nil
}
func (m *mockPrivateDataSvc) DownloadSpecificUserPrivateDataStates(ctx context.Context, req models.SyncRequest) ([]models.PrivateDataState, error) {
	return nil, nil
}

// ---- Helper ----

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	h := &Handler{
		logger: logger.Nop(),
		services: &service.Services{
			AuthService:        &mockAuthSvc{},
			AppInfoService:     &mockAppInfoSvc{},
			PrivateDataService: &mockPrivateDataSvc{},
		},
	}
	return h.Init()
}

func validAuthHeader() string { return "Bearer stub-token" }

// ---- Public routes: reachable without auth ----

func TestInit_PublicRoutes(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/auth/register"},
		{http.MethodPost, "/api/auth/login"},
		{http.MethodGet, "/api/version/"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.NotEqual(t, http.StatusNotFound, rr.Code,
				"route should be registered: %s %s", tt.method, tt.path)
		})
	}
}

// ---- Protected routes: 401 without token ----

func TestInit_ProtectedRoutes_RequireAuth(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/auth/settings/password/change"},
		{http.MethodPost, "/api/auth/settings/otp"},
		{http.MethodDelete, "/api/auth/settings/otp"},
		{http.MethodPost, "/api/data/"},
		{http.MethodGet, "/api/data/all"},
		{http.MethodPost, "/api/data/download"},
		{http.MethodPut, "/api/data/update"},
		{http.MethodDelete, "/api/data/delete"},
		{http.MethodGet, "/api/sync/"},
		{http.MethodGet, "/api/sync/specific"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path+" without token → 401", func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusUnauthorized, rr.Code,
				"missing token should result in 401")
		})
	}
}

// ---- Protected routes: pass with valid token ----

func TestInit_ProtectedRoutes_PassWithValidToken(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/data/all"},
		{http.MethodGet, "/api/sync/"},
		{http.MethodGet, "/api/sync/specific"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path+" with token → not 401", func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Authorization", validAuthHeader())
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.NotEqual(t, http.StatusUnauthorized, rr.Code,
				"valid token should not result in 401")
		})
	}
}

// ---- Unknown routes return 404 ----

func TestInit_UnknownRoutes_Return404(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		method  string
		path    string
		addAuth bool // some paths are auth-protected and require a token to reach 404
	}{
		{http.MethodGet, "/api/nonexistent", false},
		{http.MethodPost, "/api/data/unknown", true}, // /api/data/* is protected by auth
		{http.MethodGet, "/totally/wrong", false},
		{http.MethodPatch, "/api/auth/register", false},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.addAuth {
				req.Header.Set("Authorization", validAuthHeader())
			}
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusNotFound, rr.Code)
		})
	}
}

// ---- Wrong method on existing route returns 404 (CheckHTTPMethod) ----

func TestInit_WrongMethod_Returns404NotMethodNotAllowed(t *testing.T) {
	router := newTestRouter(t)

	tests := []struct {
		name    string
		method  string
		path    string
		addAuth bool // routes under h.auth require a token to reach MethodNotAllowed
	}{
		{
			name:   "GET on /api/auth/register (POST only)",
			method: http.MethodGet,
			path:   "/api/auth/register",
		},
		{
			name:   "GET on /api/auth/login (POST only)",
			method: http.MethodGet,
			path:   "/api/auth/login",
		},
		{
			name:   "POST on /api/version/ (GET only)",
			method: http.MethodPost,
			path:   "/api/version/",
		},
		{
			name:    "DELETE on /api/data/all (GET only)",
			method:  http.MethodDelete,
			path:    "/api/data/all",
			addAuth: true, // /api/data/* is wrapped by auth middleware
		},
		{
			name:    "GET on /api/data/update (PUT only)",
			method:  http.MethodGet,
			path:    "/api/data/update",
			addAuth: true, // /api/data/* is wrapped by auth middleware
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.addAuth {
				req.Header.Set("Authorization", validAuthHeader())
			}
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusNotFound, rr.Code,
				"CheckHTTPMethod should replace 405 with 404")
			assert.NotEqual(t, http.StatusMethodNotAllowed, rr.Code)
		})
	}
}

// ---- X-Trace-ID is always present in the response ----

func TestInit_TraceIDHeader_AlwaysSet(t *testing.T) {
	router := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.NotEmpty(t, rr.Header().Get("X-Trace-ID"))
}

// ---- Incoming X-Trace-ID is echoed back ----

func TestInit_TraceIDHeader_EchoedFromRequest(t *testing.T) {
	router := newTestRouter(t)
	const customTraceID = "my-custom-trace-id-12345"

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", nil)
	req.Header.Set("X-Trace-ID", customTraceID)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, customTraceID, rr.Header().Get("X-Trace-ID"))
}
