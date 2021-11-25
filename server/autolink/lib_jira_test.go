package autolink_test

import (
	"testing"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
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
			Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
		},
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should not re-link!",
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should not re-link!",
	}, {
		"Url replacement",
		autolink.Autolink{
			Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
			Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
		},
		"Welcome https://mattermost.atlassian.net/browse/MM-12345 should link!",
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345) should link!",
	}, {
		"Url replacement multiple times",
		autolink.Autolink{
			Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
			Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
		},
		"Welcome https://mattermost.atlassian.net/browse/MM-12345. should link https://mattermost.atlassian.net/browse/MM-12346 !",
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345). should link [MM-12346](https://mattermost.atlassian.net/browse/MM-12346) !",
	}, {
		"Url replacement multiple times and at beginning",
		autolink.Autolink{
			Pattern:  "(https:\\/\\/mattermost.atlassian.net\\/browse\\/)(MM)(-)(?P<jira_id>\\d+)",
			Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
		},
		"https://mattermost.atlassian.net/browse/MM-12345 https://mattermost.atlassian.net/browse/MM-12345",
		"[MM-12345](https://mattermost.atlassian.net/browse/MM-12345) [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
	}, {
		"Url replacement at end",
		autolink.Autolink{
			Pattern:  "(https://mattermost.atlassian.net/browse/)(MM)(-)(?P<jira_id>\\d+)",
			Template: "[MM-$jira_id](https://mattermost.atlassian.net/browse/MM-$jira_id)",
		},
		"Welcome https://mattermost.atlassian.net/browse/MM-12345",
		"Welcome [MM-12345](https://mattermost.atlassian.net/browse/MM-12345)",
	},
	{
		"Jump To Comment With Jira Cloud",
		autolink.Autolink{
			Pattern:  "(https://mmtest.atlassian.net/browse/)(DP)(-)(?P<jira_id>\\d+)[?](focusedCommentId)(=)(?P<comment_id>\\d+)",
			Template: "[DP-${jira_id} With Focused Comment($comment_id)](https://mmtest.atlassian.net/browse/DP-${jira_id}?focusedCommentId=$comment_id)",
		},
		"https://mmtest.atlassian.net/browse/DP-454?focusedCommentId=11347",
		"[DP-454 With Focused Comment(11347)](https://mmtest.atlassian.net/browse/DP-454?focusedCommentId=11347)",
	},
	{
		"Jump To Comment With Jira Ecc",
		autolink.Autolink{
			Pattern:  "(http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/)(DKHPROJ)(-)(?P<jira_id>\\d+)[?](focusedCommentId)(=)(?P<comment_id>\\d+)",
			Template: "[DKHPROJ-${jira_id} With Focused Comment($comment_id)](http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-${jira_id}?focusedCommentId=$comment_id)",
		},
		"http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-5?focusedCommentId=10200",
		"[DKHPROJ-5 With Focused Comment(10200)](http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-5?focusedCommentId=10200)",
	},
	{
		"Jump To Comment With Jira Ecc Long Link",
		autolink.Autolink{
			Pattern:  "(http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/)(DKHPROJ)(-)(?P<jira_id>\\d+)[?](focusedCommentId)(=)(?P<comment_id>\\d+)",
			Template: "[DKHPROJ-${jira_id} With Focused Comment($comment_id)](http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-${jira_id}?focusedCommentId=$comment_id)",
		},
		"http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-5?focusedCommentId=10200&page=com.atlassian.jira.plugin.system.issuetabpanels%3Acomment-tabpanel#comment-10200",
		"http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-5?focusedCommentId=10200&page=com.atlassian.jira.plugin.system.issuetabpanels%3Acomment-tabpanel#comment-10200",
	},
	{
		"Jump To Comment With Jira Ecc Long Short",
		autolink.Autolink{
			Pattern:  "(http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/)(DKHPROJ)(-)(?P<jira_id>\\d+)[?](focusedCommentId)(=)(?P<comment_id>\\d+)",
			Template: "[DKHPROJ-${jira_id} With Focused Comment($comment_id)](http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-${jira_id}?focusedCommentId=$comment_id)",
		},
		"http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-5",
		"http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-5",
	},
	{
		"Jump To Comment With Jira Cloud Failed",
		autolink.Autolink{
			Pattern:  "(https://mmtest.atlassian.net/browse/)(DP)(-)(?P<jira_id>\\d+)[?](focusedCommentId)(=)(?P<comment_id>\\d+)",
			Template: "[DP-${jira_id} With Focused Comment($comment_id)](https://mmtest.atlassian.net/browse/DP-${jira_id}",
		},
		"https://mmtest.atlassian.net/browse/DP-454",
		"https://mmtest.atlassian.net/browse/DP-454",
	},
	// Trials Linker
	{
		"Trial With Jira Cloud",
		autolink.Autolink{
			Pattern:  "(?P<URI>.*/)(?P<IssueKey>[A-Za-z]+-[0-9]+)[?](focusedCommentId=*(?P<FocusedCommentId>[^\\s&]+))",
			Template: "[${IssueKey} FocusedComment(${FocusedCommentId})](${URI}${IssueKey}?focusedCommentId=${FocusedCommentId})",
		},
		"https://mmtest.atlassian.net/browse/DP-454?focusedCommentId=11347",
		"[DP-454 FocusedComment(11347)](https://mmtest.atlassian.net/browse/DP-454?focusedCommentId=11347)",
	},
	{
		"Trial With Jira Server #1",
		autolink.Autolink{
			Pattern:  "(?P<URI>.*/)(?P<IssueKey>[A-Za-z]+-[0-9]+)[?](focusedCommentId=*(?P<FocusedCommentId>[^\\s&]+))",
			Template: "[${IssueKey} FocusedComment(${FocusedCommentId})](${URI}${IssueKey}?focusedCommentId=${FocusedCommentId})",
		},
		"http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-5?focusedCommentId=10200",
		"[DKHPROJ-5 FocusedComment(10200)](http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-5?focusedCommentId=10200)",
	},
	{
		"Trial With Jira Server #2", //Not Worked here ?
		autolink.Autolink{
			Pattern:  "(?P<URI>.*/)(?P<IssueKey>[A-Za-z]+-[0-9]+)[?](focusedCommentId=*(?P<FocusedCommentId>[^\\s&]+))",
			Template: "[${IssueKey} FocusedComment(${FocusedCommentId})](${URI}${IssueKey}?focusedCommentId=${FocusedCommentId})",
		},
		"http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-5?focusedCommentId=10200&page=com.atlassian.jira.plugin.system.issuetabpanels%3Acomment-tabpanel#comment-10200",
		"http://ec2-54-157-116-101.compute-1.amazonaws.com/browse/DKHPROJ-5?focusedCommentId=10200&page=com.atlassian.jira.plugin.system.issuetabpanels%3Acomment-tabpanel#comment-10200",
	},
	{
		"Trial With Jira Cloud Without Focused Comment ID", //Not Worked here ?
		autolink.Autolink{
			Pattern:  "(?P<URI>.*/)(?P<IssueKey>[A-Za-z]+-[0-9]+)[?](focusedCommentId=*(?P<FocusedCommentId>[^\\s&]+))",
			Template: "[${IssueKey} FocusedComment(${FocusedCommentId})](${URI}${IssueKey}?focusedCommentId=${FocusedCommentId})",
		},
		"https://mmtest.atlassian.net/browse/DP-454",
		"https://mmtest.atlassian.net/browse/DP-454",
	},
}

func TestJira(t *testing.T) {
	testLinks(t, jiraTests...)
}
