package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
	"github.com/mattermost/mattermost-plugin-autolink/server/autolinkclient"
)

type Store interface {
	GetLinks() []autolink.Autolink
	SaveLinks([]autolink.Autolink) error
}

type Authorization interface {
	IsAuthorizedAdmin(userID string) (bool, error)
}

type Handler struct {
	root          *mux.Router
	store         Store
	authorization Authorization
}

func NewHandler(store Store, authorization Authorization) *Handler {
	h := &Handler{
		store:         store,
		authorization: authorization,
	}

	root := mux.NewRouter()
	api := root.PathPrefix("/api/v1").Subrouter()
	api.Use(h.adminOrPluginRequired)
	link := api.PathPrefix("/link").Subrouter()
	link.HandleFunc("", h.setLink).Methods(http.MethodPost)
	link.HandleFunc("", h.deleteLink).Methods(http.MethodDelete)
	link.HandleFunc("", h.getLinks).Methods(http.MethodGet)

	api.Handle("{anything:.*}", http.NotFoundHandler())

	h.root = root

	return h
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	b, _ := json.Marshal(struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}{
		Error:   "An internal error has occurred. Check app server logs for details.",
		Details: err.Error(),
	})
	_, _ = w.Write(b)
}

func (h *Handler) adminOrPluginRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		authorized := false
		pluginID := r.Header.Get("Mattermost-Plugin-ID")
		if pluginID != "" {
			// All other plugins are allowed
			authorized = true
		}

		userID := r.Header.Get("Mattermost-User-ID")
		if !authorized && userID != "" {
			authorized, err = h.authorization.IsAuthorizedAdmin(userID)
			if err != nil {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}
		}

		if !authorized {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.root.ServeHTTP(w, r)
}

func (h *Handler) setLink(w http.ResponseWriter, r *http.Request) {
	var newLink autolink.Autolink
	if err := json.NewDecoder(r.Body).Decode(&newLink); err != nil {
		h.handleError(w, errors.Wrap(err, "unable to decode body"))
		return
	}

	links := h.store.GetLinks()
	found := false
	changed := false
	for i := range links {
		if links[i].Name == newLink.Name {
			if !links[i].Equals(newLink) {
				links[i] = newLink
				changed = true
			}
			found = true
			break
		}
	}
	if !found {
		links = append(h.store.GetLinks(), newLink)
		changed = true
	}
	status := http.StatusNotModified
	if changed {
		if err := h.store.SaveLinks(links); err != nil {
			h.handleError(w, errors.Wrap(err, "unable to save the link"))
			return
		}
		status = http.StatusOK
	}

	ReturnStatusOK(status, w)
}

func (h *Handler) deleteLink(w http.ResponseWriter, r *http.Request) {
	autolinkName := r.URL.Query().Get(autolinkclient.AutolinkNameQueryParam)
	if autolinkName == "" {
		h.handleError(w, errors.New("autolink name should not be empty"))
		return
	}

	links := h.store.GetLinks()
	found := false
	for i := 0; i < len(links); i++ {
		if links[i].Name == autolinkName {
			links = append(links[:i], links[i+1:]...)
			found = true
			break
		}
	}

	status := http.StatusNotFound
	if found {
		if err := h.store.SaveLinks(links); err != nil {
			h.handleError(w, errors.Wrap(err, "unable to save the link"))
			return
		}
		status = http.StatusOK
	}

	ReturnStatusOK(status, w)
}

func (h *Handler) getLinks(w http.ResponseWriter, r *http.Request) {
	links := h.store.GetLinks()

	autolinkName := r.URL.Query().Get(autolinkclient.AutolinkNameQueryParam)
	if autolinkName == "" {
		h.handleSendingJSONContent(w, links)
		return
	}

	var autolink *autolink.Autolink
	for _, link := range links {
		if link.Name == autolinkName {
			autolink = &link
			break
		}
	}
	if autolink == nil {
		h.handleError(w, errors.Errorf("no autolink found with name %s", autolinkName))
		return
	}

	h.handleSendingJSONContent(w, autolink)
	return
}

func (h *Handler) handleSendingJSONContent(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(v)
	if err != nil {
		h.handleError(w, errors.Wrap(err, "failed to marshal JSON response"))
		return
	}

	if _, err = w.Write(b); err != nil {
		h.handleError(w, errors.Wrap(err, "failed to write JSON response"))
		return
	}
}

func ReturnStatusOK(status int, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"status": "OK"}`))
}
