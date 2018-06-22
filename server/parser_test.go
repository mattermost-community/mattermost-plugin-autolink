package main

import (
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	var tests = []struct {
		inputMessage    string
		inputSeperator  rune
		expectedMessage []string
	}{
		{"welcome ` hello", '`', []string{"welcome ", "`", " hello"}},
		{"welcome hello", '`', []string{"welcome hello"}},
		{"`welcome ` hello", '`', []string{"`", "welcome ", "`", " hello"}},
		{"welcome ` hello`", '`', []string{"welcome ", "`", " hello", "`"}},
		{"welcome ``` hello", '`', []string{"welcome ", "`", "`", "`", " hello"}},
	}

	for _, tt := range tests {
		p := Parser{}
		actual := p.Parse(tt.inputMessage, tt.inputSeperator)
		if len(tt.expectedMessage) != len(actual) {
			t.Fatalf("parser:\n expected %v\n actual   %v", strings.Join(tt.expectedMessage, "*"), strings.Join(actual, "*"))
		}
	}
}
