package fmt

import "testing"

func TestHighlight(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "Hello, World!",
			expected: "*****************\n* Hello, World! *\n*****************\n",
		},
		{
			input:    "This is a test.\nWith multiple lines.",
			expected: "************************\n* This is a test.      *\n* With multiple lines. *\n************************\n",
		},
		{
			input:    "Short\nLonger line here.",
			expected: "*********************\n* Short             *\n* Longer line here. *\n*********************\n",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := Highlight(tc.input)
			if result != tc.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tc.expected, result)
			}
		})
	}
}
