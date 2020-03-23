package autolinkplugin

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-plugin-autolink/server/api"

	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/mattermost/mattermost-server/v5/utils/markdown"
)

// Plugin the main struct for everything
type Plugin struct {
	plugin.MattermostPlugin

	handler *api.Handler

	// configuration and a muttex to control concurrent access
	conf     *Config
	confLock sync.RWMutex
}

func New() *Plugin {
	return &Plugin{
		conf: new(Config),
	}
}

func (p *Plugin) OnActivate() error {
	p.handler = api.NewHandler(p, p)

	return nil
}

func (p *Plugin) IsAuthorizedAdmin(userId string) (bool, error) {
	user, err := p.API.GetUser(userId)
	if err != nil {
		return false, fmt.Errorf(
			"failed to obtain information about user `%s`: %w", userId, err)
	}
	if strings.Contains(user.Roles, "system_admin") {
		mlog.Info(
			fmt.Sprintf("UserId `%s` is authorized basing on the sysadmin role membership", userId))
		return true, nil
	}

	conf := p.getConfig()
	if _, ok := conf.AdminUserIds[userId]; ok {
		mlog.Info(
			fmt.Sprintf("UserId `%s` is authorized basing on the list of plugin admins list", userId))
		return true, nil
	}

	return false, nil
}

func contains(team string, channel string, list []string) bool {
	for _, channelTeam := range list {
		channelTeamSplit := strings.Split(channelTeam, "/")
		if len(channelTeamSplit) == 2 {
			if strings.EqualFold(channelTeamSplit[0], team) && strings.EqualFold(channelTeamSplit[1], channel) {
				return true
			}
		} else if len(channelTeamSplit) == 1 {
			if strings.EqualFold(channelTeamSplit[0], team) {
				return true
			}
		} else {
			mlog.Error("error splitting channel & team combination.")
		}

	}
	return false
}

func (p *Plugin) isBotUser(userId string) (bool, *model.AppError) {
	user, appErr := p.API.GetUser(userId)
	if appErr != nil {
		p.API.LogError("failed to check if message for rewriting was send by a bot", "error", appErr)
		return false, appErr
	}

	return user.IsBot, nil
}

func (p *Plugin) ProcessPost(c *plugin.Context, post *model.Post) (*model.Post, string) {
	isBot, appErr := p.isBotUser(post.UserId)
	if appErr != nil {
		// NOTE: Not sure how we want to handle errors here, we can either:
		// * assume that occasional rewrites of Bot messges are ok
		// * assume that occasional not rewriting of all messages is ok
		// Let's assume for now that former is a lesser evil and carry on.
	}
	if isBot {
		p.API.LogDebug("not rewriting message from bot", "userId", post.UserId)
		return nil, ""
	}

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

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	r.Header.Add("Mattermost-Plugin-ID", c.SourcePluginId)
	p.handler.ServeHTTP(w, r)
}

// MessageWillBePosted is invoked when a message is posted by a user before it is committed
// to the database.
func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	return p.ProcessPost(c, post)
}

// MessageWillBeUpdated is invoked when a message is updated by a user before it is committed
// to the database.
func (p *Plugin) MessageWillBeUpdated(c *plugin.Context, post *model.Post, _ *model.Post) (*model.Post, string) {
	conf := p.getConfig()
	if conf.EnableOnUpdate {
		return p.ProcessPost(c, post)
	} else {
		return post, ""
	}
}
