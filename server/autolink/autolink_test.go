package autolink_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
	"github.com/mattermost/mattermost-plugin-autolink/server/autolinkplugin"
)

func setupTestPlugin(t *testing.T, l autolink.Autolink) *autolinkplugin.Plugin {
	p := autolinkplugin.New()
	api := &plugintest.API{}

	api.On("GetChannel", mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
		api.On("GetTeam", mock.AnythingOfType("string")).Return(&model.Team{}, (*model.AppError)(nil))
	}).Return(&model.Channel{
		Id:     "thechannel_id0123456789012",
		TeamId: "theteam_id0123456789012345",
		Name:   "thechannel_name",
	}, (*model.AppError)(nil))

	testUser := model.User{
		IsBot: false,
	}
	api.On("GetUser", mock.AnythingOfType("string")).Return(&testUser, nil)

	p.SetAPI(api)

	err := l.Compile()
	require.Nil(t, err)
	p.UpdateConfig(func(conf *autolinkplugin.Config) {
		conf.Links = []autolink.Autolink{l}
	})
	return p
}

type linkTest struct {
	Name            string
	Link            autolink.Autolink
	Message         string
	ExpectedMessage string
}

var commonLinkTests = []linkTest{
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
	},
}

func testLinks(t *testing.T, tcs ...linkTest) {
	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			p := setupTestPlugin(t, tc.Link)
			post, _ := p.MessageWillBePosted(nil, &model.Post{
				Message: tc.Message,
			})

			assert.Equal(t, tc.ExpectedMessage, post.Message)
		})
	}
}

func TestCommonLinks(t *testing.T) {
	testLinks(t, commonLinkTests...)
}

func TestLegacyWordBoundaries(t *testing.T) {
	const pattern = "(KEY)(-)(?P<ID>\\d+)"
	const template = "[KEY-$ID](someurl/KEY-$ID)"
	const ref = "KEY-12345"
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
func TestEquals(t *testing.T) {
	for _, tc := range []struct {
		l1, l2      autolink.Autolink
		expectEqual bool
	}{
		{
			l1: autolink.Autolink{
				Name: "test",
			},
			expectEqual: false,
		},
		{
			l1: autolink.Autolink{
				Name: "test",
			},
			l2: autolink.Autolink{
				Name: "test",
			},
			expectEqual: true,
		},
	} {
		t.Run(tc.l1.Name+"-"+tc.l2.Name, func(t *testing.T) {
			eq := tc.l1.Equals(tc.l2)
			assert.Equal(t, tc.expectEqual, eq)
		})
	}
}

func TestWildcard(t *testing.T) {
	for _, tc := range []struct {
		Name string
		Link autolink.Autolink
	}{
		{
			Name: ".*",
			Link: autolink.Autolink{
				Pattern:  ".*",
				Template: "My template",
			},
		},
		{
			Name: ".*.",
			Link: autolink.Autolink{
				Pattern:  ".*.",
				Template: "My template",
			},
		},
	} {
		p := setupTestPlugin(t, tc.Link)

		message := "Your message"

		var post *model.Post
		done := make(chan bool)
		go func() {
			post, _ = p.MessageWillBePosted(nil, &model.Post{
				Message: message,
			})

			done <- true
		}()

		select {
		case <-done:
		case <-time.After(50 * time.Millisecond):
			panic("wildcard regex timed out")
		}

		assert.NotNil(t, post, "post is nil")
		assert.Equal(t, "My template", post.Message)
	}
}
