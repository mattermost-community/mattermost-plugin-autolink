package autolinkclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
)

const autolinkPluginID = "mattermost-autolink"

type PluginAPI interface {
	PluginHTTP(*http.Request) *http.Response
}

type Client struct {
	http.Client
}

type pluginAPIRoundTripper struct {
	api PluginAPI
}

func (p *pluginAPIRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := p.api.PluginHTTP(req)
	if resp == nil {
		return nil, fmt.Errorf("failed to make interplugin request")
	}
	return resp, nil
}

func NewClientPlugin(api PluginAPI) *Client {
	client := &Client{}
	client.Transport = &pluginAPIRoundTripper{api}
	return client
}

func (c *Client) Add(links ...autolink.Autolink) error {
	for _, link := range links {
		linkBytes, err := json.Marshal(link)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("POST", "/"+autolinkPluginID+"/api/v1/link", bytes.NewReader(linkBytes))
		if err != nil {
			return err
		}

		resp, err := c.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("unable to install autolink. Error: %v, %v", resp.StatusCode, string(respBody))
		}
	}

	return nil
}
