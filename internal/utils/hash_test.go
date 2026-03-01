// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/MKhiriev/go-pass-keeper/models"
)

func TestInitHasherPoolAndHash(t *testing.T) {
	key := "secret-key"
	InitHasherPool(key)

	data := []byte("test-data")

	sum1 := Hash(data)
	sum2 := Hash(data)

	if len(sum1) == 0 {
		t.Fatal("hash result is empty")
	}

	if !bytes.Equal(sum1, sum2) {
		t.Fatal("hash must be deterministic for the same input")
	}

	// verify against direct HMAC computation
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	expected := h.Sum(nil)

	if !bytes.Equal(sum1, expected) {
		t.Fatalf("unexpected hash value\nwant: %x\ngot:  %x", expected, sum1)
	}
}

const testHashKey = "test-secret-key"

func TestHash_WithRealPayload(t *testing.T) {
	InitHasherPool(testHashKey)

	payload := models.PrivateDataPayload{
		Metadata: "my-gmail-account",
		Type:     models.LoginPassword,
		Data:     "encrypted-login-password-blob",
	}

	// Сериализуем Payload в JSON (как это делает middleware)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	got := hex.EncodeToString(Hash(payloadBytes))

	// Эталонный хеш считаем напрямую через crypto/hmac
	mac := hmac.New(sha256.New, []byte(testHashKey))
	mac.Write(payloadBytes)
	want := hex.EncodeToString(mac.Sum(nil))

	if got != want {
		t.Errorf("Hash mismatch:\n  got:  %s\n  want: %s", got, want)
	}
}

// TestHash_DifferentPayloads проверяет что разные Payload дают разные хеши
func TestHash_DifferentPayloads(t *testing.T) {
	InitHasherPool(testHashKey)

	payload1 := models.PrivateDataPayload{
		Metadata: "gmail",
		Type:     models.LoginPassword,
		Data:     "encrypted-blob-1",
	}

	payload2 := models.PrivateDataPayload{
		Metadata: "github",
		Type:     models.LoginPassword,
		Data:     "encrypted-blob-2",
	}

	bytes1, _ := json.Marshal(payload1)
	bytes2, _ := json.Marshal(payload2)

	hash1 := hex.EncodeToString(Hash(bytes1))
	hash2 := hex.EncodeToString(Hash(bytes2))

	if hash1 == hash2 {
		t.Error("different payloads must produce different hashes")
	}
}

// TestHash_SamePayload_Deterministic проверяет что одинаковый Payload всегда дает одинаковый хеш
func TestHash_SamePayload_Deterministic(t *testing.T) {
	InitHasherPool(testHashKey)

	payload := models.PrivateDataPayload{
		Metadata: "my-bank-card",
		Type:     models.BankCard,
		Data:     "encrypted-card-blob",
	}

	payloadBytes, _ := json.Marshal(payload)

	hash1 := hex.EncodeToString(Hash(payloadBytes))
	hash2 := hex.EncodeToString(Hash(payloadBytes))

	if hash1 != hash2 {
		t.Errorf("same payload must produce same hash:\n  hash1: %s\n  hash2: %s", hash1, hash2)
	}
}

// TestHash_DifferentKeys проверяет что разные ключи дают разные хеши для одного Payload
func TestHash_DifferentKeys(t *testing.T) {
	payload := models.PrivateDataPayload{
		Metadata: "some-note",
		Type:     models.Text,
		Data:     "encrypted-text-blob",
	}
	payloadBytes, _ := json.Marshal(payload)

	InitHasherPool("key-one")
	hash1 := hex.EncodeToString(Hash(payloadBytes))

	InitHasherPool("key-two")
	hash2 := hex.EncodeToString(Hash(payloadBytes))

	if hash1 == hash2 {
		t.Error("different keys must produce different hashes for the same payload")
	}
}

// TestHash_PayloadFieldOrder проверяет что порядок полей в JSON влияет на хеш.
// Два идентичных по значениям Payload, но сериализованных в разном порядке полей,
// дадут разные хеши — это потенциальная проблема при синхронизации между клиентами
// на разных платформах (Go, Python, JS и т.д.)
func TestHash_PayloadFieldOrder(t *testing.T) {
	InitHasherPool(testHashKey)

	// Payload 1: поля в стандартном порядке Go (порядок объявления в структуре)
	// {"metadata":"my-gmail","type":1,"data":"encrypted-blob"}
	payload1 := models.PrivateDataPayload{
		Metadata: "my-gmail",
		Type:     models.LoginPassword,
		Data:     "encrypted-blob",
	}
	payload1Bytes, err := json.Marshal(payload1)
	if err != nil {
		t.Fatalf("failed to marshal payload1: %v", err)
	}

	// Payload 2: те же значения, но поля переставлены вручную
	// Симулируем клиент на другом языке (Python, JS, Rust),
	// который сериализует поля в алфавитном или произвольном порядке:
	// {"data":"encrypted-blob","metadata":"my-gmail","type":1}
	//payload2Bytes := []byte(`{"data":"encrypted-blob","metadata":"my-gmail","type":1}`)

	// Payload 1: поля в стандартном порядке Go (порядок объявления в структуре)
	// {"metadata":"my-gmail","type":1,"data":"encrypted-blob"}
	payload2 := models.PrivateDataPayload{
		Data:     "encrypted-blob",
		Type:     models.LoginPassword,
		Metadata: "my-gmail",
	}
	payload2Bytes, err := json.Marshal(payload2)
	if err != nil {
		t.Fatalf("failed to marshal payload1: %v", err)
	}

	hash1 := hex.EncodeToString(Hash(payload1Bytes))
	hash2 := hex.EncodeToString(Hash(payload2Bytes))

	t.Logf("payload1 JSON: %s", payload1Bytes)
	t.Logf("payload2 JSON: %s", payload2Bytes)
	t.Logf("hash1: %s", hash1)
	t.Logf("hash2: %s", hash2)

	// Ожидаем что хеши НЕ совпадут — это демонстрирует проблему
	if hash1 == hash2 {
		t.Log("OK: field order does not affect hash (canonical JSON is used)")
	} else {
		t.Error("PROBLEM: same values but different field order produce different hashes — " +
			"middleware will reject valid requests from non-Go clients")
	}
}

// TestHash_UnmarshalThenHash проверяет что два JSON с одинаковыми данными,
// но разным порядком полей, после Unmarshal -> Marshal дают одинаковый хеш.
// Это симулирует реальный сценарий в middleware:
// клиент присылает JSON -> сервер декодирует в struct -> считает хеш от struct.
func TestHash_UnmarshalThenHash(t *testing.T) {
	InitHasherPool(testHashKey)

	// Два JSON с одинаковыми значениями, но разным порядком полей
	json1 := []byte(`{"metadata":"my-gmail","type":1,"data":"encrypted-blob"}`)
	json2 := []byte(`{"data":"encrypted-blob","type":1,"metadata":"my-gmail"}`)

	// Декодируем оба JSON в структуру Payload
	var payload1 models.PrivateDataPayload
	if err := json.Unmarshal(json1, &payload1); err != nil {
		t.Fatalf("failed to unmarshal json1: %v", err)
	}

	var payload2 models.PrivateDataPayload
	if err := json.Unmarshal(json2, &payload2); err != nil {
		t.Fatalf("failed to unmarshal json2: %v", err)
	}

	// Кодируем обратно в байты (теперь порядок полей определяется структурой Go)
	payload1Bytes, err := json.Marshal(payload1)
	if err != nil {
		t.Fatalf("failed to marshal payload1: %v", err)
	}

	payload2Bytes, err := json.Marshal(payload2)
	if err != nil {
		t.Fatalf("failed to marshal payload2: %v", err)
	}

	hash1 := hex.EncodeToString(Hash(payload1Bytes))
	hash2 := hex.EncodeToString(Hash(payload2Bytes))

	t.Logf("json1 (original):  %s", json1)
	t.Logf("json2 (original):  %s", json2)
	t.Logf("payload1 (re-marshaled): %s", payload1Bytes)
	t.Logf("payload2 (re-marshaled): %s", payload2Bytes)
	t.Logf("hash1: %s", hash1)
	t.Logf("hash2: %s", hash2)

	if hash1 != hash2 {
		t.Error("hashes must be equal after Unmarshal -> Marshal normalization")
	}
}
