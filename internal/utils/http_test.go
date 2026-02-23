package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON_Success(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}

	n, err := WriteJSON(w, data, http.StatusOK)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes written")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
	}

	expected, _ := json.Marshal(data)
	if w.Body.String() != string(expected) {
		t.Errorf("expected body %s, got %s", expected, w.Body.String())
	}
}

func TestWriteJSON_CustomStatusCode(t *testing.T) {
	w := httptest.NewRecorder()

	_, err := WriteJSON(w, map[string]string{"error": "not found"}, http.StatusNotFound)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestWriteJSON_InvalidData(t *testing.T) {
	w := httptest.NewRecorder()

	// channels cannot be marshaled to JSON
	_, err := WriteJSON(w, make(chan int), http.StatusOK)

	if err == nil {
		t.Fatal("expected error for non-serializable data, got nil")
	}
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestWriteJSON_NilData(t *testing.T) {
	w := httptest.NewRecorder()

	_, err := WriteJSON(w, nil, http.StatusOK)

	if err != nil {
		t.Fatalf("expected no error for nil data, got: %v", err)
	}
	if w.Body.String() != "null" {
		t.Errorf("expected body 'null', got '%s'", w.Body.String())
	}
}

func TestWriteJSON_EmptyStruct(t *testing.T) {
	w := httptest.NewRecorder()

	_, err := WriteJSON(w, struct{}{}, http.StatusOK)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if w.Body.String() != "{}" {
		t.Errorf("expected body '{}', got '%s'", w.Body.String())
	}
}

func TestWriteJSON_Slice(t *testing.T) {
	w := httptest.NewRecorder()
	data := []int{1, 2, 3}

	_, err := WriteJSON(w, data, http.StatusOK)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expected, _ := json.Marshal(data)
	if w.Body.String() != string(expected) {
		t.Errorf("expected body %s, got %s", expected, w.Body.String())
	}
}

func TestWriteJSON_NestedStruct(t *testing.T) {
	type Address struct {
		City string `json:"city"`
	}
	type User struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	w := httptest.NewRecorder()
	data := User{Name: "Alice", Age: 30, Address: Address{City: "Tashkent"}}

	_, err := WriteJSON(w, data, http.StatusCreated)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	expected, _ := json.Marshal(data)
	if w.Body.String() != string(expected) {
		t.Errorf("expected body %s, got %s", expected, w.Body.String())
	}
}
