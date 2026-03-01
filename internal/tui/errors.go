// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package tui

import "strings"

func humanizeServerUnavailableError(err error) string {
	if err == nil {
		return ""
	}

	s := strings.ToLower(err.Error())
	if strings.Contains(s, "connection refused") ||
		strings.Contains(s, "dial tcp") ||
		strings.Contains(s, "no such host") ||
		strings.Contains(s, "network is unreachable") ||
		strings.Contains(s, "i/o timeout") ||
		strings.Contains(s, "context deadline exceeded") {
		return "Отсутствует сеть или Сервер недоступен"
	}

	return err.Error()
}
