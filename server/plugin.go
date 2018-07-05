package main

import (
	"strings"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/mlog"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

// Plugin the main struct for everything
type Plugin struct {
	plugin.MattermostPlugin

	links atomic.Value
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var c Configuration
	err := p.API.LoadPluginConfiguration(&c)
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

// MessageWillBePosted is invoked when a message is posted by a user before it is committed
// to the database.
func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	links := p.links.Load().([]*AutoLinker)

	cbMessages := make([]string, 0)
	codeBlocks := strings.Split(post.Message, "```")
	codeBlockSkip := false

	// split and turn linker on/off if we think we're in a code block like ```hello```
	for _, cb := range codeBlocks {
		if codeBlockSkip {
			cbMessages = append(cbMessages, cb)
			codeBlockSkip = false
		} else {
			// split and turn linker on/off if we think we're in a inline code block like `hello`
			icbMessages := make([]string, 0)
			icodeBlocks := strings.Split(cb, "`")
			icodeBlockSkip := false
			for _, icb := range icodeBlocks {
				if icodeBlockSkip {
					icbMessages = append(icbMessages, icb)
					icodeBlockSkip = false
				} else {
					for _, l := range links {
						icb = l.Replace(icb)
					}
					icbMessages = append(icbMessages, icb)
					icodeBlockSkip = true
				}
			}

			cbMessages = append(cbMessages, strings.Join(icbMessages, "`"))
			codeBlockSkip = true
		}
	}

	post.Message = strings.Join(cbMessages, "```")

	// for _, l := range links {
	// 	post.Message = l.Replace(post.Message)
	// }

	return post, ""
}
