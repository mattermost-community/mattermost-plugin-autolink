package main

import (
	"sync/atomic"

	"github.com/mattermost/mattermost-server/mlog"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

// Plugin the main struct for everything
type Plugin struct {
	api   plugin.API
	links atomic.Value
}

// OnActivate is invoked when the plugin is activated.
func (p *Plugin) OnActivate(api plugin.API) error {
	p.api = api

	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	return nil
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var c Configuration
	err := p.api.LoadPluginConfiguration(&c)
	if err != nil {
		return err
	}

	links := make([]*AutoLinker, 0)

	for _, l := range c.Links {
		al, lerr := NewAutoLinker(l)
		if lerr != nil {
			mlog.Error("Error creating autolinker: ")
		}

		links = append(links, al)
	}

	p.links.Store(links)
	return nil
}

// MessageWillBePosted is invoked when a message is posted by a user before it is commited
// to the database.
func (p *Plugin) MessageWillBePosted(post *model.Post) (*model.Post, string) {
	links := p.links.Load().([]*AutoLinker)

	for _, l := range links {
		post.Message = l.Replace(post.Message)
	}

	return post, ""
}
