# Autolink Plugin

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-autolink/master.svg)](https://circleci.com/gh/mattermost/mattermost-plugin-autolink)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-autolink/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-autolink)
[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-autolink)](https://github.com/mattermost/mattermost-plugin-autolink/releases/latest)
[![HW](https://img.shields.io/github/issues/mattermost/mattermost-plugin-autolink/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/mattermost/mattermost-plugin-autolink/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

**Maintainer:** [@levb](https://github.com/levb)
**Co-Maintainer:** [@iomodo](https://github.com/iomodo)

This plugin creates regular expression (regexp) patterns that are reformatted into a Markdown link before the message is saved into the database.

Use it to add custom auto-linking on your Mattermost system, such as adding links to your issue tracker based on the regexp patterns.

![image](https://user-images.githubusercontent.com/13119842/58675221-479ad680-8321-11e9-9ad1-b9c42238734f.png)

*Posting a message containing a Jira issue key..*

![image](https://user-images.githubusercontent.com/13119842/58675165-11f5ed80-8321-11e9-9d41-91088a79a11b.png)

*..automatically links to the corresponding issue in the Jira project*

## Configuration
1. Go to **System Console > Plugins > Management** and click **Enable** to enable the Autolink plugin.
    - If you are running Mattermost v5.11 or earlier, you must first go to the [releases page of this GitHub repository](https://github.com/mattermost/mattermost-plugin-autolink), download the latest release, and upload it to your Mattermost instance [following this documentation](https://docs.mattermost.com/administration/plugins.html#plugin-uploads).

2. Modify your `config.json` file to include the types of regexp patterns you wish to match, under the `PluginSettings`. See below for an example of what this should look like.

**Tip**: There are useful Regular Expression tools online to help test and validate that your formulas are working as expected.  One such tool is [Regex101](https://regex101.com/) . Here is an example Regular Expression to capture a post that includes a [VISA card number](https://regex101.com/r/JGKCTN/1) - which you could then obfuscate with the `Pattern` so people don't accidentally share sensitive info in your channels.

## Usage

Autolinks have 3 parts: a **Pattern** which is a regular expression search pattern utilizing the [Golang regexp library](https://golang.org/pkg/regexp/), a **Template** that gets expanded and an optional **Scope** parameter  to define which team/channel the autolink applies to. You can create variables in the pattern with the syntax `(?P<name>...)` which will then be expanded by the corresponding template.

In the template, a variable is denoted by a substring of the form `$name` or `${name}`, where `name` is a non-empty sequence of letters, digits, and underscores. A purely numeric name like $1 refers to the submatch with the corresponding index. In the $name form, name is taken to be as long as possible: $1x is equivalent to ${1x}, not ${1}x, and, $10 is equivalent to ${10}, not ${1}0. To insert a literal $ in the output, use $$ in the template.

Below is an example of regexp patterns used for autolinking at https://community.mattermost.com, modified in the `config.json` file:

```json5
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
  - Scope: `["teams/committers"]` (optional)

2. Autolinking a link to a GitHub PR to a format "pr-repo-id". Use:
  - Pattern: `https://github\\.com/mattermost/(?P<repo>.+)/pull/(?P<id>\\d+)`
  - Template: `[pr-${repo}-${id}](https://github.com/mattermost/${repo}/pull/${id})`

3. Using autolinking to create group mentions. Use (note that clicking the resulting at-mention redirects to a broken page):
  - Pattern: `@customgroup*`
  - Template: `[@customgroup]( \\* @user1 @user2 @user3 \\* )`
  
4. For servers with multiple domains (like community and community-daily on the [public Mattermost Server](https://community.mattermost)), a substitution of absolute conversation links to relative links is recommended to prevent issues in the mobile app. Add one pattern for each domain used:
  - Pattern: `https://community\\.mattermost\\.com/(?P\u003cteamname\u003e(?a-zA-Z0-9]+)/(?P\u003cid\u003e[a-zA-Z0-9]+)`
  - Template: `[<jump to convo>](/${teamname}/pl/${id})/${id})`


**You can check your pattern with those Regex Testers:**
- https://regex-golang.appspot.com/,
- https://regex101.com/,
- https://www.regextester.com/.

## Configuration Management
The /autolink commands allow the users to easily edit the configurations.

 Commands | Description | Usage
 ---|---|---|
 list | Lists all configured links |
 list \<*linkref*> | List a specific link which matched the link reference |
 test \<*linkref*> test-text | Test a link on the text provided | /autolink test Visa 4356-7891-2345-1111 -- (4111222233334444)
 enable \<*linkref*> | Enables the link | /autolink enable Visa
 disable \<*linkref*> | Disable the link |/autolink disable Visa
 add \<*name*> | Creates a new link with the name specified in the command  | /autolink add Visa
 delete \<*linkref*> |  Delete the link | /autolink delete Visa
 set \<*linkref*> \<*field*> *value* | Sets a link's field to a value <br> *Fields* - <br> <ul><li>Template - Sets the Template field</li><li>Pattern - Sets the Pattern field </li> <li> WordMatch - If true uses the [\b word boundaries](https://www.regular-expressions.info/wordboundaries.html) </li> | <br> /autolink set Visa Pattern (?P<VISA>(?P<part1>4\d{3})[ -]?(?P<part2>\d{4})[ -]?(?P<part3>\d{4})[ -]?(?P<LastFour>[0-9]{4})) <br><br> /autolink set Visa Template VISA XXXX-XXXX-XXXX-$LastFour <br><br> /autolink set Visa WordMatch true <br><br>


## Development

This plugin contains a server portion.

Use `make dist` to build distributions of the plugin that you can upload to a Mattermost server.
Use `make check-style` to check the style.
Use `make deploy` to deploy the plugin to your local server.

For additional information on developing plugins, refer to [our plugin developer documentation](https://developers.mattermost.com/extend/plugins/).
