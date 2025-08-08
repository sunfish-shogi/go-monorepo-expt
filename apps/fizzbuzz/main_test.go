package main

import "testing"

func TestFizzBuzz(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{1, "1\n"},
		{2, "1\n2\n"},
		{3, "1\n2\nFizz\n"},
		{4, "1\n2\nFizz\n4\n"},
		{5, "1\n2\nFizz\n4\nBuzz\n"},
		{15, "1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz\n"},
	}

	for _, test := range tests {
		result := fizzBuzz(test.n)
		if result != test.expected {
			t.Errorf("fizzBuzz(%d) = %q; want %q", test.n, result, test.expected)
		}
	}
}
