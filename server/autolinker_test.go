package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutolink(t *testing.T) {
	var tests = []struct {
		Name            string
		Link            *Link
		inputMessage    string
		expectedMessage string
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
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			al, _ := NewAutoLinker(tt.Link)
			actual := al.Replace(tt.inputMessage)

			assert.Equal(t, tt.expectedMessage, actual)
		})
	}
}

func TestAutolinkErrors(t *testing.T) {
	var tests = []struct {
		Name string
		Link *Link
	}{
		{
			"No Link at all",
			nil,
		}, {
			"Empty Link",
			&Link{},
		}, {
			"No pattern",
			&Link{
				Pattern:  "",
				Template: "blah",
			},
		}, {
			"No template",
			&Link{
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
