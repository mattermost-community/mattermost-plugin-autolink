package main

import (
	"github.com/mattermost/mattermost-plugin-autolink/server/autolinkplugin"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func main() {
	plugin.ClientMain(autolinkplugin.New())
}
