package autolinkplugin

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
)

const helpText = "###### Mattermost Autolink Plugin Administration\n" +
	"<linkref> is either the Name of a link, or its number in the `/autolink list` output. A partial Name can be specified, but some commands require it to be uniquely resolved.\n" +
	"* `/autolink add <name>` - add a new link, named <name>.\n" +
	"* `/autolink delete <linkref>` - delete a link.\n" +
	"* `/autolink disable <linkref>` - disable a link.\n" +
	"* `/autolink enable <linkref>` - enable a link.\n" +
	"* `/autolink list <linkref>` - list a specific link.\n" +
	"* `/autolink list` - list all configured links.\n" +
	"* `/autolink set <linkref> <field> value...` - sets a link's field to a value. The entire command line after <field> is used for the value, unescaped, leading/trailing whitespace trimmed.\n" +
	"* `/autolink test <linkref> test-text...` - test a link on a sample.\n" +
	"\n" +
	"Example:\n" +
	"```\n" +
	"/autolink add Visa\n" +
	"/autolink disable Visa\n" +
	`/autolink set Visa Pattern (?P<VISA>(?P<part1>4\d{3})[ -]?(?P<part2>\d{4})[ -]?(?P<part3>\d{4})[ -]?(?P<LastFour>[0-9]{4}))` + "\n" +
	"/autolink set Visa Template VISA XXXX-XXXX-XXXX-$LastFour\n" +
	"/autolink set Visa WordMatch true\n" +
	"/autolink set Visa Scope team/townsquare\n" +
	"/autolink test Vi 4356-7891-2345-1111 -- (4111222233334444)\n" +
	"/autolink enable Visa\n" +
	"```\n" +
	""

type CommandHandlerFunc func(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse

type CommandHandler struct {
	handlers       map[string]CommandHandlerFunc
	defaultHandler CommandHandlerFunc
}

var autolinkCommandHandler = CommandHandler{
	handlers: map[string]CommandHandlerFunc{
		"help":    executeHelp,
		"list":    executeList,
		"delete":  executeDelete,
		"disable": executeDisable,
		"enable":  executeEnable,
		"add":     executeAdd,
		"set":     executeSet,
		"test":    executeTest,
	},
	defaultHandler: executeHelp,
}

func (ch CommandHandler) Handle(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	for n := len(args); n > 0; n-- {
		h := ch.handlers[strings.Join(args[:n], "/")]
		if h != nil {
			return h(p, c, header, args[n:]...)
		}
	}
	return ch.defaultHandler(p, c, header, args...)
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, commandArgs *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	isAdmin, err := p.IsAuthorizedAdmin(commandArgs.UserId)
	if err != nil {
		return responsef("error occurred while authorizing the command: %v", err), nil
	}
	if !isAdmin {
		return responsef("`/autolink` commands can only be executed by a system administrator or `autolink` plugin admins."), nil
	}

	args := strings.Fields(commandArgs.Command)
	if len(args) == 0 || args[0] != "/autolink" {
		return responsef(helpText), nil
	}

	return autolinkCommandHandler.Handle(p, c, commandArgs, args[1:]...), nil
}

func executeList(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	links, refs, err := parseLinkRef(p, false, args...)
	if err != nil {
		return responsef("%v", err)
	}

	text := ""
	if len(refs) > 0 {
		for _, i := range refs {
			text += links[i].ToMarkdown(i + 1)
		}
	} else {
		for i, l := range links {
			text += l.ToMarkdown(i + 1)
		}
	}
	return responsef(text)
}

func executeDelete(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) != 1 {
		return responsef(helpText)
	}
	oldLinks, refs, err := parseLinkRef(p, true, args...)
	if err != nil {
		return responsef("%v", err)
	}
	n := refs[0]

	removed := oldLinks[n]
	newLinks := oldLinks[:n]
	if n+1 < len(oldLinks) {
		newLinks = append(newLinks, oldLinks[n+1:]...)
	}

	err = saveConfigLinks(p, newLinks)
	if err != nil {
		return responsef(err.Error())
	}

	return responsef("removed: \n%v", removed.ToMarkdown(0))
}

func executeSet(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) < 3 {
		return responsef(helpText)
	}

	links, refs, err := parseLinkRef(p, true, args...)
	if err != nil {
		return responsef("%v", err)
	}
	l := &links[refs[0]]

	fieldName := args[1]
	restOfCommand := header.Command[10:] // "/autolink "
	restOfCommand = restOfCommand[strings.Index(restOfCommand, args[0])+len(args[0]):]
	restOfCommand = restOfCommand[strings.Index(restOfCommand, args[1])+len(args[1]):]
	value := strings.TrimSpace(restOfCommand)

	switch fieldName {
	case "Name":
		l.Name = value
	case "Pattern":
		l.Pattern = value
	case "Template":
		l.Template = value
	case "Scope":
		l.Scope = args[2:]
	case "DisableNonWordPrefix":
		boolValue, e := parseBoolArg(value)
		if e != nil {
			return responsef("%v", e)
		}
		l.DisableNonWordPrefix = boolValue
	case "DisableNonWordSuffix":
		boolValue, e := parseBoolArg(value)
		if e != nil {
			return responsef("%v", e)
		}
		l.DisableNonWordSuffix = boolValue
	case "WordMatch":
		boolValue, e := parseBoolArg(value)
		if e != nil {
			return responsef("%v", e)
		}
		l.WordMatch = boolValue
	case "Disabled":
		boolValue, e := parseBoolArg(value)
		if e != nil {
			return responsef("%v", e)
		}
		l.Disabled = boolValue
	default:
		return responsef("%q is not a supported field, must be one of %q", fieldName,
			[]string{"Name", "Disabled", "Pattern", "Template", "Scope", "DisableNonWordPrefix", "DisableNonWordSuffix", "WordMatch"})
	}

	err = saveConfigLinks(p, links)
	if err != nil {
		return responsef(err.Error())
	}

	ref := args[0]
	if l.Name != "" {
		ref = l.Name
	}
	return executeList(p, c, header, ref)
}

func executeTest(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) < 2 {
		return responsef(helpText)
	}

	links, refs, err := parseLinkRef(p, false, args...)
	if err != nil {
		return responsef("%v", err)
	}

	restOfCommand := header.Command[10:] // "/autolink "
	restOfCommand = restOfCommand[strings.Index(restOfCommand, args[0])+len(args[0]):]
	orig := strings.TrimSpace(restOfCommand)
	out := fmt.Sprintf("- Original: `%s`\n", orig)

	for _, ref := range refs {
		l := links[ref]
		l.Disabled = false
		err = l.Compile()
		if err != nil {
			return responsef("failed to compile link %s: %v", l.DisplayName(), err)
		}
		replaced := l.Replace(orig)
		if replaced == orig {
			out += fmt.Sprintf("- Link %s: _no change_\n", l.DisplayName())
		} else {
			out += fmt.Sprintf("- Link %s: changed to `%s`\n", l.DisplayName(), replaced)
			orig = replaced
		}
	}

	return responsef(out)
}

func executeEnable(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) != 1 {
		return responsef(helpText)
	}
	return executeEnableImpl(p, c, header, args[0], true)
}

func executeDisable(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) != 1 {
		return responsef(helpText)
	}
	return executeEnableImpl(p, c, header, args[0], false)
}

func executeEnableImpl(p *Plugin, c *plugin.Context, header *model.CommandArgs, ref string, enabled bool) *model.CommandResponse {
	links, refs, err := parseLinkRef(p, true, ref)
	if err != nil {
		return responsef("%v", err)
	}
	l := &links[refs[0]]
	l.Disabled = !enabled

	err = saveConfigLinks(p, links)
	if err != nil {
		return responsef(err.Error())
	}

	if l.Name != "" {
		ref = l.Name
	}
	return executeList(p, c, header, ref)
}

func executeAdd(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) > 1 {
		return responsef(helpText)
	}
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	err := saveConfigLinks(p, append(p.getConfig().Links, autolink.Autolink{
		Name: name,
	}))
	if err != nil {
		return responsef(err.Error())
	}

	if name == "" {
		return executeList(p, c, header)
	}
	return executeList(p, c, header, name)
}

func executeHelp(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	return responsef(helpText)
}

func responsef(format string, args ...interface{}) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text:         fmt.Sprintf(format, args...),
		Type:         model.POST_DEFAULT,
	}
}

func parseLinkRef(p *Plugin, requireUnique bool, args ...string) ([]autolink.Autolink, []int, error) {
	links := p.getConfig().Sorted().Links
	if len(args) == 0 {
		if requireUnique {
			return nil, nil, errors.New("unreachable")
		}

		return links, nil, nil
	}

	n, err := strconv.ParseUint(args[0], 10, 32)
	if err == nil {
		if n < 1 || int(n) > len(links) {
			return nil, nil, errors.Errorf("%v is not a valid link number.", n)
		}
		return links, []int{int(n) - 1}, nil
	}

	found := []int{}
	for i, l := range links {
		if strings.Contains(l.Name, args[0]) {
			found = append(found, i)
		}
	}
	if len(found) == 0 {
		return nil, nil, errors.Errorf("%q not found.", args[0])
	}
	if requireUnique && len(found) > 1 {
		names := []string{}
		for _, i := range found {
			names = append(names, links[i].Name)
		}
		return nil, nil, errors.Errorf("%q matched more than one link: %q", args[0], names)
	}

	return links, found, nil
}

func parseBoolArg(arg string) (bool, error) {
	switch strings.ToLower(arg) {
	case "true", "on":
		return true, nil
	case "false", "off":
		return false, nil
	}
	return false, errors.Errorf("Not a bool, %q", arg)
}

func saveConfigLinks(p *Plugin, links []autolink.Autolink) error {
	conf := p.getConfig()
	conf.Links = links
	appErr := p.API.SavePluginConfig(conf.ToConfig())
	if appErr != nil {
		return appErr
	}
	return nil
}
