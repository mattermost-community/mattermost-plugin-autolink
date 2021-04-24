package autolinkapp

import (
	_ "embed"
	"encoding/json"
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
	GetSiteURL() string
}

type app struct {
	store   Store
	service Service
}

// //go:embed icon.png
// var iconData []byte

//go:embed manifest.json
var manifestData []byte

//go:embed bindings.json
var bindingsData []byte

func RegisterRouter(router *mux.Router, store Store, service Service) {
	a := &app{
		store:   store,
		service: service,
	}

	router.HandleFunc("/manifest", a.handleManifest).Methods("GET")
	router.HandleFunc("/bindings", a.handleBindings).Methods("POST")

	router.HandleFunc("/edit-links/form", a.handleFormFetch)
	router.HandleFunc("/edit-links/submit", a.handleFormSubmit)
	router.HandleFunc("/edit-links-command/submit", a.handleCommandSubmit)
}

func (a *app) handleManifest(w http.ResponseWriter, r *http.Request) {
	manifest := apps.Manifest{}
	err := json.Unmarshal(manifestData, &manifest)
	if err != nil {
		httputils.WriteError(w, errors.Wrap(err, "failed to unmarshal manifest"))
		return
	}

	s := a.service.GetSiteURL()
	manifest.HTTPRootURL = s + "/plugins/mattermost-autolink/app/v1"
	httputils.WriteJSON(w, manifest)
}

func (a *app) handleBindings(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write(bindingsData)
}

func (a *app) handleCommandSubmit(w http.ResponseWriter, r *http.Request) {
	f := a.getEmptyForm()
	resp := apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: f,
	}

	httputils.WriteJSON(w, resp)
}
