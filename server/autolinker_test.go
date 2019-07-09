package main

import (
	"fmt"
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupTestPlugin(t *testing.T, link Link) *Plugin {
	p := &Plugin{}
	api := &plugintest.API{}

	api.On("GetChannel", mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
		api.On("GetTeam", mock.AnythingOfType("string")).Return(&model.Team{}, (*model.AppError)(nil))
	}).Return(&model.Channel{
		Id:     "thechannel_id0123456789012",
		TeamId: "theteam_id0123456789012345",
		Name:   "thechannel_name",
	}, (*model.AppError)(nil))
	p.SetAPI(api)

	al, err := NewAutoLinker(link)
	require.Nil(t, err)
	p.links.Store([]*AutoLinker{al})
	return p
}

func TestAutolink(t *testing.T) {
	for _, tc := range []struct {
		Name            string
		Link            *Link
		Message         string
		ExpectedMessage string
	}{
		{
			"Simple pattern",
			&Link{
				Pattern:  "(Mattermost)",
				Template: "[Mattermost](https://mattermost.com)",
			},
			"Welcome to Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com)!",
		}, {
			"Pattern with variable name accessed using $variable",
			&Link{
				Pattern:  "(?P<key>Mattermost)",
				Template: "[$key](https://mattermost.com)",
			},
			"Welcome to Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com)!",
		}, {
			"Multiple replacments",
			&Link{
				Pattern:  "(?P<key>Mattermost)",
				Template: "[$key](https://mattermost.com)",
			},
			"Welcome to Mattermost and have fun with Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com) and have fun with [Mattermost](https://mattermost.com)!",
		}, {
			"Pattern with variable name accessed using ${variable}",
			&Link{
				Pattern:  "(?P<key>Mattermost)",
				Template: "[${key}](https://mattermost.com)",
			},
			"Welcome to Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com)!",
		}, {
			"Jira example",
			&Link{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		}, {
			"Jira example 2 (within a ())",
			&Link{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Link in brackets should link (see MM-12345)",
			"Link in brackets should link (see [MM-12345](https://mattermost.atlassian.net/browse/MM-12345))",
		}, {
			"Jira example 3 (before ,)",
			&Link{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Link a ticket MM-12345, before a comma",
			"Link a ticket [MM-12345](https://mattermost.atlassian.net/browse/MM-12345), before a comma",
		}, {
			"Jira example 3 (at begin of the message)",
			&Link{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"MM-12345 should link!",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		}, {
			"Pattern word prefix and suffix disabled",
			&Link{
				Pattern:              "(?P<previous>^|\\s)(MM)(-)(?P<jira_id>\\d+)",
				Template:             "${previous}[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
				DisableNonWordPrefix: true,
				DisableNonWordSuffix: true,
			},
			"Welcome MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		}, {
			"Pattern word prefix and suffix disabled (at begin of the message)",
			&Link{
				Pattern:              "(?P<previous>^|\\s)(MM)(-)(?P<jira_id>\\d+)",
				Template:             "${previous}[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
				DisableNonWordPrefix: true,
				DisableNonWordSuffix: true,
			},
			"MM-12345 should link!",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		}, {
			"Pattern word prefix and suffix enable (in the middle of other text)",
			&Link{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"WelcomeMM-12345should not link!",
			"WelcomeMM-12345should not link!",
		}, {
			"Pattern word prefix and suffix disabled (in the middle of other text)",
			&Link{
				Pattern:              "(MM)(-)(?P<jira_id>\\d+)",
				Template:             "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
				DisableNonWordPrefix: true,
				DisableNonWordSuffix: true,
			},
			"WelcomeMM-12345should link!",
			"Welcome[MM-12345](https://mattermost.atlassian.net/browse/MM-12345)should link!",
		}, {
			"Not relinking",
			&Link{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should not re-link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should not re-link!",
		}, {
			"Url replacement",
			&Link{
				Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		}, {
			"Url replacement multiple times",
			&Link{
				Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345. should link https://mattermost.atlassian.net/browse/MM-12346 !",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345). should link [MM-12346](https://mattermost.atlassian.net/browse/MM-12346) !",
		}, {
			"Url replacement multiple times and at beginning",
			&Link{
				Pattern:  "(https:\\/\\/mattermost.atlassian.net\\/browse\\/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"https://mattermost.atlassian.net/browse/MM-12345 https://mattermost.atlassian.net/browse/MM-12345",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		}, {
			"Url replacement at end",
			&Link{
				Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		},
	} {

		t.Run(tc.Name, func(t *testing.T) {
			p := setupTestPlugin(t, *tc.Link)
			post, _ := p.MessageWillBePosted(nil, &model.Post{
				Message: tc.Message,
			})

			assert.Equal(t, tc.ExpectedMessage, post.Message)
		})

	}
}

func TestAutolinkWordBoundaries(t *testing.T) {
	const pattern = "(KEY)(-)(?P<ID>\\d+)"
	const template = "[KEY-$ID](someurl/KEY-$ID)"
	const ref = "KEY-12345"
	const ID = "12345"
	const markdown = "[KEY-12345](someurl/KEY-12345)"

	var defaultLink = Link{
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
		Link       *Link
		Prefix     string
		Suffix     string
		ExpectFail bool
	}{
		{Name: "space both sides both breaks required", Prefix: " ", Suffix: " "},
		{Name: "space both sides left break not required", Prefix: " ", Suffix: " ", Link: &linkNoPrefix},
		{Name: "space both sides right break not required", Prefix: " ", Suffix: " ", Link: &linkNoSuffix},
		{Name: "space both sides neither break required", Prefix: " ", Suffix: " ", Link: &linkNoPrefixNoSuffix},

		{Name: "space left side both breaks required", Prefix: " ", ExpectFail: true},
		{Name: "space left side left break not required", Prefix: " ", Link: &linkNoPrefix, ExpectFail: true},
		{Name: "space left side right break not required", Prefix: " ", Link: &linkNoSuffix},
		{Name: "space left side neither break required", Prefix: " ", Link: &linkNoPrefixNoSuffix},

		{Name: "space right side both breaks required", Suffix: " ", ExpectFail: true},
		{Name: "space right side left break not required", Suffix: " ", Link: &linkNoPrefix},
		{Name: "space right side right break not required", Suffix: " ", Link: &linkNoSuffix, ExpectFail: true},
		{Name: "space right side neither break required", Prefix: " ", Link: &linkNoPrefixNoSuffix},

		{Name: "none both breaks required", ExpectFail: true},
		{Name: "none left break not required", Link: &linkNoPrefix, ExpectFail: true},
		{Name: "none right break not required", Link: &linkNoSuffix, ExpectFail: true},
		{Name: "none neither break required", Link: &linkNoPrefixNoSuffix},

		{Sep: "paren", Name: "2 parens", Prefix: "(", Suffix: ")"},
		{Sep: "paren", Name: "left paren", Prefix: "(", Link: &linkNoSuffix},
		{Sep: "paren", Name: "right paren", Suffix: ")", Link: &linkNoPrefix},
		{Sep: "sbracket", Name: "2 brackets", Prefix: "[", Suffix: "]"},
		{Sep: "lsbracket", Name: "both breaks", Prefix: "[", ExpectFail: true},
		{Sep: "lsbracket", Name: "bracket no prefix", Prefix: "[", Link: &linkNoPrefix, ExpectFail: true},
		{Sep: "lsbracket", Name: "bracket no suffix", Prefix: "[", Link: &linkNoSuffix},
		{Sep: "lsbracket", Name: "bracket neither prefix suffix", Prefix: "[", Link: &linkNoPrefixNoSuffix},
		{Sep: "rsbracket", Name: "bracket", Suffix: "]", Link: &linkNoPrefix},
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
			if l == nil {
				l = &defaultLink
			}
			p := setupTestPlugin(t, *l)

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

func TestAutolinkErrors(t *testing.T) {
	var tests = []struct {
		Name string
		Link Link
	}{
		{
			"Empty Link",
			Link{},
		}, {
			"No pattern",
			Link{
				Pattern:  "",
				Template: "blah",
			},
		}, {
			"No template",
			Link{
				Pattern:  "blah",
				Template: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			_, err := NewAutoLinker(tt.Link)
			assert.NotNil(t, err)
		})
	}
}
