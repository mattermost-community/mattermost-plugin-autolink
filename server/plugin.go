package main

import (
	"fmt"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/mlog"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/utils/markdown"
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

func (p *Plugin) processPost(c *plugin.Context, post *model.Post) (*model.Post, string) {
	links := p.links.Load().([]*AutoLinker)
	postText := post.Message
	offset := 0

	doReplacements := func(nodeText string, textRange markdown.Range) {
		startPos, endPos := textRange.Position+offset, textRange.End+offset
		origText := postText[startPos:endPos]
		if nodeText != origText {
			// TODO: ignore if the difference is because of http:// prefix
			mlog.Error(fmt.Sprintf("Markdown text did not match range text, '%s' != '%s'", nodeText, origText))
			return
		}

		newText := origText
		for _, l := range links {
			newText = l.Replace(newText)
		}

		if origText != newText {
			postText = postText[:startPos] + newText + postText[endPos:]
			offset += len(newText) - len(origText)
		}
	}

	markdown.Inspect(post.Message, func(node interface{}) bool {
		switch thisnode := node.(type) {
		// never descend into the text content of a link/image
		case *markdown.InlineLink:
			return false
		case *markdown.InlineImage:
			return false
		case *markdown.ReferenceLink:
			return false
		case *markdown.ReferenceImage:
			return false
		case *markdown.Text:
			doReplacements(thisnode.Text, thisnode.Range)
		case *markdown.Autolink:
			doReplacements(thisnode.Destination(), thisnode.RawDestination)
		}

		return true
	})
	post.Message = postText

	return post, ""
}

// MessageWillBePosted is invoked when a message is posted by a user before it is committed
// to the database.
func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	return p.processPost(c, post)
}

// MessageWillBeUpdated is invoked when a message is updated by a user before it is committed
// to the database.
func (p *Plugin) MessageWillBeUpdated(c *plugin.Context, post *model.Post, _ *model.Post) (*model.Post, string) {
	return p.processPost(c, post)
}
