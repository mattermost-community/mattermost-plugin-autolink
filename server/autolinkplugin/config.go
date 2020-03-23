package autolinkplugin

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
	"github.com/mattermost/mattermost-server/v5/mlog"
	"github.com/mattermost/mattermost-server/v5/model"
)

// Config from config.json
type Config struct {
	EnableAdminCommand bool
	EnableOnUpdate     bool
	PluginAdmins       string
	Links              []autolink.Autolink

	// AdminUserIds is a set of UserIds that are permitted to perform
	// administrative operations on the plugin configuration (i.e. plugin
	// admins). On each configuration change the contents of PluginAdmins
	// config field is parsed into this field.
	AdminUserIds map[string]struct{}
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var c Config
	if err := p.API.LoadPluginConfiguration(&c); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	for i := range c.Links {
		if err := c.Links[i].Compile(); err != nil {
			p.API.LogError("Error creating autolinker", "link", c.Links[i], "error", err.Error())
		}
	}

	// Plugin admin UserId parsing and validation errors are
	// not fatal, if everything fails only sysadmin will be able to manage the
	// config which is still OK
	c.parsePluginAdminList(p)

	p.UpdateConfig(func(conf *Config) {
		*conf = c
	})

	go func() {
		if c.EnableAdminCommand {
			_ = p.API.RegisterCommand(&model.Command{
				Trigger:          "autolink",
				DisplayName:      "Autolink",
				Description:      "Autolink administration.",
				AutoComplete:     true,
				AutoCompleteDesc: "Available commands: add, delete, disable, enable, list, set, test",
				AutoCompleteHint: "[command]",
			})
		} else {
			_ = p.API.UnregisterCommand("", "autolink")
		}
	}()

	return nil
}

func (p *Plugin) getConfig() *Config {
	p.confLock.RLock()
	defer p.confLock.RUnlock()

	return p.conf
}

func (p *Plugin) GetLinks() []autolink.Autolink {
	p.confLock.RLock()
	defer p.confLock.RUnlock()

	return p.conf.Links
}

func (p *Plugin) SaveLinks(links []autolink.Autolink) error {
	p.UpdateConfig(func(conf *Config) {
		conf.Links = links
	})
	appErr := p.API.SavePluginConfig(p.getConfig().ToConfig())
	if appErr != nil {
		return fmt.Errorf("Unable to save links: %w", appErr)
	}

	return nil
}

func (p *Plugin) UpdateConfig(f func(conf *Config)) {
	p.confLock.Lock()
	defer p.confLock.Unlock()

	f(p.conf)
}

// ToConfig marshals Config into a tree of map[string]interface{} to pass down
// to p.API.SavePluginConfig, otherwise RPC/gob barfs at the unknown type.
func (conf *Config) ToConfig() map[string]interface{} {
	links := []interface{}{}
	for _, l := range conf.Links {
		links = append(links, l.ToConfig())
	}
	return map[string]interface{}{
		"EnableAdminCommand": conf.EnableAdminCommand,
		"EnableOnUpdate":     conf.EnableOnUpdate,
		"PluginAdmins":       conf.PluginAdmins,
		"Links":              links,
	}
}

// Sorted returns a clone of the Config, with links sorted alphabetically
func (conf *Config) Sorted() *Config {
	sorted := conf
	sorted.Links = append([]autolink.Autolink{}, conf.Links...)
	sort.Slice(conf.Links, func(i, j int) bool {
		return strings.Compare(conf.Links[i].DisplayName(), conf.Links[j].DisplayName()) < 0
	})
	return conf
}

// parsePluginAdminList parses the contents of PluginAdmins config field
func (conf *Config) parsePluginAdminList(p *Plugin) {
	conf.AdminUserIds = make(map[string]struct{})

	if len(conf.PluginAdmins) == 0 {
		// There were no plugin admin users defined
		return
	}

	userIDs := strings.Split(conf.PluginAdmins, ",")

	for _, v := range userIDs {
		userId := strings.TrimSpace(v)
		// Let's verify that the given user really exists
		_, appErr := p.API.GetUser(userId)
		if appErr != nil {
			mlog.Error(fmt.Sprintf(
				"error occured while verifying userId %s: %v", v, appErr))
		} else {
			conf.AdminUserIds[userId] = struct{}{}
		}
	}
}
