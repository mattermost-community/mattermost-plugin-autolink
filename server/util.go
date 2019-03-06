package main

import (
	"strings"

	"github.com/mattermost/mattermost-server/mlog"
)

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
