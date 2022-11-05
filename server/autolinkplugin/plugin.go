package autolinkplugin

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/plugin"
	"github.com/mattermost/mattermost-server/v6/shared/markdown"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-autolink/server/api"
)

// Plugin the main struct for everything
type Plugin struct {
	plugin.MattermostPlugin

	handler *api.Handler

	// configuration and a mutex to control concurrent access
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

func (p *Plugin) IsAuthorizedAdmin(userID string) (bool, error) {
	user, err := p.API.GetUser(userID)
	if err != nil {
		return false, errors.Wrapf(err, "failed to obtain information about user `%s`", userID)
	}
	if strings.Contains(user.Roles, "system_admin") {
		p.API.LogInfo(
			fmt.Sprintf("UserID `%s` is authorized basing on the sysadmin role membership", userID))
		return true, nil
	}

	conf := p.getConfig()
	if _, ok := conf.AdminUserIds[userID]; ok {
		p.API.LogInfo(
			fmt.Sprintf("UserID `%s` is authorized basing on the list of plugin admins list", userID))
		return true, nil
	}

	return false, nil
}

func (p *Plugin) resolveScope(channelID string) (string, string, *model.AppError) {
	channel, cErr := p.API.GetChannel(channelID)
	if cErr != nil {
		return "", "", cErr
	}

	if channel.TeamId == "" {
		return channel.Name, "", nil
	}

	team, tErr := p.API.GetTeam(channel.TeamId)
	if tErr != nil {
		return "", "", tErr
	}

	return channel.Name, team.Name, nil
}

func (p *Plugin) inScope(scope []string, channelName string, teamName string) bool {
	if len(scope) == 0 {
		return true
	}

	if teamName == "" {
		return false
	}

	for _, teamChannel := range scope {
		split := strings.Split(teamChannel, "/")

		splitLength := len(split)

		if splitLength == 1 && split[0] == "" {
			return false
		}

		if splitLength == 1 && strings.EqualFold(split[0], teamName) {
			return true
		}

		scopeMatch := strings.EqualFold(split[0], teamName) && strings.EqualFold(split[1], channelName)
		if splitLength == 2 && scopeMatch {
			return true
		}
	}

	return false
}

func (p *Plugin) ProcessPost(c *plugin.Context, post *model.Post) (*model.Post, string) {
	conf := p.getConfig()

	message := post.Message
	changed := false
	offset := 0

	hasOneOrMoreScopes := false
	for _, link := range conf.Links {
		if len(link.Scope) > 0 {
			hasOneOrMoreScopes = true
			break
		}
	}

	channelName := ""
	teamName := ""
	if hasOneOrMoreScopes {
		cn, tn, rsErr := p.resolveScope(post.ChannelId)
		channelName = cn
		teamName = tn

		if rsErr != nil {
			p.API.LogError("Failed to resolve scope", "error", rsErr.Error())
		}
	}

	var author *model.User
	var authorErr *model.AppError

	markdown.Inspect(post.Message, func(node interface{}) bool {
		if node == nil {
			return false
		}

		toProcess, start, end := "", 0, 0
		switch node := node.(type) {
		// never descend into the text content of a link/image
		case *markdown.InlineLink, *markdown.InlineImage, *markdown.ReferenceLink, *markdown.ReferenceImage:
			return false

		case *markdown.Autolink:
			start, end = node.RawDestination.Position+offset, node.RawDestination.End+offset
			toProcess = message[start:end]
			// Do not process escaped links. Not exactly sure why but preserving the previous behavior.
			// https://mattermost.atlassian.net/browse/MM-42669
			if markdown.Unescape(toProcess) != toProcess {
				p.API.LogDebug("skipping escaped autolink", "original", toProcess, "post_id", post.Id)
				return true
			}

		case *markdown.Text:
			start, end = node.Range.Position+offset, node.Range.End+offset
			toProcess = message[start:end]
			if node.Text != toProcess {
				p.API.LogDebug("skipping text: parsed markdown did not match original", "parsed", node.Text, "original", toProcess, "post_id", post.Id)
				return true
			}
		}

		if toProcess == "" {
			return true
		}

		processed := toProcess
		for _, link := range conf.Links {
			if !p.inScope(link.Scope, channelName, teamName) {
				continue
			}

			out := link.Replace(processed)
			if out == processed {
				continue
			}

			if !link.ProcessBotPosts {
				if author == nil && authorErr == nil {
					author, authorErr = p.API.GetUser(post.UserId)
					if authorErr != nil {
						// NOTE: Not sure how we want to handle errors here, we can either:
						// * assume that occasional rewrites of Bot messges are ok
						// * assume that occasional not rewriting of all messages is ok
						// Let's assume for now that former is a lesser evil and carry on.
						p.API.LogError("failed to check if message for rewriting was send by a bot", "error", authorErr)
					}
				}

				if author != nil && author.IsBot {
					continue
				}
			}

			processed = out
		}

		if toProcess != processed {
			message = message[:start] + processed + message[end:]
			offset += len(processed) - len(toProcess)
			changed = true
		}

		return true
	})

	if changed {
		post.Message = message
		post.Hashtags, _ = model.ParseHashtags(message)
	}
	return post, ""
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
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
	if !conf.EnableOnUpdate {
		return post, ""
	}

	return p.ProcessPost(c, post)
}
