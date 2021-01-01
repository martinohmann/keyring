package cmd

import (
	"strings"

	"github.com/MakeNowJust/heredoc"
)

func longDesc(s string) string {
	return trim(heredoc.Doc(s))
}

func example(s string) string {
	return indent(trim(s), "  ")
}

func indent(s, indent string) string {
	lines := strings.Split(s, "\n")

	for i, line := range lines {
		lines[i] = indent + trim(line)
	}

	return strings.Join(lines, "\n")
}

func trim(s string) string {
	return strings.TrimSpace(s)
}
