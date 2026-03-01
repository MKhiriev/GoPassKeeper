// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package models

// Metadata describes non-secret attributes of a PrivateData item.
// These fields are used for organization and presentation only.
type Metadata struct {
	// Name is the human-readable display name of the item.
	Name string

	// Folder is an optional logical container used to group items.
	Folder *string
}
