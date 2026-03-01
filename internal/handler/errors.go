// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package handler

import "errors"

// errNoHandlersAreCreated is returned by NewHandlers when neither an HTTP nor
// a gRPC address is provided in the server configuration, resulting in no
// transport handlers being initialized. This is treated as a fatal
// misconfiguration and causes the application to fail at startup.
var errNoHandlersAreCreated = errors.New("no handlers are created")
