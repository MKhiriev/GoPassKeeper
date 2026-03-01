// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package tui

import (
	"strings"

	"github.com/MKhiriev/go-pass-keeper/models"
)

func renderBuildInfoWindow(info models.AppBuildInfo) string {
	var b strings.Builder

	b.WriteString("Название приложения: GoPassKeeper\n")
	b.WriteString("Версия: ")
	b.WriteString(valueOrNA(info.BuildVersion()))
	b.WriteString("\n")
	b.WriteString("Дата: ")
	b.WriteString(valueOrNA(info.BuildDate()))
	b.WriteString("\n")
	b.WriteString("Коммит: ")
	b.WriteString(valueOrNA(info.BuildCommit()))

	return renderPage("ИНФОРМАЦИЯ О ПРОГРАММЕ", b.String(), "esc: назад")
}

func valueOrNA(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "N/A"
	}
	return v
}
