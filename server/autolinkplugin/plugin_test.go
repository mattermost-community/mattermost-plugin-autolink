package autolinkplugin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost-plugin-autolink/server/link"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlugin(t *testing.T) {
	conf := Config{
		Links: []link.Link{
			link.Link{
				Pattern:  "(Mattermost)",
				Template: "[Mattermost](https://mattermost.com)",
			},
		},
	}

	testChannel := model.Channel{
		Name: "TestChanel",
	}

	testTeam := model.Team{
		Name: "TestTeam",
	}

	api := &plugintest.API{}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*autolinkplugin.Config")).Return(func(dest interface{}) error {
		*dest.(*Config) = conf
		return nil
	})
	api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*model.AppError)(nil))

	api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil)
	api.On("GetTeam", mock.AnythingOfType("string")).Return(&testTeam, nil)

	p := Plugin{}
	p.SetAPI(api)
	p.OnConfigurationChange()

	post := &model.Post{Message: "Welcome to Mattermost!"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "Welcome to [Mattermost](https://mattermost.com)!", rpost.Message)
}

func TestSpecialCases(t *testing.T) {
	links := make([]link.Link, 0)
	links = append(links, link.Link{
		Pattern:  "(Mattermost)",
		Template: "[Mattermost](https://mattermost.com)",
	}, link.Link{
		Pattern:  "(Example)",
		Template: "[Example](https://example.com)",
	}, link.Link{
		Pattern:  "MM-(?P<jira_id>\\d+)",
		Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
	}, link.Link{
		Pattern:  "https://mattermost.atlassian.net/browse/MM-(?P<jira_id>\\d+)",
		Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
	}, link.Link{
		Pattern:  "(foo!bar)",
		Template: "fb",
	}, link.Link{
		Pattern:  "(example)",
		Template: "test",
		Scope:    []string{"team/off-topic"},
	}, link.Link{
		Pattern:  "(example)",
		Template: "test",
		Scope:    []string{"other-team/town-square"},
	})
	validConfig := Config{
		EnableAdminCommand: false,
		EnableOnUpdate:     true,
		Links:              links,
	}

	testChannel := model.Channel{
		Name: "TestChanel",
	}

	testTeam := model.Team{
		Name: "TestTeam",
	}

	api := &plugintest.API{}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*autolinkplugin.Config")).Return(func(dest interface{}) error {
		*dest.(*Config) = validConfig
		return nil
	})
	api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*model.AppError)(nil))

	api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil)
	api.On("GetTeam", mock.AnythingOfType("string")).Return(&testTeam, nil)

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
			"check out MM-12345 too",
			"check out [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) too",
		},
	}

	for _, tt := range tests {
		t.Run(tt.inputMessage, func(t *testing.T) {
			{
				// user creates a new post

				post := &model.Post{
					Message: tt.inputMessage,
				}

				rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

				assert.Equal(t, tt.expectedMessage, rpost.Message)
			}
			{
				// user updates the modified post but with no changes

				post := &model.Post{
					Message: tt.expectedMessage,
				}

				rpost, _ := p.MessageWillBeUpdated(&plugin.Context{}, post, post)

				assert.Equal(t, tt.expectedMessage, rpost.Message)
			}
			{
				// user updates the modified post and sets it back to the original text

				originalPost := &model.Post{
					Message: tt.expectedMessage,
				}
				post := &model.Post{
					Message: tt.inputMessage,
				}

				rpost, _ := p.MessageWillBeUpdated(&plugin.Context{}, originalPost, post)

				assert.Equal(t, tt.expectedMessage, rpost.Message)
			}
			{
				// user updates an empty post to the original text

				emptyPost := &model.Post{}
				post := &model.Post{
					Message: tt.inputMessage,
				}

				rpost, _ := p.MessageWillBeUpdated(&plugin.Context{}, post, emptyPost)

				assert.Equal(t, tt.expectedMessage, rpost.Message)
			}
		})
	}
}

func TestHashtags(t *testing.T) {
	conf := Config{
		Links: []link.Link{
			link.Link{
				Pattern:  "foo",
				Template: "#bar",
			},
			link.Link{
				Pattern:  "hash tags",
				Template: "#hash #tags",
			},
		},
	}

	testChannel := model.Channel{
		Name: "TestChanel",
	}

	testTeam := model.Team{
		Name: "TestTeam",
	}

	api := &plugintest.API{}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*autolinkplugin.Config")).Return(func(dest interface{}) error {
		*dest.(*Config) = conf
		return nil
	})
	api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*model.AppError)(nil))

	api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil)
	api.On("GetTeam", mock.AnythingOfType("string")).Return(&testTeam, nil)

	p := Plugin{}
	p.SetAPI(api)
	p.OnConfigurationChange()

	post := &model.Post{Message: "foo"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "#bar", rpost.Hashtags)

	post.Message = "hash tags"
	rpost, _ = p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "#hash #tags", rpost.Hashtags)
}

func TestAPI(t *testing.T) {
	conf := Config{
		Links: []link.Link{
			link.Link{
				Name:     "existing",
				Pattern:  "thing",
				Template: "otherthing",
			},
		},
	}

	testChannel := model.Channel{
		Name: "TestChanel",
	}

	testTeam := model.Team{
		Name: "TestTeam",
	}

	api := &plugintest.API{}
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*autolinkplugin.Config")).Return(func(dest interface{}) error {
		*dest.(*Config) = conf
		return nil
	})
	api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*model.AppError)(nil))
	api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil)
	api.On("GetTeam", mock.AnythingOfType("string")).Return(&testTeam, nil)
	api.On("GetUser", mock.AnythingOfType("string")).Return(&testTeam, nil)
	api.On("SavePluginConfig", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	p := Plugin{}
	p.SetAPI(api)
	p.OnConfigurationChange()
	p.OnActivate()

	jbyte, _ := json.Marshal(&link.Link{Name: "new", Pattern: "newpat", Template: "newtemp"})
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/link", bytes.NewReader(jbyte))
	p.ServeHTTP(&plugin.Context{SourcePluginId: "somthing"}, recorder, req)
	resp := recorder.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, p.conf.Links, 2)
	assert.Equal(t, "new", p.conf.Links[1].Name)
}
