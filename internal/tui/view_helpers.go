// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const minDividerWidth = 54

func renderPage(title, data, hotKeys string) string {
	var b strings.Builder
	divider := strings.Repeat("─", pageContentWidth(title, data, hotKeys))

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(divider)
	b.WriteString("\n\n")

	if strings.TrimSpace(data) != "" {
		lines := strings.Split(data, "\n")
		for _, line := range lines {
			b.WriteString("  ")
			b.WriteString(line)
			b.WriteString("\n")
		}
	} else {
		b.WriteString("  -\n")
	}

	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(divider)
	b.WriteString("\n")

	if strings.TrimSpace(hotKeys) != "" {
		b.WriteString("  ")
		b.WriteString(hotKeys)
		b.WriteString("\n")
	}
	b.WriteString("  ctrl+c: выход")

	return b.String()
}

func pageContentWidth(title, data, hotKeys string) int {
	width := minDividerWidth

	width = max(width, lipgloss.Width(title))
	width = max(width, maxLineWidth(data))
	width = max(width, maxLineWidth(hotKeys))
	width = max(width, lipgloss.Width("ctrl+c: выход"))

	return width
}

func maxLineWidth(block string) int {
	if strings.TrimSpace(block) == "" {
		return 0
	}

	maxWidth := 0
	for _, line := range strings.Split(block, "\n") {
		maxWidth = max(maxWidth, lipgloss.Width(line))
	}
	return maxWidth
}

func valueOrDash(v *string) string {
	if v == nil || *v == "" {
		return "-"
	}
	return *v
}

func fitText(v string, max int) string {
	if max <= 0 || len(v) <= max {
		return v
	}
	if max <= 3 {
		return v[:max]
	}
	return v[:max-3] + "..."
}
