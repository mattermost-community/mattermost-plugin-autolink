module github.com/mattermost/mattermost-plugin-autolink

go 1.16

require (
	github.com/gorilla/mux v1.8.0
	github.com/mattermost/mattermost-plugin-api v0.0.15
	github.com/mattermost/mattermost-plugin-apps v0.6.1-0.20210613215434-e58b800b61c7
	github.com/mattermost/mattermost-server/v5 v5.3.2-0.20210618105530-04b893f0155d
	github.com/mholt/archiver/v3 v3.5.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
)

replace github.com/mattermost/mattermost-plugin-apps => /home/sumacheb/src/mattermost/plugins/apps
