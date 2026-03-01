// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package models

type (
	// CipheredData holds the primary encrypted payload of a vault item
	// (e.g. login credentials, card details, binary reference).
	// Opaque to the server — only the owning client can decrypt it.
	CipheredData string

	// CipheredMetadata holds encrypted non-sensitive descriptive information
	// such as display name and folder placement.
	// Opaque to the server — only the owning client can decrypt it.
	CipheredMetadata string

	// CipheredNotes holds encrypted optional free-form user notes
	// attached to a vault item.
	// Opaque to the server — only the owning client can decrypt it.
	CipheredNotes string

	// CipheredCustomFields holds encrypted user-defined key-value fields
	// attached to a vault item. Each field is independently typed.
	// Opaque to the server — only the owning client can decrypt it.
	CipheredCustomFields string
)
