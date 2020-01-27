package autolink_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
	"github.com/mattermost/mattermost-plugin-autolink/server/autolinkplugin"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupTestPlugin(t *testing.T, l autolink.Autolink) *autolinkplugin.Plugin {
	p := &autolinkplugin.Plugin{}
	api := &plugintest.API{}

	api.On("GetChannel", mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
		api.On("GetTeam", mock.AnythingOfType("string")).Return(&model.Team{}, (*model.AppError)(nil))
	}).Return(&model.Channel{
		Id:     "thechannel_id0123456789012",
		TeamId: "theteam_id0123456789012345",
		Name:   "thechannel_name",
	}, (*model.AppError)(nil))
	p.SetAPI(api)

	err := l.Compile()
	require.Nil(t, err)
	p.UpdateConfig(func(conf *autolinkplugin.Config) {
		conf.Links = []autolink.Autolink{l}
	})
	return p
}

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

func TestCCRegex(t *testing.T) {
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

const (
	reSSN      = `(?P<SSN>(?P<part1>\d{3})[ -]?(?P<part2>\d{2})[ -]?(?P<LastFour>[0-9]{4}))`
	replaceSSN = `XXX-XX-$LastFour`
)

func TestSSNRegex(t *testing.T) {
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
			_ = tt.Link.Compile()
			actual := tt.Link.Replace(tt.inputMessage)
			assert.Equal(t, tt.expectedMessage, actual)
		})
	}
}

func TestLink(t *testing.T) {
	for _, tc := range []struct {
		Name            string
		Link            autolink.Autolink
		Message         string
		ExpectedMessage string
	}{
		{
			"Simple pattern",
			autolink.Autolink{
				Pattern:  "(Mattermost)",
				Template: "[Mattermost](https://mattermost.com)",
			},
			"Welcome to Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com)!",
		}, {
			"Pattern with variable name accessed using $variable",
			autolink.Autolink{
				Pattern:  "(?P<key>Mattermost)",
				Template: "[$key](https://mattermost.com)",
			},
			"Welcome to Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com)!",
		}, {
			"Multiple replacments",
			autolink.Autolink{
				Pattern:  "(?P<key>Mattermost)",
				Template: "[$key](https://mattermost.com)",
			},
			"Welcome to Mattermost and have fun with Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com) and have fun with [Mattermost](https://mattermost.com)!",
		}, {
			"Pattern with variable name accessed using ${variable}",
			autolink.Autolink{
				Pattern:  "(?P<key>Mattermost)",
				Template: "[${key}](https://mattermost.com)",
			},
			"Welcome to Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com)!",
		}, {
			"Jira example",
			autolink.Autolink{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		}, {
			"Jira example 2 (within a ())",
			autolink.Autolink{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Link in brackets should link (see MM-12345)",
			"Link in brackets should link (see [MM-12345](https://mattermost.atlassian.net/browse/MM-12345))",
		}, {
			"Jira example 3 (before ,)",
			autolink.Autolink{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Link a ticket MM-12345, before a comma",
			"Link a ticket [MM-12345](https://mattermost.atlassian.net/browse/MM-12345), before a comma",
		}, {
			"Jira example 3 (at begin of the message)",
			autolink.Autolink{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"MM-12345 should link!",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		}, {
			"Pattern word prefix and suffix disabled",
			autolink.Autolink{
				Pattern:              "(?P<previous>^|\\s)(MM)(-)(?P<jira_id>\\d+)",
				Template:             "${previous}[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
				DisableNonWordPrefix: true,
				DisableNonWordSuffix: true,
			},
			"Welcome MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		}, {
			"Pattern word prefix and suffix disabled (at begin of the message)",
			autolink.Autolink{
				Pattern:              "(?P<previous>^|\\s)(MM)(-)(?P<jira_id>\\d+)",
				Template:             "${previous}[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
				DisableNonWordPrefix: true,
				DisableNonWordSuffix: true,
			},
			"MM-12345 should link!",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		}, {
			"Pattern word prefix and suffix enable (in the middle of other text)",
			autolink.Autolink{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"WelcomeMM-12345should not link!",
			"WelcomeMM-12345should not link!",
		}, {
			"Pattern word prefix and suffix disabled (in the middle of other text)",
			autolink.Autolink{
				Pattern:              "(MM)(-)(?P<jira_id>\\d+)",
				Template:             "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
				DisableNonWordPrefix: true,
				DisableNonWordSuffix: true,
			},
			"WelcomeMM-12345should link!",
			"Welcome[MM-12345](https://mattermost.atlassian.net/browse/MM-12345)should link!",
		}, {
			"Not relinking",
			autolink.Autolink{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should not re-link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should not re-link!",
		}, {
			"Url replacement",
			autolink.Autolink{
				Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		}, {
			"Url replacement multiple times",
			autolink.Autolink{
				Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345. should link https://mattermost.atlassian.net/browse/MM-12346 !",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345). should link [MM-12346](https://mattermost.atlassian.net/browse/MM-12346) !",
		}, {
			"Url replacement multiple times and at beginning",
			autolink.Autolink{
				Pattern:  "(https:\\/\\/mattermost.atlassian.net\\/browse\\/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"https://mattermost.atlassian.net/browse/MM-12345 https://mattermost.atlassian.net/browse/MM-12345",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		}, {
			"Url replacement at end",
			autolink.Autolink{
				Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		},
	} {

		t.Run(tc.Name, func(t *testing.T) {
			p := setupTestPlugin(t, tc.Link)
			post, _ := p.MessageWillBePosted(nil, &model.Post{
				Message: tc.Message,
			})

			assert.Equal(t, tc.ExpectedMessage, post.Message)
		})
	}
}

func TestLegacyWordBoundaries(t *testing.T) {
	const pattern = "(KEY)(-)(?P<ID>\\d+)"
	const template = "[KEY-$ID](someurl/KEY-$ID)"
	const ref = "KEY-12345"
	const ID = "12345"
	const markdown = "[KEY-12345](someurl/KEY-12345)"

	var defaultLink = autolink.Autolink{
		Pattern:  pattern,
		Template: template,
	}

	linkNoPrefix := defaultLink
	linkNoPrefix.DisableNonWordPrefix = true

	linkNoSuffix := defaultLink
	linkNoSuffix.DisableNonWordSuffix = true

	linkNoPrefixNoSuffix := defaultLink
	linkNoPrefixNoSuffix.DisableNonWordSuffix = true
	linkNoPrefixNoSuffix.DisableNonWordPrefix = true

	for _, tc := range []struct {
		Name       string
		Sep        string
		Link       autolink.Autolink
		Prefix     string
		Suffix     string
		ExpectFail bool
	}{
		{Name: "space both sides both breaks required", Prefix: " ", Suffix: " "},
		{Name: "space both sides left break not required", Prefix: " ", Suffix: " ", Link: linkNoPrefix},
		{Name: "space both sides right break not required", Prefix: " ", Suffix: " ", Link: linkNoSuffix},
		{Name: "space both sides neither break required", Prefix: " ", Suffix: " ", Link: linkNoPrefixNoSuffix},

		{Name: "space left side both breaks required", Prefix: " ", ExpectFail: true},
		{Name: "space left side left break not required", Prefix: " ", Link: linkNoPrefix, ExpectFail: true},
		{Name: "space left side right break not required", Prefix: " ", Link: linkNoSuffix},
		{Name: "space left side neither break required", Prefix: " ", Link: linkNoPrefixNoSuffix},

		{Name: "space right side both breaks required", Suffix: " ", ExpectFail: true},
		{Name: "space right side left break not required", Suffix: " ", Link: linkNoPrefix},
		{Name: "space right side right break not required", Suffix: " ", Link: linkNoSuffix, ExpectFail: true},
		{Name: "space right side neither break required", Prefix: " ", Link: linkNoPrefixNoSuffix},

		{Name: "none both breaks required", ExpectFail: true},
		{Name: "none left break not required", Link: linkNoPrefix, ExpectFail: true},
		{Name: "none right break not required", Link: linkNoSuffix, ExpectFail: true},
		{Name: "none neither break required", Link: linkNoPrefixNoSuffix},

		// '(', '[' are not start separators
		{Sep: "paren", Name: "2 parens", Prefix: "(", Suffix: ")", ExpectFail: true},
		{Sep: "paren", Name: "2 parens no suffix", Prefix: "(", Suffix: ")", Link: linkNoSuffix, ExpectFail: true},
		{Sep: "paren", Name: "left paren", Prefix: "(", Link: linkNoSuffix, ExpectFail: true},
		{Sep: "sbracket", Name: "2 brackets", Prefix: "[", Suffix: "]", ExpectFail: true},
		{Sep: "lsbracket", Name: "bracket no prefix", Prefix: "[", Link: linkNoPrefix, ExpectFail: true},
		{Sep: "lsbracket", Name: "both breaks", Prefix: "[", ExpectFail: true},
		{Sep: "lsbracket", Name: "bracket no suffix", Prefix: "[", Link: linkNoSuffix, ExpectFail: true},

		// ']' is not a finish separator
		{Sep: "rsbracket", Name: "bracket", Suffix: "]", Link: linkNoPrefix, ExpectFail: true},

		{Sep: "paren", Name: "2 parens no prefix", Prefix: "(", Suffix: ")", Link: linkNoPrefix},
		{Sep: "paren", Name: "right paren", Suffix: ")", Link: linkNoPrefix},
		{Sep: "lsbracket", Name: "bracket neither prefix suffix", Prefix: "[", Link: linkNoPrefixNoSuffix},
		{Sep: "rand", Name: "random separators", Prefix: "%() ", Suffix: "?! $%^&"},
	} {

		orig := fmt.Sprintf("word1%s%s%sword2", tc.Prefix, ref, tc.Suffix)
		expected := fmt.Sprintf("word1%s%s%sword2", tc.Prefix, markdown, tc.Suffix)

		pref := tc.Prefix
		suff := tc.Suffix
		if tc.Sep != "" {
			pref = "_" + tc.Sep + "_"
			suff = "_" + tc.Sep + "_"
		}
		name := fmt.Sprintf("word1%s%s%sword2", pref, ref, suff)
		if tc.Name != "" {
			name = tc.Name + " " + name
		}

		t.Run(name, func(t *testing.T) {
			l := tc.Link
			if l.Pattern == "" {
				l = defaultLink
			}
			p := setupTestPlugin(t, l)

			post, _ := p.MessageWillBePosted(nil, &model.Post{
				Message: orig,
			})
			if tc.ExpectFail {
				assert.Equal(t, orig, post.Message)
				return
			}
			assert.Equal(t, expected, post.Message)
		})
	}
}

func TestWordMatch(t *testing.T) {
	const pattern = "(KEY)(-)(?P<ID>\\d+)"
	const template = "[KEY-$ID](someurl/KEY-$ID)"
	const ref = "KEY-12345"
	const ID = "12345"
	const markdown = "[KEY-12345](someurl/KEY-12345)"

	var defaultLink = autolink.Autolink{
		Pattern:   pattern,
		Template:  template,
		WordMatch: true,
	}

	linkNoPrefix := defaultLink
	linkNoPrefix.DisableNonWordPrefix = true

	linkNoSuffix := defaultLink
	linkNoSuffix.DisableNonWordSuffix = true

	linkNoPrefixNoSuffix := defaultLink
	linkNoPrefixNoSuffix.DisableNonWordSuffix = true
	linkNoPrefixNoSuffix.DisableNonWordPrefix = true

	for _, tc := range []struct {
		Name       string
		Sep        string
		Link       autolink.Autolink
		Prefix     string
		Suffix     string
		ExpectFail bool
	}{
		{Name: "space both sides both breaks required", Prefix: " ", Suffix: " "},
		{Name: "space both sides left break not required", Prefix: " ", Suffix: " ", Link: linkNoPrefix},
		{Name: "space both sides right break not required", Prefix: " ", Suffix: " ", Link: linkNoSuffix},
		{Name: "space both sides neither break required", Prefix: " ", Suffix: " ", Link: linkNoPrefixNoSuffix},

		{Name: "space left side both breaks required", Prefix: " ", ExpectFail: true},
		{Name: "space left side left break not required", Prefix: " ", Link: linkNoPrefix, ExpectFail: true},
		{Name: "space left side right break not required", Prefix: " ", Link: linkNoSuffix},
		{Name: "space left side neither break required", Prefix: " ", Link: linkNoPrefixNoSuffix},

		{Name: "space right side both breaks required", Suffix: " ", ExpectFail: true},
		{Name: "space right side left break not required", Suffix: " ", Link: linkNoPrefix},
		{Name: "space right side right break not required", Suffix: " ", Link: linkNoSuffix, ExpectFail: true},
		{Name: "space right side neither break required", Prefix: " ", Link: linkNoPrefixNoSuffix},

		{Name: "none both breaks required", ExpectFail: true},
		{Name: "none left break not required", Link: linkNoPrefix, ExpectFail: true},
		{Name: "none right break not required", Link: linkNoSuffix, ExpectFail: true},
		{Name: "none neither break required", Link: linkNoPrefixNoSuffix},

		{Sep: "paren", Name: "2 parens", Prefix: "(", Suffix: ")"},
		{Sep: "paren", Name: "left paren", Prefix: "(", Link: linkNoSuffix},
		{Sep: "paren", Name: "right paren", Suffix: ")", Link: linkNoPrefix},
		{Sep: "sbracket", Name: "2 brackets", Prefix: "[", Suffix: "]"},
		{Sep: "lsbracket", Name: "both breaks", Prefix: "[", ExpectFail: true},
		{Sep: "lsbracket", Name: "bracket no prefix", Prefix: "[", Link: linkNoPrefix, ExpectFail: true},
		{Sep: "lsbracket", Name: "bracket no suffix", Prefix: "[", Link: linkNoSuffix},
		{Sep: "lsbracket", Name: "bracket neither prefix suffix", Prefix: "[", Link: linkNoPrefixNoSuffix},
		{Sep: "rsbracket", Name: "bracket", Suffix: "]", Link: linkNoPrefix},
		{Sep: "rand", Name: "random separators", Prefix: "% (", Suffix: "-- $%^&"},
	} {

		orig := fmt.Sprintf("word1%s%s%sword2", tc.Prefix, ref, tc.Suffix)
		expected := fmt.Sprintf("word1%s%s%sword2", tc.Prefix, markdown, tc.Suffix)

		pref := tc.Prefix
		suff := tc.Suffix
		if tc.Sep != "" {
			pref = "_" + tc.Sep + "_"
			suff = "_" + tc.Sep + "_"
		}
		name := fmt.Sprintf("word1%s%s%sword2", pref, ref, suff)
		if tc.Name != "" {
			name = tc.Name + " " + name
		}

		t.Run(name, func(t *testing.T) {
			l := tc.Link
			if l.Pattern == "" {
				l = defaultLink
			}
			p := setupTestPlugin(t, l)

			post, _ := p.MessageWillBePosted(nil, &model.Post{
				Message: orig,
			})
			if tc.ExpectFail {
				assert.Equal(t, orig, post.Message)
				return
			}
			assert.Equal(t, expected, post.Message)
		})
	}
}
