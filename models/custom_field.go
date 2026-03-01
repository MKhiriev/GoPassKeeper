// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package models

// CustomField represents a user-defined field attached to PrivateData.
// Each custom field has its own semantic type and encrypted value.
type CustomField struct {
	// Type defines the data type of the custom field.
	Type DataType `json:"type"`

	// Data contains the encrypted value of the custom field.
	Data CipheredData `json:"data"`
}
