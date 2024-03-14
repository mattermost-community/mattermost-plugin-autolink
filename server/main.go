package main

import (
	"github.com/mattermost/mattermost/server/public/plugin"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolinkplugin"
)

func main() {
	plugin.ClientMain(autolinkplugin.New())
}
