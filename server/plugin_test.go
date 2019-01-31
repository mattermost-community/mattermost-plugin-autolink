package main

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
)

func TestPlugin(t *testing.T) {
	links := make([]*Link, 0)
	links = append(links, &Link{
		Pattern:  "(Mattermost)",
		Template: "[Mattermost](https://mattermost.com)",
	})
	validConfiguration := Configuration{links}

	api := &plugintest.API{}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(func(dest interface{}) error {
		*dest.(*Configuration) = validConfiguration
		return nil
	})

	p := Plugin{}
	p.SetAPI(api)
	p.OnConfigurationChange()

	post := &model.Post{Message: "Welcome to Mattermost!"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "Welcome to [Mattermost](https://mattermost.com)!", rpost.Message)
}

func TestSpecialCases(t *testing.T) {
	links := make([]*Link, 0)
	links = append(links, &Link{
		Pattern:  "https://mattermost.com",
		Template: "[the mattermost portal](https://mattermost.com)",
	}, &Link{
		Pattern:  "(Mattermost)",
		Template: "[Mattermost](https://mattermost.com)",
	}, &Link{
		Pattern:  "https://mattermost.atlassian.net/browse/MM-(?P<jira_id>\\d+)",
		Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
	}, &Link{
		Pattern:  "MM-(?P<jira_id>\\d+)",
		Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
	}, &Link{
		Pattern:  "(Example)",
		Template: "[Example](https://example.com)",
	}, &Link{
		Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
		Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
	}, &Link{
		Pattern:  "(foo!bar)",
		Template: "fb",
	})
	validConfiguration := Configuration{links}

	api := &plugintest.API{}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(func(dest interface{}) error {
		*dest.(*Configuration) = validConfiguration
		return nil
	})

	p := Plugin{}
	p.SetAPI(api)
	p.OnConfigurationChange()

	var tests = []struct {
		inputMessage    string
		expectedMessage string
	}{
		{
			"hello ``` Mattermost ``` goodbye",
			"hello ``` Mattermost ``` goodbye",
		}, {
			"hello\n```\nMattermost\n```\ngoodbye",
			"hello\n```\nMattermost\n```\ngoodbye",
		}, {
			"Mattermost ``` Mattermost ``` goodbye",
			"[Mattermost](https://mattermost.com) ``` Mattermost ``` goodbye",
		}, {
			"``` Mattermost ``` Mattermost",
			"``` Mattermost ``` [Mattermost](https://mattermost.com)",
		}, {
			"Mattermost ``` Mattermost ```",
			"[Mattermost](https://mattermost.com) ``` Mattermost ```",
		}, {
			"Mattermost ``` Mattermost ```\n\n",
			"[Mattermost](https://mattermost.com) ``` Mattermost ```\n\n",
		}, {
			"hello ` Mattermost ` goodbye",
			"hello ` Mattermost ` goodbye",
		}, {
			"hello\n`\nMattermost\n`\ngoodbye",
			"hello\n`\nMattermost\n`\ngoodbye",
		}, {
			"Mattermost ` Mattermost ` goodbye",
			"[Mattermost](https://mattermost.com) ` Mattermost ` goodbye",
		}, {
			"` Mattermost ` Mattermost",
			"` Mattermost ` [Mattermost](https://mattermost.com)",
		}, {
			"Mattermost ` Mattermost `",
			"[Mattermost](https://mattermost.com) ` Mattermost `",
		}, {
			"Mattermost ` Mattermost `\n\n",
			"[Mattermost](https://mattermost.com) ` Mattermost `\n\n",
		}, {
			"hello ``` Mattermost ``` goodbye ` Mattermost ` end",
			"hello ``` Mattermost ``` goodbye ` Mattermost ` end",
		}, {
			"hello\n```\nMattermost\n```\ngoodbye ` Mattermost ` end",
			"hello\n```\nMattermost\n```\ngoodbye ` Mattermost ` end",
		}, {
			"Mattermost ``` Mattermost ``` goodbye ` Mattermost ` end",
			"[Mattermost](https://mattermost.com) ``` Mattermost ``` goodbye ` Mattermost ` end",
		}, {
			"``` Mattermost ``` Mattermost",
			"``` Mattermost ``` [Mattermost](https://mattermost.com)",
		}, {
			"```\n` Mattermost `\n```\nMattermost",
			"```\n` Mattermost `\n```\n[Mattermost](https://mattermost.com)",
		}, {
			"  Mattermost",
			"  [Mattermost](https://mattermost.com)",
		}, {
			"    Mattermost",
			"    Mattermost",
		}, {
			"    ```\nMattermost\n    ```",
			"    ```\n[Mattermost](https://mattermost.com)\n    ```",
		}, {
			"` ``` `\nMattermost\n` ``` `",
			"` ``` `\n[Mattermost](https://mattermost.com)\n` ``` `",
		}, {
			"Mattermost \n Mattermost",
			"[Mattermost](https://mattermost.com) \n [Mattermost](https://mattermost.com)",
		}, {
			"[Mattermost](https://mattermost.com)",
			"[Mattermost](https://mattermost.com)",
		}, {
			"[  Mattermost  ](https://mattermost.com)",
			"[  Mattermost  ](https://mattermost.com)",
		}, {
			"[  Mattermost  ][1]\n\n[1]: https://mattermost.com",
			"[  Mattermost  ][1]\n\n[1]: https://mattermost.com",
		}, {
			"![  Mattermost  ](https://mattermost.com/example.png)",
			"![  Mattermost  ](https://mattermost.com/example.png)",
		}, {
			"![  Mattermost  ][1]\n\n[1]: https://mattermost.com/example.png",
			"![  Mattermost  ][1]\n\n[1]: https://mattermost.com/example.png",
		}, {
			"Why not visit https://mattermost.com?",
			"Why not visit [the mattermost portal](https://mattermost.com)?",
		}, {
			"Please check https://mattermost.atlassian.net/browse/MM-123 for details",
			"Please check [MM-123](https://mattermost.atlassian.net/browse/MM-123) for details",
		}, {
			"Please check MM-123 for details",
			"Please check [MM-123](https://mattermost.atlassian.net/browse/MM-123) for details",
		}, {
			"foo!bar\nExample\nfoo!bar Mattermost",
			"fb\n[Example](https://example.com)\nfb [Mattermost](https://mattermost.com)",
		}, {
			"foo!bar",
			"fb",
		}, {
			"foo!barfoo!bar",
			"foo!barfoo!bar",
		}, {
			"foo!bar & foo!bar",
			"fb & fb",
		}, {
			"foo!bar & foo!bar\nfoo!bar & foo!bar\nfoo!bar & foo!bar",
			"fb & fb\nfb & fb\nfb & fb",
		}, {
			"https://mattermost.atlassian.net/browse/MM-12345",
			"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		}, {
			"Welcome https://mattermost.atlassian.net/browse/MM-12345",
			"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
		}, {
			"text https://mattermost.atlassian.net/browse/MM-12345 other text",
			"text [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) other text",
		}, {
			"text [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) other text",
			"text [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) other text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.inputMessage, func(t *testing.T) {
			post := &model.Post{
				Message: tt.inputMessage,
			}

			expectedMessagePost := &model.Post{
				Message: tt.expectedMessage,
			}

			rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

			assert.Equal(t, tt.expectedMessage, rpost.Message)

			// Reapplying mustn't change the result
			rpost, _ = p.MessageWillBePosted(&plugin.Context{}, expectedMessagePost)

			assert.Equal(t, tt.expectedMessage, rpost.Message)

			upost, _ := p.MessageWillBeUpdated(&plugin.Context{}, post, post)

			assert.Equal(t, tt.expectedMessage, upost.Message)

			// Reapplying mustn't change the result
			upost, _ = p.MessageWillBeUpdated(&plugin.Context{}, expectedMessagePost, expectedMessagePost)

			assert.Equal(t, tt.expectedMessage, upost.Message)
		})
	}
}
