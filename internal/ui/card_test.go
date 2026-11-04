package ui

import (
	"strings"
	"testing"

	"github.com/espresso20/acronomicon/internal/data"
)

func TestRuneWidths(t *testing.T) {
	tests := []struct {
		str      string
		expected int
	}{
		{"hello", 5},
		{"[ K8s ]", 7},
		{"👔 Corporate", 12}, // 2 (tie) + 1 (space) + 9 (Corporate) = 12
		{"👾 Nerdism", 10},   // 2 (alien) + 1 (space) + 7 (Nerdism) = 10
		{"✦ Kubernetes", 12}, // 1 (star) + 1 (space) + 10 (Kubernetes) = 12
		{"▸ Example", 9},     // 1 (arrow) + 1 (space) + 7 (Example) = 9
		{"• Item", 6},        // 1 (bullet) + 1 (space) + 4 (Item) = 6
	}

	for _, tt := range tests {
		actual := visibleLen(tt.str)
		if actual != tt.expected {
			t.Errorf("visibleLen(%q) = %d, want %d", tt.str, actual, tt.expected)
		}
	}
}

func TestEveryCardBorderAlignment(t *testing.T) {
	const expectedWidth = 80 // 1 (left border) + 78 (box width) + 1 (right border) = 80
	all := data.All()

	for _, colorDisabled := range []bool{false, true} {
		ColorDisabled = colorDisabled

		for _, e := range all {
			rendered := RenderCard(e)
			lines := strings.Split(rendered, "\n")

			for lineNum, line := range lines {
				// Skip empty leading/trailing lines
				if line == "" {
					continue
				}

				w := visibleLen(line)
				if w != expectedWidth {
					t.Fatalf("Entry %q (ColorDisabled=%v) line %d width is %d, want %d:\nLine: %q",
						e.Acronym, colorDisabled, lineNum, w, expectedWidth, line)
				}
			}
		}
	}
}
