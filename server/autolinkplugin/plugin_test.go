package autolinkplugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestPlugin(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{
			autolink.Autolink{
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

	testUser := model.User{
		IsBot: false,
	}
	api.On("GetUser", mock.AnythingOfType("string")).Return(&testUser, nil)

	p := New()
	p.SetAPI(api)
	p.OnConfigurationChange()

	post := &model.Post{Message: "Welcome to Mattermost!"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "Welcome to [Mattermost](https://mattermost.com)!", rpost.Message)
}

type SuiteAuthorization struct {
	suite.Suite

	api *plugintest.API

	adminUsernames string
	userInfo       map[string]*model.User
}

func (suite *SuiteAuthorization) SetupTest() {
	suite.adminUsernames = ""
	suite.userInfo = make(map[string]*model.User)

	suite.api = &plugintest.API{}
	suite.api.On(
		"LoadPluginConfiguration",
		mock.AnythingOfType("*autolinkplugin.Config"),
	).Return(
		func(dest interface{}) error {
			*dest.(*Config) = Config{
				PluginAdmins: suite.adminUsernames,
			}
			return nil
		},
	)
	suite.api.On(
		"GetUser",
		mock.AnythingOfType("string"),
	).Return(
		func(userId string) *model.User {
			return suite.userInfo[userId]
		},
		func(userId string) *model.AppError {
			if _, ok := suite.userInfo[userId]; ok {
				return nil
			} else {
				return &model.AppError{
					Message: fmt.Sprintf("user %s not found", userId),
				}
			}
		},
	)
	suite.api.On(
		"UnregisterCommand",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
	).Return(
		(*model.AppError)(nil),
	)
}

func (suite *SuiteAuthorization) TestSysadminIsAuthorized() {
	suite.userInfo["marynaId"] = &model.User{
		Username: "maryna",
		Id:       "marynaId",
		Roles:    "smurf,system_admin,reaper",
	}

	p := New()
	p.SetAPI(suite.api)

	err := p.OnConfigurationChange()
	require.NoError(suite.T(), err)

	allowed, err := p.IsAuthorizedAdmin("marynaId")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), allowed)
}

func (suite *SuiteAuthorization) TestPlainUserIsDenied() {
	suite.userInfo["marynaId"] = &model.User{
		Username: "maryna",
		Id:       "marynaId",
		Roles:    "smurf,reaper",
	}

	p := New()
	p.SetAPI(suite.api)

	err := p.OnConfigurationChange()
	require.NoError(suite.T(), err)

	allowed, err := p.IsAuthorizedAdmin("marynaId")
	require.NoError(suite.T(), err)
	assert.False(suite.T(), allowed)
}

func (suite *SuiteAuthorization) TestAdminUserIsAuthorized() {
	suite.userInfo["marynaId"] = &model.User{
		Username: "maryna",
		Id:       "marynaId",
		Roles:    "smurf,reaper",
	}
	suite.adminUsernames = "marynaId"

	p := New()
	p.SetAPI(suite.api)

	err := p.OnConfigurationChange()
	require.NoError(suite.T(), err)

	allowed, err := p.IsAuthorizedAdmin("marynaId")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), allowed)
}

func (suite *SuiteAuthorization) TestMultipleUsersAreAuthorized() {
	suite.userInfo["marynaId"] = &model.User{
		Username: "maryna",
		Id:       "marynaId",
		Roles:    "smurf,reaper",
	}
	suite.userInfo["borynaId"] = &model.User{
		Username: "boryna",
		Id:       "borynaId",
		Roles:    "smurf",
	}
	suite.userInfo["karynaId"] = &model.User{
		Username: "karyna",
		Id:       "karynaId",
		Roles:    "screamer",
	}
	suite.adminUsernames = "marynaId,karynaId"

	p := New()
	p.SetAPI(suite.api)

	err := p.OnConfigurationChange()
	require.NoError(suite.T(), err)

	allowed, err := p.IsAuthorizedAdmin("marynaId")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), allowed)

	allowed, err = p.IsAuthorizedAdmin("karynaId")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), allowed)

	allowed, err = p.IsAuthorizedAdmin("borynaId")
	require.NoError(suite.T(), err)
	assert.False(suite.T(), allowed)
}

func (suite *SuiteAuthorization) TestWhitespaceIsIgnored() {
	suite.userInfo["marynaId"] = &model.User{
		Username: "maryna",
		Id:       "marynaId",
		Roles:    "smurf,reaper",
	}
	suite.userInfo["borynaId"] = &model.User{
		Username: "boryna",
		Id:       "borynaId",
		Roles:    "smurf",
	}
	suite.userInfo["karynaId"] = &model.User{
		Username: "karyna",
		Id:       "karynaId",
		Roles:    "screamer",
	}
	suite.adminUsernames = "marynaId , karynaId, borynaId "

	p := New()
	p.SetAPI(suite.api)

	err := p.OnConfigurationChange()
	require.NoError(suite.T(), err)

	allowed, err := p.IsAuthorizedAdmin("marynaId")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), allowed)

	allowed, err = p.IsAuthorizedAdmin("karynaId")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), allowed)

	allowed, err = p.IsAuthorizedAdmin("borynaId")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), allowed)
}

func (suite *SuiteAuthorization) TestNonExistantUsersAreIgnored() {
	suite.userInfo["marynaId"] = &model.User{
		Username: "maryna",
		Id:       "marynaId",
		Roles:    "smurf,reaper",
	}
	suite.adminUsernames = "marynaId,karynaId"

	p := New()
	p.SetAPI(suite.api)

	err := p.OnConfigurationChange()
	require.NoError(suite.T(), err)

	allowed, err := p.IsAuthorizedAdmin("marynaId")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), allowed)

	allowed, err = p.IsAuthorizedAdmin("karynaId")
	require.Error(suite.T(), err)
}

func TestSuiteAuthorization(t *testing.T) {
	suite.Run(t, new(SuiteAuthorization))
}

func TestSpecialCases(t *testing.T) {
	links := make([]autolink.Autolink, 0)
	links = append(links, autolink.Autolink{
		Pattern:  "(Mattermost)",
		Template: "[Mattermost](https://mattermost.com)",
	}, autolink.Autolink{
		Pattern:  "(Example)",
		Template: "[Example](https://example.com)",
	}, autolink.Autolink{
		Pattern:  "MM-(?P<jira_id>\\d+)",
		Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
	}, autolink.Autolink{
		Pattern:  "https://mattermost.atlassian.net/browse/MM-(?P<jira_id>\\d+)",
		Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
	}, autolink.Autolink{
		Pattern:  "(foo!bar)",
		Template: "fb",
	}, autolink.Autolink{
		Pattern:  "(example)",
		Template: "test",
		Scope:    []string{"team/off-topic"},
	}, autolink.Autolink{
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

	testUser := model.User{
		IsBot: false,
	}
	api.On("GetUser", mock.AnythingOfType("string")).Return(&testUser, nil)

	p := New()
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

func TestBotMessagesAreRewritenWhenGetUserFails(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{
			autolink.Autolink{
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
	}).Once()
	api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*model.AppError)(nil)).Once()

	api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil).Once()
	api.On("GetTeam", mock.AnythingOfType("string")).Return(&testTeam, nil).Once()
	api.On("LogError", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*model.AppError"))

	err := &model.AppError{
		Message: "foo error!",
	}
	api.On("GetUser", mock.AnythingOfType("string")).Return(nil, err).Once()

	p := New()
	p.SetAPI(api)
	p.OnConfigurationChange()

	post := &model.Post{Message: "Welcome to Mattermost!"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "Welcome to [Mattermost](https://mattermost.com)!", rpost.Message)
}

func TestGetUserApiCallIsNotExecutedWhenThereAreNoChanges(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{
			autolink.Autolink{
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
	}).Once()
	api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*model.AppError)(nil)).Once()

	api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil).Once()
	api.On("GetTeam", mock.AnythingOfType("string")).Return(&testTeam, nil).Once()
	api.On("LogDebug", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"))

	p := New()
	p.SetAPI(api)
	p.OnConfigurationChange()

	post := &model.Post{Message: "Welcome to FooBarism!"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "Welcome to FooBarism!", rpost.Message)
}

func TestBotMessagesAreNotRewriten(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{
			autolink.Autolink{
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
	}).Once()
	api.On("UnregisterCommand", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return((*model.AppError)(nil)).Once()

	api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil).Once()
	api.On("GetTeam", mock.AnythingOfType("string")).Return(&testTeam, nil).Once()
	testUser := model.User{
		IsBot: true,
	}
	api.On("GetUser", mock.AnythingOfType("string")).Return(&testUser, nil).Once()
	api.On("LogDebug", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"))

	p := New()
	p.SetAPI(api)
	p.OnConfigurationChange()

	post := &model.Post{Message: "Welcome to Mattermost!"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Nil(t, rpost)
}

func TestHashtags(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{
			autolink.Autolink{
				Pattern:  "foo",
				Template: "#bar",
			},
			autolink.Autolink{
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

	testUser := model.User{
		IsBot: false,
	}
	api.On("GetUser", mock.AnythingOfType("string")).Return(&testUser, nil)

	p := New()
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
		Links: []autolink.Autolink{
			autolink.Autolink{
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
	api.On("SavePluginConfig", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	p := New()
	p.SetAPI(api)
	p.OnConfigurationChange()
	p.OnActivate()

	jbyte, _ := json.Marshal(&autolink.Autolink{Name: "new", Pattern: "newpat", Template: "newtemp"})
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/link", bytes.NewReader(jbyte))
	p.ServeHTTP(&plugin.Context{SourcePluginId: "somthing"}, recorder, req)
	resp := recorder.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, p.conf.Links, 2)
	assert.Equal(t, "new", p.conf.Links[1].Name)
}
