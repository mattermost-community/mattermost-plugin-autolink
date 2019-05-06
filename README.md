# Autolink Plugin (Beta) ![CircleCI branch](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-autolink/master.svg)

This plugin creates regular expression (regexp) patterns that are reformatted into a Markdown link before the message is saved into the database.

Use it to add custom auto-linking on your Mattermost system, such as adding links to your issue tracker based on the regexp patterns.

**Supported Mattermost Server Versions: 5.2+ for 0.3+. 5.0 and 5.1 for all versions before 0.3**

## Installation

1. Go to the [releases page of this GitHub repository](https://github.com/mattermost/mattermost-plugin-autolink) and download the latest release for your Mattermost server.
2. Upload this file in the Mattermost **System Console > Plugins > Management** page to install the plugin. To learn more about how to upload a plugin, [see the documentation](https://docs.mattermost.com/administration/plugins.html#plugin-uploads).
3. Modify your `config.json` file to include the types of regexp patterns you wish to match, under the `PluginSettings`. See below for an example of what this should look like.

## Usage

Autolinks have 2 parts: a **Pattern** which is a regular expression search pattern utilizing the https://golang.org/pkg/regexp/ library, and a **Template** that gets expanded. You can create variables in the pattern with the syntax `(?P<name>...)` which will then be expanded by the corresponding template.

In the template, a variable is denoted by a substring of the form `$name` or `${name}`, where `name` is a non-empty sequence of letters, digits, and underscores. A purely numeric name like $1 refers to the submatch with the corresponding index. In the $name form, name is taken to be as long as possible: $1x is equivalent to ${1x}, not ${1}x, and, $10 is equivalent to ${10}, not ${1}0. To insert a literal $ in the output, use $$ in the template.

Below is an example of regexp patterns used for autolinking, modified in the `config.json` file:

```
"PluginSettings": {
    ...
    "Plugins": {
        "mattermost-autolink": {
            "links": [
                {
                    "Pattern": "(LHS)",
                    "Template": "[LHS](https://docs.mattermost.com/process/training.html#lhs)",
                    "Scope": ["team/off-topic"]
                },
                {
                    "Pattern": "(RHS)",
                    "Template": "[RHS](https://docs.mattermost.com/process/training.html#rhs)",
                    "Scope": ["team/town-square"]
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
    "PluginStates": {
        ...
        "mattermost-autolink": {
            "Enable": true
        },
        ...
    }
},
```

## Examples

1. Autolinking `Ticket ####:text with alphanumberic characters and spaces` to a ticket link. Use:
  - Pattern: `(?i)(ticket )(?P<ticket_id>.+)(:)(?P<ticket_info>.*)`, or if the ticket_id is a number, then `(?i)(ticket )(?P<ticket_id>\d+)(:)(?P<ticket_info>.*)`
  - Template: `[Ticket ${ticket_id}: ${ticket_info}](https://github.com/mattermost/mattermost-server/issues/${ticket_id})`

2. Autolinking a link to a GitHub PR to a link with format "pr-repo-id". Use:
  - Pattern: `https://github\\.com/mattermost/(?P<repo>.+)/pull/(?P<id>\\d+)`
  - Template: `[pr-${repo}-${id}](https://github.com/mattermost/${repo}/pull/${id})`

3. Using autolinking to create group mentions. Use:
  - Pattern: `@customgroup*`
  - Template: `[@customgroup]( \\* @user1 @user2 @user3 \\* )`
