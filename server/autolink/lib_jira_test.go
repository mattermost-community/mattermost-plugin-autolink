package autolink_test

import (
	"testing"

	"github.com/mattermost-community/mattermost-plugin-autolink/server/autolink"
)

var jiraTests = []linkTest{
	{
		"Jira example",
		autolink.Autolink{
			Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
			Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
		},
		"Welcome MM-12345 should link!",
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
	}, {
		"Jira example 2 (within a ())",
		autolink.Autolink{
			Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
			Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
		},
		"Link in brackets should link (see MM-12345)",
		"Link in brackets should link (see [MM-12345](https://mattermost.atlassian.net/browse/MM-12345))",
	}, {
		"Jira example 3 (before ,)",
		autolink.Autolink{
			Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
			Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
		},
		"Link a ticket MM-12345, before a comma",
		"Link a ticket [MM-12345](https://mattermost.atlassian.net/browse/MM-12345), before a comma",
	}, {
		"Jira example 3 (at begin of the message)",
		autolink.Autolink{
			Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
			Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
		},
		"MM-12345 should link!",
		"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
	}, {
		"Pattern word prefix and suffix disabled",
		autolink.Autolink{
			Pattern:              "(?P<previous>^|\\s)(MM)(-)(?P<jira_id>\\d+)",
			Template:             "${previous}[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			DisableNonWordPrefix: true,
			DisableNonWordSuffix: true,
		},
		"Welcome MM-12345 should link!",
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
	}, {
		"Pattern word prefix and suffix disabled (at begin of the message)",
		autolink.Autolink{
			Pattern:              "(?P<previous>^|\\s)(MM)(-)(?P<jira_id>\\d+)",
			Template:             "${previous}[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			DisableNonWordPrefix: true,
			DisableNonWordSuffix: true,
		},
		"MM-12345 should link!",
		"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
	}, {
		"Pattern word prefix and suffix enable (in the middle of other text)",
		autolink.Autolink{
			Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
			Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
		},
		"WelcomeMM-12345should not link!",
		"WelcomeMM-12345should not link!",
	}, {
		"Pattern word prefix and suffix disabled (in the middle of other text)",
		autolink.Autolink{
			Pattern:              "(MM)(-)(?P<jira_id>\\d+)",
			Template:             "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
			DisableNonWordPrefix: true,
			DisableNonWordSuffix: true,
		},
		"WelcomeMM-12345should link!",
		"Welcome[MM-12345](https://mattermost.atlassian.net/browse/MM-12345)should link!",
	}, {
		"Not relinking",
		autolink.Autolink{
			Pattern:  "(MM)(-)(?P<jira_id>\\d+)",
			Template: "[MM-${jira_id}](https://mattermost.atlassian.net/browse/MM-${jira_id})",
		},
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should not re-link!",
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should not re-link!",
	}, {
		"Url replacement",
		autolink.Autolink{
			Pattern:  "(https://mattermost.atlassian.net/browse/)(?P<project_id>\\w+)(-)(?P<jira_id>\\d+)",
			Template: "[${project_id}-${jira_id}](https://mattermost.atlassian.net/browse/${project_id}-${jira_id})",
		},
		"Welcome https://mattermost.atlassian.net/browse/MM-12345 should link!",
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
	}, {
		"Url replacement multiple times",
		autolink.Autolink{
			Pattern:  "(https://mattermost.atlassian.net/browse/)(?P<project_id>\\w+)(-)(?P<jira_id>\\d+)",
			Template: "[${project_id}-${jira_id}](https://mattermost.atlassian.net/browse/${project_id}-${jira_id})",
		},
		"Welcome https://mattermost.atlassian.net/browse/MM-12345. should link https://mattermost.atlassian.net/browse/MM-12346 !",
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345). should link [MM-12346](https://mattermost.atlassian.net/browse/MM-12346) !",
	}, {
		"Url replacement multiple times and at beginning",
		autolink.Autolink{
			Pattern:  "(https:\\/\\/mattermost.atlassian.net\\/browse\\/)(?P<project_id>\\w+)(-)(?P<jira_id>\\d+)",
			Template: "[${project_id}-${jira_id}](https://mattermost.atlassian.net/browse/${project_id}-${jira_id})",
		},
		"https://mattermost.atlassian.net/browse/MM-12345 https://mattermost.atlassian.net/browse/MM-12345",
		"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
	}, {
		"Url replacement at end",
		autolink.Autolink{
			Pattern:  "(https://mattermost.atlassian.net/browse/)(?P<project_id>\\w+)(-)(?P<jira_id>\\d+)",
			Template: "[${project_id}-${jira_id}](https://mattermost.atlassian.net/browse/${project_id}-${jira_id})",
		},
		"Welcome https://mattermost.atlassian.net/browse/MM-12345",
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
	}, {
		"Comment url replacement",
		autolink.Autolink{
			Pattern:  "(https://mattermost.atlassian.net/browse/)(?P<project_id>\\w+)(-)(?P<jira_id>\\d+)[?](focusedCommentId)(=)(?P<comment_id>\\d+)",
			Template: "[${project_id}-${jira_id} (comment)](https://mattermost.atlassian.net/browse/${project_id}-${jira_id}?focusedCommentId=${comment_id})",
		},
		"Welcome https://mattermost.atlassian.net/browse/MM-12345?focusedCommentId=10210 should link!",
		"Welcome [MM-12345 (comment)](https://mattermost.atlassian.net/browse/MM-12345?focusedCommentId=10210) should link!",
	}, {
		"Comment url replacement multiple times",
		autolink.Autolink{
			Pattern:  "(https://mattermost.atlassian.net/browse/)(?P<project_id>\\w+)(-)(?P<jira_id>\\d+)[?](focusedCommentId)(=)(?P<comment_id>\\d+)",
			Template: "[${project_id}-${jira_id} With Comment #${comment_id}](https://mattermost.atlassian.net/browse/${project_id}-${jira_id}?focusedCommentId=${comment_id})",
		},
		"Welcome https://mattermost.atlassian.net/browse/MM-12345?focusedCommentId=10210. should link https://mattermost.atlassian.net/browse/MM-12346?focusedCommentId=10210!",
		"Welcome [MM-12345 With Comment #10210](https://mattermost.atlassian.net/browse/MM-12345?focusedCommentId=10210). should link [MM-12346 With Comment #10210](https://mattermost.atlassian.net/browse/MM-12346?focusedCommentId=10210)!",
	}, {
		"Comment url replacement multiple times and at beginning",
		autolink.Autolink{
			Pattern:  "(https:\\/\\/mattermost.atlassian.net\\/browse\\/)(?P<project_id>\\w+)(-)(?P<jira_id>\\d+)[?](focusedCommentId)(=)(?P<comment_id>\\d+)",
			Template: "[${project_id}-${jira_id} With Comment #${comment_id}](https://mattermost.atlassian.net/browse/${project_id}-${jira_id}?focusedCommentId=${comment_id})",
		},
		"https://mattermost.atlassian.net/browse/MM-12345?focusedCommentId=10210 https://mattermost.atlassian.net/browse/MM-12345?focusedCommentId=10210",
		"[MM-12345 With Comment #10210](https://mattermost.atlassian.net/browse/MM-12345?focusedCommentId=10210) [MM-12345 With Comment #10210](https://mattermost.atlassian.net/browse/MM-12345?focusedCommentId=10210)",
	}, {
		"Comment url replacement at end",
		autolink.Autolink{
			Pattern:  "(https://mattermost.atlassian.net/browse/)(?P<project_id>\\w+)(-)(?P<jira_id>\\d+)[?](focusedCommentId)(=)(?P<comment_id>\\d+)",
			Template: "[${project_id}-${jira_id} With Comment #${comment_id}](https://mattermost.atlassian.net/browse/${project_id}-${jira_id}?focusedCommentId=${comment_id})",
		},
		"Welcome https://mattermost.atlassian.net/browse/MM-12345?focusedCommentId=10210",
		"Welcome [MM-12345 With Comment #10210](https://mattermost.atlassian.net/browse/MM-12345?focusedCommentId=10210)",
	},
}

func TestJira(t *testing.T) {
	testLinks(t, jiraTests...)
}
