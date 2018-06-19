# Autolink Plugin

This plugin allows you to create various regular expression patterns that will be reformatted into a markdown link before the post is saved into the database.

## Installation

Go to the [releases page of this Github repository](https://github.com/mattermost/mattermost-plugin-autolink/releases) and download the latest release for your server architecture. You can upload this file in the Mattermost system console to install the plugin.

You'll need to modify your config.json to include the types of regexp patterns you wish to match.  You'll need to add a section undernieth `PluginSettings` for this plugin.  Below you'll find an example of what this should look like

Autolinks have 2 parts.  A `Pattern` which is a regular expression search pattern utilizing the https://golang.org/pkg/regexp/ library and a `Template` that gets exanded.  You can create variables in the pattern with the syntax `(?P<name>...)` which will then be expanded by the corresponding template.  In the template, a variable is denoted by a substring of the form $name or ${name}, where name is a non-empty sequence of letters, digits, and underscores.  A purely numeric name like $1 refers to the submatch with the corresponding index.  In the $name form, name is taken to be as long as possible: $1x is equivalent to ${1x}, not ${1}x, and, $10 is equivalent to ${10}, not ${1}0.  To insert a literal $ in the output, use $$ in the template.

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
                    "Pattern": "((?P<level>0|1|2|3|4|5)/5)",
                    "Template": "[${level}/5](https://docs.mattermost.com/process/training.html#id8)"
                },
                {
                    "Pattern": "(MM)(-)(?P<jira_id>\\d+)",
                    "Template": "[MM-${jira_id}](https://mattermost.atlassian.net/browse/MM-${jira_id})"
                },
                {
                    "Pattern": "https://pre-release\\.mattermost\\.com/core/pl/(?P<id>[a-zA-Z0-9]+)",
                    "Template": "[<jump to convo>](https://pre-release.mattermost.com/core/pl/${id})"
                },
                {
                    "Pattern": "(https://mattermost\\.atlassian\\.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
                    "Template": "[MM-${jira_id}](https://mattermost.atlassian.net/browse/MM-${jira_id})"
                },
                {
                    "Pattern": "https://github\\.com/mattermost/(?P<repo>.+)/pull/(?P<id>\\d+)",
                    "Template": "[pr-${repo}-${id}](https://github.com/mattermost/${repo}/pull/${id})"
                },
                {
                    "Pattern": "https://github\\.com/mattermost/(?P<repo>.+)/issues/(?P<id>\\d+)",
                    "Template": "[issue-${repo}-${id}](https://github.com/mattermost/${repo}/issues/${id})"
                },
                {
                    "Pattern": "(PLT)(-)(?P<jira_id>\\d+)",
                    "Template": "[PLT-${jira_id}](https://mattermost.atlassian.net/browse/PLT-${jira_id})"
                },
                {
                    "Pattern": "(https://mattermost\\.atlassian\\.net/browse/)(PLT)(-)(?P<jira_id>\\d+)",
                    "Template": "[PLT-${jira_id}](https://mattermost.atlassian.net/browse/PLT-${jira_id})"
                }
            ]
        },
    },
    ...
}
```
