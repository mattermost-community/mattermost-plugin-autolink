package autolink_test

import (
	"testing"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
)

var productboardLink = autolink.Autolink{
	Pattern:   `(?P<url>https://mattermost\.productboard\.com/.+)`,
	Template:  "[ProductBoard link]($url)",
	WordMatch: true,
}

var productboardTests = []linkTest{
	{
		"Url replacement",
		productboardLink,
		"Welcome to https://mattermost.productboard.com/somepage should link!",
		"Welcome to [ProductBoard link](https://mattermost.productboard.com/somepage) should link!",
	}, {
		"Not relinking",
		productboardLink,
		"Welcome to [other link](https://mattermost.productboard.com/somepage) should not re-link!",
		"Welcome to [other link](https://mattermost.productboard.com/somepage) should not re-link!",
	}, {
		"Word boundary happy",
		productboardLink,
		"Welcome to (https://mattermost.productboard.com/somepage) should link!",
		"Welcome to ([ProductBoard link](https://mattermost.productboard.com/somepage)) should link!",
	}, {
		"Word boundary un-happy",
		productboardLink,
		"Welcome to (BADhttps://mattermost.productboard.com/somepage) should not link!",
		"Welcome to (BADhttps://mattermost.productboard.com/somepage) should not link!",
	},
}

func TestProductBoard(t *testing.T) {
	testLinks(t, productboardTests...)
}
