package store

import (
	"context"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPrivateDataFileStorage(t *testing.T) {
	s := NewPrivateDataFileStorage()
	require.NotNil(t, s)
}

func TestSaveBinaryDataToFile_NotImplemented(t *testing.T) {
	s := NewPrivateDataFileStorage()
	ctx := context.Background()

	t.Run("single item", func(t *testing.T) {
		err := s.SaveBinaryDataToFile(ctx, "test.bin", models.PrivateData{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("multiple items", func(t *testing.T) {
		err := s.SaveBinaryDataToFile(ctx, "test.bin", models.PrivateData{}, models.PrivateData{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("no items", func(t *testing.T) {
		err := s.SaveBinaryDataToFile(ctx, "test.bin")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("empty filename", func(t *testing.T) {
		err := s.SaveBinaryDataToFile(ctx, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented")
	})
}

func TestLoadBinaryDataFromFile_NotImplemented(t *testing.T) {
	s := NewPrivateDataFileStorage()
	ctx := context.Background()

	t.Run("normal filename", func(t *testing.T) {
		data, err := s.LoadBinaryDataFromFile(ctx, "test.bin")
		require.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "not implemented")
	})

	t.Run("empty filename", func(t *testing.T) {
		data, err := s.LoadBinaryDataFromFile(ctx, "")
		require.Error(t, err)
		assert.Nil(t, data)
		assert.Contains(t, err.Error(), "not implemented")
	})
}
