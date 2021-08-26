package autolinkapp

import (
	"bytes"
	_ "embed" // needed for app icon
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/apps/mmclient"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
)

type Store interface {
	GetLinks() []autolink.Autolink
	SaveLinks([]autolink.Autolink) error
}

type Service interface {
	IsAuthorizedAdmin(userID string) (bool, error)
	DisableApp(appID apps.AppID, sessionID, actingUserID string) error
}

type app struct {
	store   Store
	service Service
}

//go:embed icon.png
var iconData []byte

const (
	appID          = "com.mattermost.autolink"
	appDisplayName = "Autolink"
	appHomepageURL = "https://github.com/mattermost/mattermost-plugin-autolink"

	autolinkCommand = "autolink"

	appIcon       = "icon.png"
	iconRoute     = "/static/" + appIcon
	bindingsRoute = "/bindings"

	editLinksRoute        = "/edit-links"
	editLinksCommandRoute = "/edit-links-command"
	appEnableRoute        = "/app/enable"
	appDisableRoute       = "/app/disable"
)

var Manifest = apps.Manifest{
	AppID:       appID,
	DisplayName: appDisplayName,
	HomepageURL: appHomepageURL,
	AppType:     apps.AppTypePlugin,
	Icon:        appIcon,
	RequestedPermissions: apps.Permissions{
		apps.PermissionActAsBot,
	},
	RequestedLocations: apps.Locations{
		apps.LocationCommand,
	},
	PluginID: "mattermost-autolink",
}

func RegisterHandler(root *mux.Router, store Store, service Service) {
	a := &app{
		store:   store,
		service: service,
	}

	appRouter := root.PathPrefix(apps.PluginAppPath).Subrouter()
	appRouter.Use(a.checkAppsPlugin)
	appRouter.HandleFunc(iconRoute, a.handleIcon).Methods(http.MethodGet)

	adminRoutes := appRouter.NewRoute().Subrouter()
	adminRoutes.Use(a.adminRequired)

	adminRoutes.HandleFunc(bindingsRoute, a.handleBindings).Methods(http.MethodPost)
	adminRoutes.HandleFunc(editLinksRoute+"/form", a.handleFormFetch).Methods(http.MethodPost)
	adminRoutes.HandleFunc(editLinksRoute+"/submit", a.handleFormSubmit).Methods(http.MethodPost)
	adminRoutes.HandleFunc(editLinksCommandRoute+"/submit", a.handleCommandSubmit).Methods(http.MethodPost)

	adminRoutes.HandleFunc(appEnableRoute+"/"+string(apps.CallTypeSubmit), a.handleAppEnable).Methods(http.MethodPost)
	adminRoutes.HandleFunc(appDisableRoute+"/"+string(apps.CallTypeSubmit), a.handleAppDisable).Methods(http.MethodPost)
}

func (a *app) handleBindings(w http.ResponseWriter, r *http.Request) {
	call := &apps.CallRequest{}
	err := json.NewDecoder(r.Body).Decode(call)
	if err != nil {
		httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.Wrap(err, "failed to decode call request body")))
		return
	}

	resp := &apps.CallResponse{
		Type: apps.CallResponseTypeOK,
		Data: []apps.Binding{
			{
				Location: apps.LocationCommand,
				Bindings: []*apps.Binding{
					{
						Icon:        appIcon,
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
							}, {
								Location: "experimental",
								Label:    "experimental",
								Bindings: []*apps.Binding{
									{
										Location: "app",
										Label:    "app",
										Bindings: []*apps.Binding{
											{
												Location: "on",
												Label:    "on",
												Form:     &apps.Form{},
												Call: &apps.Call{
													Path: appEnableRoute,
												},
											}, {
												Location: "off",
												Label:    "off",
												Form:     &apps.Form{},
												Call: &apps.Call{
													Path: appDisableRoute,
												},
											},
										},
									},
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

func (a *app) handleAppEnable(w http.ResponseWriter, r *http.Request) {
	resp := apps.CallResponse{
		Markdown: "App is already enabled",
	}
	httputils.WriteJSON(w, resp)
}

func (a *app) handleAppDisable(w http.ResponseWriter, r *http.Request) {
	call := &apps.CallRequest{}
	err := json.NewDecoder(r.Body).Decode(call)
	if err != nil {
		httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.Wrap(err, "failed to decode call request body")))
		return
	}

	log.Printf("call.Context.UserID: %#+v\n", call.Context.UserID)
	log.Printf("call.Context.ActingUserAccessToken: %#+v\n", call.Context.ActingUserAccessToken)

	err = a.service.DisableApp(Manifest.AppID, call.Context.UserID, call.Context.ActingUserAccessToken)
	if err != nil {
		httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.Wrap(err, "failed to disable app")))
		return
	}

	resp := apps.CallResponse{
		Markdown: "Successfully disabled autolink app",
	}
	httputils.WriteJSON(w, resp)
}

func (a *app) handleIcon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	_, _ = w.Write(iconData)
}

func (a *app) checkAppsPlugin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pluginID := r.Header.Get("Mattermost-Plugin-ID")
		if pluginID != mmclient.AppsPluginName {
			httputils.WriteError(w, errors.New("not authorized"))
			return
		}

		next.ServeHTTP(w, r)
	})
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
