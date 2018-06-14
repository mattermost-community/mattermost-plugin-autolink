package main

import (
	"testing"
)

func TestAutolink(t *testing.T) {
	var tests = []struct {
		Link            *Link
		inputMessage    string
		expectedMessage string
	}{
		{
			&Link{
				Pattern:  "(Mattermost)",
				Template: "[Mattermost](https://mattermost.com)",
			},
			"Welcome to Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com)!",
		},
		{
			&Link{
				Pattern:  "(?P<key>Mattermost)",
				Template: "[$key](https://mattermost.com)",
			},
			"Welcome to Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com)!",
		},
		{
			&Link{
				Pattern:  "(?P<key>Mattermost)",
				Template: "[${key}](https://mattermost.com)",
			},
			"Welcome to Mattermost!",
			"Welcome to [Mattermost](https://mattermost.com)!",
		},
		{
			&Link{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		},
		{
			&Link{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"MM-12345 should link!",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		},
		{
			&Link{
				Pattern:              "(?P<previous>^|\\s)(MM)(-)(?P<jira_id>\\d+)",
				Template:             "${previous}[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
				DisableNonWordPrefix: true,
				DisableNonWordSuffix: true,
			},
			"Welcome MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		},
		{
			&Link{
				Pattern:              "(?P<previous>^|\\s)(MM)(-)(?P<jira_id>\\d+)",
				Template:             "${previous}[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
				DisableNonWordPrefix: true,
				DisableNonWordSuffix: true,
			},
			"MM-12345 should link!",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		},
		{
			&Link{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"WelcomeMM-12345should not link!",
			"WelcomeMM-12345should not link!",
		},
		{
			&Link{
				Pattern:              "(MM)(-)(?P<jira_id>\\d+)",
				Template:             "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
				DisableNonWordPrefix: true,
				DisableNonWordSuffix: true,
			},
			"WelcomeMM-12345should link!",
			"Welcome[MM-12345](https://mattermost.atlassian.net/browse/MM-12345)should link!",
		},
		{
			&Link{
				Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should not re-link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should not re-link!",
		},
		{
			&Link{
				Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345 should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
		},
		{
			&Link{
				Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345. should link!",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345). should link!",
		},
		{
			&Link{
				Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		},
		{
			&Link{
				Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
				Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			},
			"Welcome https://mattermost.atlassian.net/browse/MM-12345",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		},
	}

	for _, tt := range tests {
		al, _ := NewAutoLinker(tt.Link)
		actual := al.Replace(tt.inputMessage)
		if actual != tt.expectedMessage {
			t.Fatalf("autolink:\n expected '%v'\n actual   '%v'", tt.expectedMessage, actual)
		}
	}
}

func TestAutolinkErrors(t *testing.T) {
	var tests = []struct {
		Link *Link
	}{
		{},
		{
			&Link{},
		},
		{
			&Link{
				Pattern:  "",
				Template: "blah",
			},
		},
		{
			&Link{
				Pattern:  "blah",
				Template: "",
			},
		},
	}

	for _, tt := range tests {
		_, err := NewAutoLinker(tt.Link)
		if err == nil {
			t.Fatalf("should have failed to parse regex")
		}
	}
}
