# mattermost-plugin-autolink

This plugin allows you to create various regular expression patterns that will be reformatted into a markdown link before the post is saved into the database.

## Installation

Go to the [releases page of this Github repository](https://github.com/mattermost/mattermost-plugin-autlink/releases) and download the latest release for your server architecture. You can upload this file in the Mattermost system console to install the plugin.

You'll need to modify your config.json to include the types of regexp patterns you wish to match.  You'll need to add a section undernieth `PluginSettings` for this plugin.  Below you'll find an example of what this should look like

```JSON
"PluginSettings": {
        ...
        "Plugins": {
            "mattermost-autolink": {
                "links": [
                    {
                        "Pattern": "(LHS)",
                        "Template": "[LHS](https://docs.mattermost.com/process/training.html#lhs)"
                    },
                    {
                        "Pattern": "(RHS)",
                        "Template": "[RHS](https://docs.mattermost.com/process/training.html#rhs)"
                    },
                    {
                        "Pattern": "(?i)(Mana)",
                        "Template": "[Mana](https://docs.mattermost.com/process/training.html#mana)"
                    },
                    {
                        "Pattern": "(?i)(ESR)",
                        "Template": "[ESR](https://docs.mattermost.com/process/training.html#esr)"
                    },
                    {
                        "Pattern": "((?P\u003clevel\u003e0|1|2|3|4|5)/5)",
                        "Template": "[${level}/5](https://docs.mattermost.com/process/training.html#id8)"
                    },
                    {
                        "Pattern": "(MM)(-)(?P\u003cjira_id\u003e\\d+)",
                        "Template": "[MM-${jira_id}](https://mattermost.atlassian.net/browse/MM-${jira_id})"
                    },
                    {
                        "Pattern": "https://pre-release\\.mattermost\\.com/core/pl/(?P\u003cid\u003e[a-zA-Z0-9]+)",
                        "Template": "[\u003cjump to convo\u003e](https://pre-release.mattermost.com/core/pl/${id})"
                    },
                    {
                        "Pattern": "(https://mattermost\\.atlassian\\.net/browse/)(MM)(-)(?P\u003cjira_id\u003e\\d+)",
                        "Template": "[MM-${jira_id}](https://mattermost.atlassian.net/browse/MM-${jira_id})"
                    },
                    {
                        "Pattern": "https://github\\.com/mattermost/(?P\u003crepo\u003e.+)/pull/(?P\u003cid\u003e\\d+)",
                        "Template": "[pr-${repo}-${id}](https://github.com/mattermost/${repo}/pull/${id})"
                    },
                    {
                        "Pattern": "https://github\\.com/mattermost/(?P\u003crepo\u003e.+)/issues/(?P\u003cid\u003e\\d+)",
                        "Template": "[issue-${repo}-${id}](https://github.com/mattermost/${repo}/issues/${id})"
                    },
                    {
                        "Pattern": "(PLT)(-)(?P\u003cjira_id\u003e\\d+)",
                        "Template": "[PLT-${jira_id}](https://mattermost.atlassian.net/browse/PLT-${jira_id})"
                    },
                    {
                        "Pattern": "(https://mattermost\\.atlassian\\.net/browse/)(PLT)(-)(?P\u003cjira_id\u003e\\d+)",
                        "Template": "[PLT-${jira_id}](https://mattermost.atlassian.net/browse/PLT-${jira_id})"
                    }
                ]
            },
        },
    ...
    }
```