package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLogger_NotNil verifies that NewLogger returns a non-nil *Logger.
func TestNewLogger_NotNil(t *testing.T) {
	l := NewLogger("test")
	require.NotNil(t, l)
}

// TestNewLogger_RoleField verifies that every log entry produced by a logger
// created with NewLogger contains the expected "role" field.
func TestNewLogger_RoleField(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger("test-role")
	// redirect output to buffer for inspection
	l.Logger = l.Output(&buf)

	l.Info().Msg("hello")

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "test-role", entry["role"])
}

// TestNewLogger_ContainsTimestamp verifies that log entries contain a timestamp field.
func TestNewLogger_ContainsTimestamp(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger("ts-role")
	l.Logger = l.Output(&buf)

	l.Info().Msg("ts check")

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	_, hasTime := entry["time"]
	assert.True(t, hasTime, "expected 'time' field in log entry")
}

// TestNewLogger_CallerFieldName verifies that the caller field is named "func".
func TestNewLogger_CallerFieldName(t *testing.T) {
	NewLogger("caller-role") // sets zerolog.CallerFieldName as a side-effect
	assert.Equal(t, "func", zerolog.CallerFieldName)
}

// TestNewLogger_GlobalLevelIsDebug verifies that NewLogger sets the global
// zerolog level to Debug.
func TestNewLogger_GlobalLevelIsDebug(t *testing.T) {
	NewLogger("level-role")
	assert.Equal(t, zerolog.DebugLevel, zerolog.GlobalLevel())
}

// TestNop_NotNil verifies that Nop returns a non-nil *Logger.
func TestNop_NotNil(t *testing.T) {
	l := Nop()
	require.NotNil(t, l)
}

// TestNop_DiscardsOutput verifies that a Nop logger produces no output.
func TestNop_DiscardsOutput(t *testing.T) {
	var buf bytes.Buffer
	l := Nop()
	l.Logger = l.Output(&buf)

	l.Info().Msg("should be discarded")

	assert.Empty(t, buf.String(), "Nop logger should produce no output")
}

// TestGetChildLogger_NotNil verifies that GetChildLogger returns a non-nil *Logger.
func TestGetChildLogger_NotNil(t *testing.T) {
	parent := NewLogger("parent")
	child := parent.GetChildLogger()
	require.NotNil(t, child)
}

// TestGetChildLogger_IsIndependent verifies that the child logger is a
// distinct instance from the parent.
func TestGetChildLogger_IsIndependent(t *testing.T) {
	parent := NewLogger("parent")
	child := parent.GetChildLogger()
	assert.NotSame(t, parent, child)
}

// TestGetChildLogger_InheritsFields verifies that the child logger inherits
// context fields (e.g. "role") from the parent.
func TestGetChildLogger_InheritsFields(t *testing.T) {
	var buf bytes.Buffer
	parent := NewLogger("inherited-role")
	parent.Logger = parent.Output(&buf)

	child := parent.GetChildLogger()
	// write through child — buf is shared via the underlying writer
	child.Logger = child.Output(&buf)
	child.Info().Msg("child message")

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "inherited-role", entry["role"])
}

// TestFromContext_NotNil verifies that FromContext never returns nil, even
// when no logger has been explicitly attached to the context.
func TestFromContext_NotNil(t *testing.T) {
	l := FromContext(context.Background())
	require.NotNil(t, l)
}

// TestFromContext_ReturnsAttachedLogger verifies that FromContext returns the
// logger that was previously attached to the context via zerolog.
func TestFromContext_ReturnsAttachedLogger(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Str("ctx-key", "ctx-value").Logger()
	ctx := zl.WithContext(context.Background())

	l := FromContext(ctx)
	require.NotNil(t, l)

	l.Info().Msg("from context")

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "ctx-value", entry["ctx-key"])
}

// TestFromRequest_NotNil verifies that FromRequest never returns nil.
func TestFromRequest_NotNil(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	l := FromRequest(req)
	require.NotNil(t, l)
}

// TestFromRequest_ReturnsAttachedLogger verifies that FromRequest returns the
// logger attached to the request's context.
func TestFromRequest_ReturnsAttachedLogger(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf).With().Str("req-key", "req-value").Logger()
	ctx := zl.WithContext(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)

	// also register with zerolog's global log.Ctx mechanism
	log.Ctx(ctx) // warm up

	l := FromRequest(req)
	require.NotNil(t, l)

	l.Info().Msg("from request")

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "req-value", entry["req-key"])
}
