package service

import (
	"fmt"

	"github.com/MKhiriev/go-pass-keeper/internal/crypto"
	"github.com/MKhiriev/go-pass-keeper/internal/utils"
	"github.com/MKhiriev/go-pass-keeper/models"
)

type clientCryptoService struct {
	key    []byte // DEK — set after successful login via SetEncryptionKey
	crypto crypto.KeyChainService
}

func NewClientCryptoService(crypto crypto.KeyChainService) ClientCryptoService {
	return &clientCryptoService{crypto: crypto}
}

func (c *clientCryptoService) SetEncryptionKey(key []byte) {
	c.key = key
}

// dataPayload bundles all typed data fields into a single value before encryption.
// Only one field will be non-nil depending on the DataType.
type dataPayload struct {
	LoginData    *models.LoginData    `json:"login_data,omitempty"`
	LoginURI     *models.LoginURI     `json:"login_uri,omitempty"`
	TextData     *models.TextData     `json:"text_data,omitempty"`
	BinaryData   *models.BinaryData   `json:"binary_data,omitempty"`
	BankCardData *models.BankCardData `json:"bank_card_data,omitempty"`
}

func (c *clientCryptoService) EncryptPayload(plain models.DecipheredPayload) (models.PrivateDataPayload, error) {
	// --- Metadata ---
	encMeta, err := c.crypto.EncryptData(plain.Metadata, c.key)
	if err != nil {
		return models.PrivateDataPayload{}, fmt.Errorf("encrypt metadata: %w", err)
	}

	// --- Data: bundle all typed fields into one struct, then encrypt ---
	dp := dataPayload{
		LoginData:    plain.LoginData,
		LoginURI:     plain.LoginURI,
		TextData:     plain.TextData,
		BinaryData:   plain.BinaryData,
		BankCardData: plain.BankCardData,
	}
	encData, err := c.crypto.EncryptData(dp, c.key)
	if err != nil {
		return models.PrivateDataPayload{}, fmt.Errorf("encrypt data: %w", err)
	}

	out := models.PrivateDataPayload{
		Metadata: models.CipheredMetadata(encMeta),
		Type:     plain.Type, // Type is NOT encrypted — plain int
		Data:     models.CipheredData(encData),
	}

	// --- Notes (optional) ---
	if plain.Notes != nil {
		encNotes, err := c.crypto.EncryptData(plain.Notes, c.key)
		if err != nil {
			return models.PrivateDataPayload{}, fmt.Errorf("encrypt notes: %w", err)
		}
		out.Notes = (*models.CipheredNotes)(&encNotes)
	}

	// --- AdditionalFields (optional) ---
	if plain.AdditionalFields != nil {
		encFields, err := c.crypto.EncryptData(plain.AdditionalFields, c.key)
		if err != nil {
			return models.PrivateDataPayload{}, fmt.Errorf("encrypt additional fields: %w", err)
		}
		out.AdditionalFields = (*models.CipheredCustomFields)(&encFields)
	}

	return out, nil
}

func (c *clientCryptoService) DecryptPayload(enc models.PrivateDataPayload) (models.DecipheredPayload, error) {
	// --- Metadata ---
	var meta models.Metadata
	if err := c.crypto.DecryptData(string(enc.Metadata), c.key, &meta); err != nil {
		return models.DecipheredPayload{}, fmt.Errorf("decrypt metadata: %w", err)
	}

	// --- Data ---
	var dp dataPayload
	if err := c.crypto.DecryptData(string(enc.Data), c.key, &dp); err != nil {
		return models.DecipheredPayload{}, fmt.Errorf("decrypt data: %w", err)
	}

	out := models.DecipheredPayload{
		Metadata:     meta,
		Type:         enc.Type, // Type is NOT encrypted — plain int
		LoginData:    dp.LoginData,
		LoginURI:     dp.LoginURI,
		TextData:     dp.TextData,
		BinaryData:   dp.BinaryData,
		BankCardData: dp.BankCardData,
	}

	// --- Notes (optional) ---
	if enc.Notes != nil {
		var notes models.Notes
		if err := c.crypto.DecryptData(string(*enc.Notes), c.key, &notes); err != nil {
			return models.DecipheredPayload{}, fmt.Errorf("decrypt notes: %w", err)
		}
		out.Notes = &notes
	}

	// --- AdditionalFields (optional) ---
	if enc.AdditionalFields != nil {
		var fields []models.CustomField
		if err := c.crypto.DecryptData(string(*enc.AdditionalFields), c.key, &fields); err != nil {
			return models.DecipheredPayload{}, fmt.Errorf("decrypt additional fields: %w", err)
		}
		out.AdditionalFields = &fields
	}

	return out, nil
}

func (c *clientCryptoService) ComputeHash(payload any) (string, error) {
	return utils.HashJSONToString(payload)
}
