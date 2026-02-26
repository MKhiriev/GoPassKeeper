package tui

import (
	"fmt"
	"strings"
)

const uiDivider = "──────────────────────────────────────────────────────"

func viewTitle(title string) string {
	return fmt.Sprintf("%s\n%s\n", title, uiDivider)
}

func renderPage(title, data, hotKeys string) string {
	var b strings.Builder

	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(uiDivider)
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
	b.WriteString(uiDivider)
	b.WriteString("\n")

	if strings.TrimSpace(hotKeys) != "" {
		b.WriteString("  ")
		b.WriteString(hotKeys)
		b.WriteString("\n")
	}
	b.WriteString("  ctrl+c: выход")

	return b.String()
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
