package autolink_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	reSSN      = `(?P<SSN>(?P<part1>\d{3})[ -]?(?P<part2>\d{2})[ -]?(?P<LastFour>[0-9]{4}))`
	replaceSSN = `XXX-XX-$LastFour`
)

func TestSocialSecurityNumberRegex(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		RE      string
		Replace string
		In      string
		Out     string
	}{
		{"SSN happy spaces", reSSN, replaceSSN, " abc 652 47 3356 def", " abc XXX-XX-3356 def"},
		{"SSN happy dashes", reSSN, replaceSSN, " abc 652-47-3356 def", " abc XXX-XX-3356 def"},
		{"SSN happy digits", reSSN, replaceSSN, " abc 652473356 def", " abc XXX-XX-3356 def"},
		{"SSN happy mixed1", reSSN, replaceSSN, " abc 65247-3356 def", " abc XXX-XX-3356 def"},
		{"SSN happy mixed2", reSSN, replaceSSN, " abc 652 47-3356 def", " abc XXX-XX-3356 def"},
		{"SSN non-match 19-09-9999", reSSN, replaceSSN, " abc 19-09-9999 def", " abc 19-09-9999 def"},
		{"SSN non-match 652_47-3356", reSSN, replaceSSN, " abc 652_47-3356 def", " abc 652_47-3356 def"},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			re := regexp.MustCompile(tc.RE)
			result := re.ReplaceAllString(tc.In, tc.Replace)
			if tc.Out != "" {
				assert.Equal(t, tc.Out, result)
			} else {
				assert.Equal(t, tc.In, result)
			}
		})
	}
}
