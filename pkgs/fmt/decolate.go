package fmt

import "strings"

func Highlight(text string) string {
	line := strings.Split(text, "\n")
	if len(line) == 0 {
		return ""
	}
	var maxLen int
	for _, l := range line {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}
	var result strings.Builder
	result.WriteString(" " + strings.Repeat("^", maxLen+2) + " \n")
	for _, l := range line {
		result.WriteString("< " + l + strings.Repeat(" ", maxLen-len(l)) + " >\n")
	}
	result.WriteString(" " + strings.Repeat("v", maxLen+2) + " \n")
	return result.String()
}
