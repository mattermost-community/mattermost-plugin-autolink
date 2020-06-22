package autolink_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
)

const (
	reVISA       = `(?P<VISA>(?P<part1>4\d{3})[ -]?(?P<part2>\d{4})[ -]?(?P<part3>\d{4})[ -]?(?P<LastFour>[0-9]{4}))`
	reMasterCard = `(?P<MasterCard>(?P<part1>5[1-5]\d{2})[ -]?(?P<part2>\d{4})[ -]?(?P<part3>\d{4})[ -]?(?P<LastFour>[0-9]{4}))`
	reSwitchSolo = `(?P<SwitchSolo>(?P<part1>67\d{2})[ -]?(?P<part2>\d{4})[ -]?(?P<part3>\d{4})[ -]?(?P<LastFour>[0-9]{4}))`
	reDiscover   = `(?P<Discover>(?P<part1>6011)[ -]?(?P<part2>\d{4})[ -]?(?P<part3>\d{4})[ -]?(?P<LastFour>[0-9]{4}))`
	reAMEX       = `(?P<AMEX>(?P<part1>3[47]\d{2})[ -]?(?P<part2>\d{6})[ -]?(?P<part3>\d)(?P<LastFour>[0-9]{4}))`

	replaceVISA       = "VISA XXXX-XXXX-XXXX-$LastFour"
	replaceMasterCard = "MasterCard XXXX-XXXX-XXXX-$LastFour"
	replaceSwitchSolo = "Switch/Solo XXXX-XXXX-XXXX-$LastFour"
	replaceDiscover   = "Discover XXXX-XXXX-XXXX-$LastFour"
	replaceAMEX       = "American Express XXXX-XXXXXX-X$LastFour"
)

func TestCreditCardRegex(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		RE      string
		Replace string
		In      string
		Out     string
	}{
		{"Visa happy spaces", reVISA, replaceVISA, " abc 4111 1111 1111 1234 def", " abc VISA XXXX-XXXX-XXXX-1234 def"},
		{"Visa happy dashes", reVISA, replaceVISA, "4111-1111-1111-1234", "VISA XXXX-XXXX-XXXX-1234"},
		{"Visa happy mixed", reVISA, replaceVISA, "41111111 1111-1234", "VISA XXXX-XXXX-XXXX-1234"},
		{"Visa happy digits", reVISA, replaceVISA, "abc 4111111111111234 def", "abc VISA XXXX-XXXX-XXXX-1234 def"},
		{"Visa non-match start", reVISA, replaceVISA, "3111111111111234", ""},
		{"Visa non-match num digits", reVISA, replaceVISA, " 4111-1111-1111-123", ""},
		{"Visa non-match sep", reVISA, replaceVISA, "4111=1111=1111_1234", ""},
		{"Visa non-match no break before", reVISA, replaceVISA, "abc4111-1111-1111-1234", "abcVISA XXXX-XXXX-XXXX-1234"},
		{"Visa non-match no break after", reVISA, replaceVISA, "4111-1111-1111-1234def", "VISA XXXX-XXXX-XXXX-1234def"},

		{"MasterCard happy spaces", reMasterCard, replaceMasterCard, " abc 5111 1111 1111 1234 def", " abc MasterCard XXXX-XXXX-XXXX-1234 def"},
		{"MasterCard happy dashes", reMasterCard, replaceMasterCard, "5211-1111-1111-1234", "MasterCard XXXX-XXXX-XXXX-1234"},
		{"MasterCard happy mixed", reMasterCard, replaceMasterCard, "53111111 1111-1234", "MasterCard XXXX-XXXX-XXXX-1234"},
		{"MasterCard happy digits", reMasterCard, replaceMasterCard, "abc 5411111111111234 def", "abc MasterCard XXXX-XXXX-XXXX-1234 def"},
		{"MasterCard non-match start", reMasterCard, replaceMasterCard, "3111111111111234", ""},
		{"MasterCard non-match num digits", reMasterCard, replaceMasterCard, " 5111-1111-1111-123", ""},
		{"MasterCard non-match sep", reMasterCard, replaceMasterCard, "5111=1111=1111_1234", ""},
		{"MasterCard non-match no break before", reMasterCard, replaceMasterCard, "abc5511-1111-1111-1234", "abcMasterCard XXXX-XXXX-XXXX-1234"},
		{"MasterCard non-match no break after", reMasterCard, replaceMasterCard, "5111-1111-1111-1234def", "MasterCard XXXX-XXXX-XXXX-1234def"},

		{"SwitchSolo happy spaces", reSwitchSolo, replaceSwitchSolo, " abc 6711 1111 1111 1234 def", " abc Switch/Solo XXXX-XXXX-XXXX-1234 def"},
		{"SwitchSolo happy dashes", reSwitchSolo, replaceSwitchSolo, "6711-1111-1111-1234", "Switch/Solo XXXX-XXXX-XXXX-1234"},
		{"SwitchSolo happy mixed", reSwitchSolo, replaceSwitchSolo, "67111111 1111-1234", "Switch/Solo XXXX-XXXX-XXXX-1234"},
		{"SwitchSolo happy digits", reSwitchSolo, replaceSwitchSolo, "abc 6711111111111234 def", "abc Switch/Solo XXXX-XXXX-XXXX-1234 def"},
		{"SwitchSolo non-match start", reSwitchSolo, replaceSwitchSolo, "3111111111111234", ""},
		{"SwitchSolo non-match num digits", reSwitchSolo, replaceSwitchSolo, " 6711-1111-1111-123", ""},
		{"SwitchSolo non-match sep", reSwitchSolo, replaceSwitchSolo, "6711=1111=1111_1234", ""},
		{"SwitchSolo non-match no break before", reSwitchSolo, replaceSwitchSolo, "abc6711-1111-1111-1234", "abcSwitch/Solo XXXX-XXXX-XXXX-1234"},
		{"SwitchSolo non-match no break after", reSwitchSolo, replaceSwitchSolo, "6711-1111-1111-1234def", "Switch/Solo XXXX-XXXX-XXXX-1234def"},

		{"Discover happy spaces", reDiscover, replaceDiscover, " abc 6011 1111 1111 1234 def", " abc Discover XXXX-XXXX-XXXX-1234 def"},
		{"Discover happy dashes", reDiscover, replaceDiscover, "6011-1111-1111-1234", "Discover XXXX-XXXX-XXXX-1234"},
		{"Discover happy mixed", reDiscover, replaceDiscover, "60111111 1111-1234", "Discover XXXX-XXXX-XXXX-1234"},
		{"Discover happy digits", reDiscover, replaceDiscover, "abc 6011111111111234 def", "abc Discover XXXX-XXXX-XXXX-1234 def"},
		{"Discover non-match start", reDiscover, replaceDiscover, "3111111111111234", ""},
		{"Discover non-match num digits", reDiscover, replaceDiscover, " 6011-1111-1111-123", ""},
		{"Discover non-match sep", reDiscover, replaceDiscover, "6011=1111=1111_1234", ""},
		{"Discover non-match no break before", reDiscover, replaceDiscover, "abc6011-1111-1111-1234", "abcDiscover XXXX-XXXX-XXXX-1234"},
		{"Discover non-match no break after", reDiscover, replaceDiscover, "6011-1111-1111-1234def", "Discover XXXX-XXXX-XXXX-1234def"},

		{"AMEX happy spaces", reAMEX, replaceAMEX, " abc 3411 123456 12345 def", " abc American Express XXXX-XXXXXX-X2345 def"},
		{"AMEX happy dashes", reAMEX, replaceAMEX, "3711-123456-12345", "American Express XXXX-XXXXXX-X2345"},
		{"AMEX happy mixed", reAMEX, replaceAMEX, "3411-123456 12345", "American Express XXXX-XXXXXX-X2345"},
		{"AMEX happy digits", reAMEX, replaceAMEX, "abc 371112345612345 def", "abc American Express XXXX-XXXXXX-X2345 def"},
		{"AMEX non-match start 41", reAMEX, replaceAMEX, "411112345612345", ""},
		{"AMEX non-match start 31", reAMEX, replaceAMEX, "3111111111111234", ""},
		{"AMEX non-match num digits", reAMEX, replaceAMEX, " 4111-1111-1111-123", ""},
		{"AMEX non-match sep", reAMEX, replaceAMEX, "4111-1111=1111-1234", ""},
		{"AMEX non-match no break before", reAMEX, replaceAMEX, "abc3711-123456-12345", "abcAmerican Express XXXX-XXXXXX-X2345"},
		{"AMEX non-match no break after", reAMEX, replaceAMEX, "3711-123456-12345def", "American Express XXXX-XXXXXX-X2345def"},
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

func TestCreditCard(t *testing.T) {
	var tests = []struct {
		Name            string
		Link            autolink.Autolink
		inputMessage    string
		expectedMessage string
	}{
		{
			"VISA happy",
			autolink.Autolink{
				Pattern:  reVISA,
				Template: replaceVISA,
			},
			"A credit card 4111-1111-2222-1234 mentioned",
			"A credit card VISA XXXX-XXXX-XXXX-1234 mentioned",
		}, {
			"VISA",
			autolink.Autolink{
				Pattern:              reVISA,
				Template:             replaceVISA,
				DisableNonWordPrefix: true,
				DisableNonWordSuffix: true,
			},
			"A credit card4111-1111-2222-3333mentioned",
			"A credit cardVISA XXXX-XXXX-XXXX-3333mentioned",
		}, {
			"Multiple VISA replacements",
			autolink.Autolink{
				Pattern:  reVISA,
				Template: replaceVISA,
			},
			"Credit cards 4111-1111-2222-3333 4222-3333-4444-5678 mentioned",
			"Credit cards VISA XXXX-XXXX-XXXX-3333 VISA XXXX-XXXX-XXXX-5678 mentioned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := tt.Link.Compile()
			actual := tt.Link.Replace(tt.inputMessage)

			assert.Equal(t, tt.expectedMessage, actual)
			assert.NoError(t, err)
		})
	}
}
