package main

import (
	"testing"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/mattermost/mattermost-server/plugin/plugintest/mock"
)

func TestPlugin(t *testing.T) {
	links := make([]*Link, 0)
	links = append(links, &Link{
		Pattern:  "(Mattermost)",
		Template: "[Mattermost](https://mattermost.com)",
	})
	validConfiguration := Configuration{links}

	api := &plugintest.API{Store: &plugintest.KeyValueStore{}}

	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.Configuration")).Return(func(dest interface{}) error {
		*dest.(*Configuration) = validConfiguration
		return nil
	})

	p := Plugin{}
	p.OnActivate(api)

	post := &model.Post{Message: "Welcome to Mattermost!"}
	rpost, _ := p.MessageWillBePosted(post)

	if rpost.Message != "Welcome to [Mattermost](https://mattermost.com)!" {
		t.Fatal("Posted didn't get transformed")
	}
}
