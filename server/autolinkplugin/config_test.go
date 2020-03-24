package autolinkplugin

import (
	"errors"
	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest"
	"github.com/mattermost/mattermost-server/v5/plugin/plugintest/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOnConfigurationChange(t *testing.T) {

	t.Run("Invalid Configuration", func(t *testing.T) {
		api := &plugintest.API{}
		api.On("LoadPluginConfiguration",
			mock.AnythingOfType("*autolinkplugin.Config")).Return(func(dest interface{}) error {
			return errors.New("LoadPluginConfiguration Error")
		})

		p := Plugin{}
		p.SetAPI(api)

		err := p.OnConfigurationChange()
		assert.Error(t, err)
	})

	t.Run("Invalid Autolink", func(t *testing.T) {
		conf := Config{
			Links: []autolink.Autolink{
				autolink.Autolink{
					Name:     "existing",
					Pattern:  ")",
					Template: "otherthing",
				},
			},
		}

		api := &plugintest.API{}
		api.On("LoadPluginConfiguration",
			mock.AnythingOfType("*autolinkplugin.Config")).Return(func(dest interface{}) error {
			*dest.(*Config) = conf
			return nil
		})

		api.On("LogError",
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("Autolink"),
			mock.AnythingOfType("string"),
			mock.AnythingOfType("string")).Return(nil)

		api.On("UnregisterCommand", mock.AnythingOfType("string"),
			mock.AnythingOfType("string")).Return((*model.AppError)(nil))

		p := New()
		p.SetAPI(api)
		p.OnConfigurationChange()

		api.AssertNumberOfCalls(t, "LogError", 1)
	})
}
