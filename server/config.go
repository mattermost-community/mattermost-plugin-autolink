package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/pkg/errors"
)

// Config from config.json
type Config struct {
	EnableAdminCommand bool
	Links              []Link
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var c Config
	err := p.API.LoadPluginConfiguration(&c)
	if err != nil {
		return err
	}

	for i := range c.Links {
		err = c.Links[i].Compile()
		if err != nil {
			mlog.Error(fmt.Sprintf("Error creating autolinker: %+v: %v", c.Links[i], err))
		}
	}

	p.updateConfig(func(conf *Config) {
		*conf = c
	})

	if c.EnableAdminCommand {
		err := p.API.RegisterCommand(&model.Command{
			Trigger:          "autolink",
			DisplayName:      "Autolink",
			Description:      "Autolink administration.",
			AutoComplete:     true,
			AutoCompleteHint: "[command]",
		})
		if err != nil {
			return errors.WithMessage(err, "failed to register /autolink command")
		}
	} else {
		_ = p.API.UnregisterCommand("", "autolink")
	}

	return nil
}

func (p *Plugin) getConfig() Config {
	p.confLock.RLock()
	defer p.confLock.RUnlock()
	return p.conf
}

func (p *Plugin) updateConfig(f func(conf *Config)) Config {
	p.confLock.Lock()
	defer p.confLock.Unlock()

	f(&p.conf)
	return p.conf
}

// ToConfig marshals Config into a tree of map[string]interface{} to pass down
// to p.API.SavePluginConfig, otherwise RPC/gob barfs at the unknown type.
func (conf Config) ToConfig() map[string]interface{} {
	links := []interface{}{}
	for _, l := range conf.Links {
		links = append(links, l.ToConfig())
	}
	return map[string]interface{}{
		"Links": links,
	}
}

// Sorted returns a clone of the Config, with links sorted alphabetically
func (conf Config) Sorted() Config {
	sorted := conf
	sorted.Links = append([]Link{}, conf.Links...)
	sort.Slice(conf.Links, func(i, j int) bool {
		return strings.Compare(conf.Links[i].DisplayName(), conf.Links[j].DisplayName()) < 0
	})
	return conf
}
