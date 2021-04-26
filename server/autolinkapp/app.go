package autolinkapp

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
)

type Store interface {
	GetLinks() []autolink.Autolink
	SaveLinks([]autolink.Autolink) error
}

type Service interface {
	GetPluginURL() string
	IsAuthorizedAdmin(userID string) (bool, error)
}

type app struct {
	store   Store
	service Service
}

const (
	appID          = "com.mattermost.autolink"
	appDisplayName = "Autolink"

	autolinkCommand = "autolink-app"

	appRoute      = "/app/v1"
	manifestRoute = "/manifest"
	bindingsRoute = "/bindings"
	iconRoute     = "/public/icon.png" // Served from the plugin's public folder

	editLinksRoute        = "/edit-links"
	editLinksCommandRoute = "/edit-links-command"
)

func RegisterHandler(root *mux.Router, store Store, service Service) {
	a := &app{
		store:   store,
		service: service,
	}

	appRouter := root.PathPrefix(appRoute).Subrouter()
	appRouter.HandleFunc(manifestRoute, a.handleManifest).Methods(http.MethodGet)

	adminRoutes := appRouter.PathPrefix("").Subrouter()
	adminRoutes.Use(a.adminRequired)

	adminRoutes.HandleFunc(bindingsRoute, a.handleBindings).Methods(http.MethodPost)
	adminRoutes.HandleFunc(editLinksRoute+"/form", a.handleFormFetch).Methods(http.MethodPost)
	adminRoutes.HandleFunc(editLinksRoute+"/submit", a.handleFormSubmit).Methods(http.MethodPost)
	adminRoutes.HandleFunc(editLinksCommandRoute+"/submit", a.handleCommandSubmit).Methods(http.MethodPost)
}

func (a *app) handleManifest(w http.ResponseWriter, r *http.Request) {
	pluginURL := a.service.GetPluginURL()
	rootURL := pluginURL + appRoute

	manifest := apps.Manifest{
		AppID:       appID,
		DisplayName: appDisplayName,
		AppType:     apps.AppTypeHTTP,
		HTTPRootURL: rootURL,
		RequestedPermissions: apps.Permissions{
			apps.PermissionActAsBot,
		},
		RequestedLocations: apps.Locations{
			apps.LocationCommand,
		},
	}

	httputils.WriteJSON(w, manifest)
}

func (a *app) handleBindings(w http.ResponseWriter, r *http.Request) {
	call := &apps.CallRequest{}
	err := json.NewDecoder(r.Body).Decode(call)
	if err != nil {
		httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.Wrap(err, "failed to decode call request body")))
		return
	}

	icon := a.service.GetPluginURL() + iconRoute
	resp := &apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: []apps.Binding{
			{
				Location: apps.LocationCommand,
				Bindings: []*apps.Binding{
					{
						Icon:        icon,
						Label:       autolinkCommand,
						Location:    autolinkCommand,
						Description: appDisplayName,
						Hint:        "[edit]",
						Bindings: []*apps.Binding{
							{
								Location: "edit",
								Label:    "edit",
								Form:     &apps.Form{Fields: []*apps.Field{}},
								Call: &apps.Call{
									Path: editLinksCommandRoute,
								},
							},
						},
					},
				},
			},
		},
	}

	httputils.WriteJSON(w, resp)
}

func (a *app) handleCommandSubmit(w http.ResponseWriter, r *http.Request) {
	resp := a.getEmptyFormResponse()
	httputils.WriteJSON(w, resp)
}

func (a *app) handleIcon(w http.ResponseWriter, r *http.Request) {
	resp := a.getEmptyFormResponse()
	httputils.WriteJSON(w, resp)
}

func (a *app) adminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.Wrap(err, "failed to read call request body")))
			return
		}

		r.Body.Close()
		r.Body = ioutil.NopCloser(bytes.NewBuffer(b))

		call := &apps.CallRequest{}
		err = json.Unmarshal(b, call)
		if err != nil {
			httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.Wrap(err, "failed to decode call request body")))
			return
		}

		userID := call.Context.ActingUserID
		if userID == "" {
			httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.New("no acting user id provided")))
			return
		}

		authorized, err := a.service.IsAuthorizedAdmin(userID)
		if err != nil {
			httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.Wrap(err, "not authorized")))
			return
		}

		if !authorized {
			httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.New("not authorized")))
			return
		}

		next.ServeHTTP(w, r)
	})
}
