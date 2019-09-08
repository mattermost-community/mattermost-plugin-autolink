package main

import (
	"fmt"
	"sync"

	"github.com/mattermost/mattermost-server/mlog"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/utils/markdown"
)

// Plugin the main struct for everything
type Plugin struct {
	plugin.MattermostPlugin

	// configuration and a muttex to control concurrent access
	conf     Config
	confLock sync.RWMutex
}

// MessageWillBePosted is invoked when a message is posted by a user before it is committed
// to the database.
func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	conf := p.getConfig()
	postText := post.Message
	offset := 0
	markdown.Inspect(post.Message, func(node interface{}) bool {
		switch node.(type) {
		// never descend into the text content of a link/image
		case *markdown.InlineLink:
			return false
		case *markdown.InlineImage:
			return false
		case *markdown.ReferenceLink:
			return false
		case *markdown.ReferenceImage:
			return false
		}

		origText := ""
		startPos := 0
		endPos := 0

		if autolinkNode, ok := node.(*markdown.Autolink); ok {
			startPos, endPos = autolinkNode.RawDestination.Position+offset, autolinkNode.RawDestination.End+offset
			origText = postText[startPos:endPos]
			if autolinkNode.Destination() != origText {
				mlog.Error(fmt.Sprintf("Markdown autolink did not match range text, '%s' != '%s'", autolinkNode.Destination(), origText))
				return true
			}
		} else if textNode, ok := node.(*markdown.Text); ok {
			startPos, endPos = textNode.Range.Position+offset, textNode.Range.End+offset
			origText = postText[startPos:endPos]
			if textNode.Text != origText {
				mlog.Error(fmt.Sprintf("Markdown text did not match range text, '%s' != '%s'", textNode.Text, origText))
				return true
			}
		}

		if origText != "" {
			newText := origText

			channel, cErr := p.API.GetChannel(post.ChannelId)
			if cErr != nil {
				mlog.Error(cErr.Error())
				return false
			}
			teamName := ""
			if channel.TeamId != "" {
				team, tErr := p.API.GetTeam(channel.TeamId)
				if tErr != nil {
					mlog.Error(tErr.Error())
					return false
				}
				teamName = team.Name
			}

			for _, l := range conf.Links {
				if len(l.Scope) == 0 {
					newText = l.Replace(newText)
				} else if teamName != "" && contains(teamName, channel.Name, l.Scope) {
					newText = l.Replace(newText)
				}
			}
			if origText != newText {
				postText = postText[:startPos] + newText + postText[endPos:]
				offset += len(newText) - len(origText)
			}
		}

		return true
	})
	post.Message = postText

	post.Hashtags, _ = model.ParseHashtags(post.Message)

	return post, ""
}
