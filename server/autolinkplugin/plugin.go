package autolinkplugin

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/shared/markdown"
	"github.com/pkg/errors"

	"github.com/mattermost-community/mattermost-plugin-autolink/server/api"
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

func (p *Plugin) ProcessPost(_ *plugin.Context, post *model.Post) (*model.Post, string) {
	conf := p.getConfig()

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

	// Author (bot) resolution is shared across the message body and any
	// attachments, resolved lazily on first need.
	var author *model.User
	var authorErr *model.AppError
	authorResolved := false
	isBotAuthor := func() bool {
		if !authorResolved {
			author, authorErr = p.API.GetUser(post.UserId)
			authorResolved = true
			if authorErr != nil {
				// NOTE: Not sure how we want to handle errors here, we can either:
				// * assume that occasional rewrites of Bot messges are ok
				// * assume that occasional not rewriting of all messages is ok
				// Let's assume for now that former is a lesser evil and carry on.
				p.API.LogError("failed to check if message for rewriting was send by a bot", "error", authorErr)
			}
		}
		return author != nil && author.IsBot
	}

	if message, changed := p.autolinkText(post.Message, conf, channelName, teamName, isBotAuthor, post.Id); changed {
		post.Message = message
		post.Hashtags, _ = model.ParseHashtags(message)
	}

	// Integrations such as Gitea, Jira and GitHub deliver their payload as
	// message attachments, so the linkable tokens live in attachment text
	// rather than in post.Message. Rewrite the visible attachment fields too.
	attachments := post.Attachments()
	attachmentsChanged := false
	for _, attachment := range attachments {
		if v, changed := p.autolinkText(attachment.Pretext, conf, channelName, teamName, isBotAuthor, post.Id); changed {
			attachment.Pretext = v
			attachmentsChanged = true
		}
		if v, changed := p.autolinkText(attachment.Title, conf, channelName, teamName, isBotAuthor, post.Id); changed {
			attachment.Title = v
			attachmentsChanged = true
		}
		if v, changed := p.autolinkText(attachment.Text, conf, channelName, teamName, isBotAuthor, post.Id); changed {
			attachment.Text = v
			attachmentsChanged = true
		}
		for _, field := range attachment.Fields {
			if s, ok := field.Value.(string); ok {
				if v, changed := p.autolinkText(s, conf, channelName, teamName, isBotAuthor, post.Id); changed {
					field.Value = v
					attachmentsChanged = true
				}
			}
		}
	}
	if attachmentsChanged {
		post.AddProp("attachments", attachments)
	}

	return post, ""
}

// autolinkText applies the configured link rules to a single markdown string,
// returning the rewritten text and whether anything changed. It backs both the
// post body and message-attachment rewriting so the two share identical
// matching, scope and bot-author semantics.
func (p *Plugin) autolinkText(text string, conf *Config, channelName, teamName string, isBotAuthor func() bool, postID string) (string, bool) {
	if text == "" {
		return text, false
	}

	message := text
	changed := false
	offset := 0

	markdown.Inspect(text, func(node interface{}) bool {
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
				p.API.LogDebug("skipping escaped autolink", "original", toProcess, "post_id", postID)
				return true
			}

		case *markdown.Text:
			start, end = node.Range.Position+offset, node.Range.End+offset
			toProcess = message[start:end]
			if node.Text != toProcess {
				p.API.LogDebug("skipping text: parsed markdown did not match original", "parsed", node.Text, "original", toProcess, "post_id", postID)
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

			if !link.ProcessBotPosts && isBotAuthor() {
				continue
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

	return message, changed
}

func (p *Plugin) ServeHTTP(_ *plugin.Context, w http.ResponseWriter, r *http.Request) {
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
