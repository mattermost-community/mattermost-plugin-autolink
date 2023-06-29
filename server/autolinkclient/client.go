package autolinkclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
)

const (
	autolinkPluginID       = "mattermost-autolink"
	AutolinkNameQueryParam = "autolinkName"
)

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

		resp, err := c.call("/"+autolinkPluginID+"/api/v1/link", http.MethodPost, linkBytes, nil)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("unable to add the link %s. Error: %v, %v", link.Name, resp.StatusCode, string(respBody))
		}
	}

	return nil
}

func (c *Client) Delete(links ...string) error {
	for _, link := range links {
		queryParams := url.Values{
			AutolinkNameQueryParam: {link},
		}

		resp, err := c.call("/"+autolinkPluginID+"/api/v1/link", http.MethodDelete, nil, queryParams)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := ioutil.ReadAll(resp.Body)
			return fmt.Errorf("unable to delete the link %s. Error: %v, %v", link, resp.StatusCode, string(respBody))
		}
	}

	return nil
}

func (c *Client) Get(autolinkName string) ([]*autolink.Autolink, error) {
	queryParams := url.Values{
		AutolinkNameQueryParam: {autolinkName},
	}

	resp, err := c.call("/"+autolinkPluginID+"/api/v1/link", http.MethodGet, nil, queryParams)
	if err != nil {
		return nil, err
	}

	var respBody []byte
	if resp.StatusCode != http.StatusOK {
		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unable to get the link %s. Error: %v, %v", autolinkName, resp.StatusCode, string(respBody))
	}

	var response []*autolink.Autolink
	if err = json.Unmarshal(respBody, &response); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) call(url, method string, body []byte, queryParams url.Values) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = queryParams.Encode()

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
