package autolinkplugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/mattermost/mattermost/server/public/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mattermost-community/mattermost-plugin-autolink/server/autolink"
)

func TestPlugin(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{{
			Pattern:  "(Mattermost)",
			Template: "[Mattermost](https://mattermost.com)",
		}},
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
	err := p.OnConfigurationChange()
	require.NoError(t, err)

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
		func(userID string) *model.User {
			return suite.userInfo[userID]
		},
		func(userID string) *model.AppError {
			if _, ok := suite.userInfo[userID]; !ok {
				return &model.AppError{
					Message: fmt.Sprintf("user %s not found", userID),
				}
			}

			return nil
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

	suite.api.On("LogInfo", mock.AnythingOfType("string")).Return(nil)

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

	suite.api.On("LogInfo", mock.AnythingOfType("string")).Return(nil)

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

	suite.api.On("LogInfo", mock.AnythingOfType("string")).Return(nil)

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

	suite.api.On("LogInfo", mock.AnythingOfType("string")).Return(nil)

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

	suite.api.On("LogWarn", mock.AnythingOfType("string"),
		"userID",
		"karynaId",
		"error",
		mock.AnythingOfType("*model.AppError"),
	).Return(nil)
	suite.api.On("LogInfo", mock.AnythingOfType("string")).Return(nil)

	p := New()
	p.SetAPI(suite.api)

	err := p.OnConfigurationChange()
	require.NoError(suite.T(), err)

	allowed, err := p.IsAuthorizedAdmin("marynaId")
	require.NoError(suite.T(), err)
	assert.True(suite.T(), allowed)

	allowed, err = p.IsAuthorizedAdmin("karynaId")
	require.Error(suite.T(), err)
	require.False(suite.T(), allowed)
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
	err := p.OnConfigurationChange()
	require.NoError(t, err)

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
		Links: []autolink.Autolink{{
			Pattern:  "(Mattermost)",
			Template: "[Mattermost](https://mattermost.com)",
		}},
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

	api.On("GetUser", mock.AnythingOfType("string")).Return(nil, &model.AppError{
		Message: "foo error!",
	}).Once()

	p := New()
	p.SetAPI(api)
	err := p.OnConfigurationChange()
	require.NoError(t, err)

	post := &model.Post{Message: "Welcome to Mattermost!"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "Welcome to [Mattermost](https://mattermost.com)!", rpost.Message)
}

func TestGetUserApiCallIsNotExecutedWhenThereAreNoChanges(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{{
			Pattern:  "(Mattermost)",
			Template: "[Mattermost](https://mattermost.com)",
		}},
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
	err := p.OnConfigurationChange()
	require.NoError(t, err)

	post := &model.Post{Message: "Welcome to FooBarism!"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "Welcome to FooBarism!", rpost.Message)
}

func TestBotMessagesAreNotRewriten(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{{
			Pattern:  "(Mattermost)",
			Template: "[Mattermost](https://mattermost.com)",
		}},
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
	err := p.OnConfigurationChange()
	require.NoError(t, err)

	post := &model.Post{Message: "Welcome to Mattermost!"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, post.Message, rpost.Message)
}

func TestBotMessagesAreRewritenWhenConfigAllows(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{{
			Pattern:         "(Mattermost)",
			Template:        "[Mattermost](https://mattermost.com)",
			ProcessBotPosts: true,
		}},
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

	testUser := model.User{
		IsBot: true,
	}
	api.On("GetUser", mock.AnythingOfType("string")).Return(&testUser, nil).Once()

	p := New()
	p.SetAPI(api)
	err := p.OnConfigurationChange()
	require.NoError(t, err)

	post := &model.Post{Message: "Welcome to Mattermost!"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)
	require.NotNil(t, rpost)

	assert.Equal(t, "Welcome to [Mattermost](https://mattermost.com)!", rpost.Message)
}

func TestHashtags(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{{
			Pattern:  "foo",
			Template: "#bar",
		}, {
			Pattern:  "hash tags",
			Template: "#hash #tags",
		}},
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
	err := p.OnConfigurationChange()
	require.NoError(t, err)

	post := &model.Post{Message: "foo"}
	rpost, _ := p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "#bar", rpost.Hashtags)

	post.Message = "hash tags"
	rpost, _ = p.MessageWillBePosted(&plugin.Context{}, post)

	assert.Equal(t, "#hash #tags", rpost.Hashtags)
}

func TestAPI(t *testing.T) {
	conf := Config{
		Links: []autolink.Autolink{{
			Name:     "existing",
			Pattern:  "thing",
			Template: "otherthing",
		}},
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
	err := p.OnConfigurationChange()
	require.NoError(t, err)
	err = p.OnActivate()
	require.NoError(t, err)

	jbyte, err := json.Marshal(&autolink.Autolink{Name: "new", Pattern: "newpat", Template: "newtemp"})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/api/v1/link", bytes.NewReader(jbyte))
	require.NoError(t, err)
	req.Header.Set("Mattermost-Plugin-ID", "somthing")
	p.ServeHTTP(&plugin.Context{}, recorder, req)
	resp := recorder.Result()
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Len(t, p.conf.Links, 2)
	assert.Equal(t, "new", p.conf.Links[1].Name)
}

func TestResolveScope(t *testing.T) {
	t.Run("resolve channel name and team name", func(t *testing.T) {
		testChannel := model.Channel{
			Name:   "TestChannel",
			TeamId: "TestId",
		}

		testTeam := model.Team{
			Name: "TestTeam",
		}

		api := &plugintest.API{}
		api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil)
		api.On("GetTeam", mock.AnythingOfType("string")).Return(&testTeam, nil)

		p := Plugin{}
		p.SetAPI(api)
		channelName, teamName, _ := p.resolveScope("TestId")
		assert.Equal(t, "TestChannel", channelName)
		assert.Equal(t, "TestTeam", teamName)
	})

	t.Run("resolve channel name and returns empty team name", func(t *testing.T) {
		testChannel := model.Channel{
			Name: "TestChannel",
		}

		api := &plugintest.API{}
		api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil)

		p := Plugin{}
		p.SetAPI(api)

		channelName, teamName, _ := p.resolveScope("TestId")
		assert.Equal(t, "TestChannel", channelName)
		assert.Equal(t, "", teamName)
	})

	t.Run("error when api fails to get channel", func(t *testing.T) {
		api := &plugintest.API{}

		api.On("GetChannel",
			mock.AnythingOfType("string")).Return(nil, &model.AppError{})

		p := Plugin{}
		p.SetAPI(api)

		channelName, teamName, err := p.resolveScope("TestId")
		assert.Error(t, err)
		assert.Equal(t, teamName, "")
		assert.Equal(t, channelName, "")
	})

	t.Run("error when api fails to get team", func(t *testing.T) {
		testChannel := model.Channel{
			Name:   "TestChannel",
			TeamId: "TestId",
		}

		api := &plugintest.API{}

		api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil)
		api.On("GetTeam", mock.AnythingOfType("string")).Return(nil, &model.AppError{})

		p := Plugin{}
		p.SetAPI(api)

		channelName, teamName, err := p.resolveScope("TestId")
		assert.Error(t, err)
		assert.Equal(t, channelName, "")
		assert.Equal(t, teamName, "")
	})
}

func TestProcessPost(t *testing.T) {
	t.Run("cannot resolve scope", func(t *testing.T) {
		conf := Config{
			Links: []autolink.Autolink{
				{
					Pattern:  "(Mattermost)",
					Template: "[Mattermost](https://mattermost.com)",
					Scope:    []string{"TestTeam/TestChannel"},
				},
			},
		}

		api := &plugintest.API{}

		api.On("LoadPluginConfiguration",
			mock.AnythingOfType("*autolinkplugin.Config")).Return(func(dest interface{}) error {
			*dest.(*Config) = conf
			return nil
		})
		api.On("UnregisterCommand", mock.AnythingOfType("string"),
			mock.AnythingOfType("string")).Return((*model.AppError)(nil))

		api.On("GetChannel", mock.AnythingOfType("string")).Return(nil, &model.AppError{})

		api.On("LogError",
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string")).Return(nil)

		testUser := model.User{
			IsBot: false,
		}
		api.On("GetUser", mock.AnythingOfType("string")).Return(&testUser, nil)

		p := New()
		p.SetAPI(api)
		_ = p.OnConfigurationChange()

		post := &model.Post{Message: "Welcome to Mattermost!"}
		rpost, _ := p.ProcessPost(&plugin.Context{}, post)

		assert.Equal(t, "Welcome to Mattermost!", rpost.Message)
	})

	t.Run("team name is empty", func(t *testing.T) {
		conf := Config{
			Links: []autolink.Autolink{
				{
					Pattern:  "(Mattermost)",
					Template: "[Mattermost](https://mattermost.com)",
					Scope:    []string{"TestTeam/TestChannel"},
				},
			},
		}

		testChannel := model.Channel{
			Name: "TestChannel",
		}

		api := &plugintest.API{}

		api.On("LoadPluginConfiguration",
			mock.AnythingOfType("*autolinkplugin.Config")).Return(func(dest interface{}) error {
			*dest.(*Config) = conf
			return nil
		})
		api.On("UnregisterCommand", mock.AnythingOfType("string"),
			mock.AnythingOfType("string")).Return((*model.AppError)(nil))
		api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil)
		api.On("LogError",
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string")).Return(nil)

		testUser := model.User{
			IsBot: false,
		}
		api.On("GetUser", mock.AnythingOfType("string")).Return(&testUser, nil)

		p := New()
		p.SetAPI(api)
		_ = p.OnConfigurationChange()

		post := &model.Post{Message: "Welcome to Mattermost!"}
		rpost, _ := p.ProcessPost(&plugin.Context{}, post)

		assert.Equal(t, "Welcome to Mattermost!", rpost.Message)
	})

	t.Run("valid scope replaces text", func(t *testing.T) {
		conf := Config{
			Links: []autolink.Autolink{
				{
					Pattern:  "(Mattermost)",
					Template: "[Mattermost](https://mattermost.com)",
					Scope:    []string{"TestTeam/TestChannel"},
				},
			},
		}

		testChannel := model.Channel{
			Name:   "TestChannel",
			TeamId: "TestId",
		}

		testTeam := model.Team{
			Name: "TestTeam",
		}

		api := &plugintest.API{}

		api.On("LoadPluginConfiguration",
			mock.AnythingOfType("*autolinkplugin.Config")).Return(func(dest interface{}) error {
			*dest.(*Config) = conf
			return nil
		})
		api.On("UnregisterCommand", mock.AnythingOfType("string"),
			mock.AnythingOfType("string")).Return((*model.AppError)(nil))

		api.On("GetChannel", mock.AnythingOfType("string")).Return(&testChannel, nil)
		api.On("GetTeam", mock.AnythingOfType("string")).Return(&testTeam, nil)

		testUser := model.User{
			IsBot: false,
		}
		api.On("GetUser", mock.AnythingOfType("string")).Return(&testUser, nil)

		p := New()
		p.SetAPI(api)
		_ = p.OnConfigurationChange()

		post := &model.Post{Message: "Welcome to Mattermost!"}
		rpost, _ := p.ProcessPost(&plugin.Context{}, post)

		assert.Equal(t, "Welcome to [Mattermost](https://mattermost.com)!", rpost.Message)
	})
}

func TestInScope(t *testing.T) {
	t.Run("returns true if scope array is empty", func(t *testing.T) {
		p := &Plugin{}
		result := p.inScope([]string{}, "TestChannel", "TestTeam")
		assert.Equal(t, true, result)
	})

	t.Run("returns true when team and channels are valid", func(t *testing.T) {
		p := &Plugin{}
		result := p.inScope([]string{"TestTeam/TestChannel"}, "TestChannel", "TestTeam")
		assert.Equal(t, true, result)
	})

	t.Run("returns false when channel is empty", func(t *testing.T) {
		p := &Plugin{}
		result := p.inScope([]string{"TestTeam/"}, "TestChannel", "TestTeam")
		assert.Equal(t, false, result)
	})

	t.Run("returns false when team is empty", func(t *testing.T) {
		p := &Plugin{}
		result := p.inScope([]string{"TestTeam/TestChannel"}, "TestChannel", "")
		assert.Equal(t, false, result)
	})

	t.Run("returns false on empty scope", func(t *testing.T) {
		p := &Plugin{}
		result := p.inScope([]string{""}, "TestChannel", "TestTeam")
		assert.Equal(t, false, result)
	})

	t.Run("returns true on team scope only", func(t *testing.T) {
		p := &Plugin{}
		result := p.inScope([]string{"TestTeam"}, "TestChannel", "TestTeam")
		assert.Equal(t, true, result)
	})
}
