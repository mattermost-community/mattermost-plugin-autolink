package autolinkplugin

import "github.com/mattermost/mattermost-plugin-autolink/server/autolink"

// GetPreConFigLinks gets the preconfigured links from the plugin config
func (p *Plugin) AddPreConfigLinks(c *Config) []autolink.Autolink {

	var link autolink.Autolink
	if c.EnableZenDesk {
		link = autolink.Autolink{
			Name:     "SSN",
			Pattern:  "(?P<SSN>(?P<part1>\\d{3})[ -]?(?P<part2>\\d{2})[ -]?(?P<LastFour>[0-9]{4}))",
			Template: "XXX-XX-$LastFour",
			Disabled: !c.EnableZenDesk,
		}
		c.Links = append(c.Links, link)
	}

	return c.Links
}
