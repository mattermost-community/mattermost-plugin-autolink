package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
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
	api.HandleFunc("/link", h.setLink).Methods("POST")

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
		if links[i].Name == newLink.Name || links[i].Pattern == newLink.Pattern {
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
			h.handleError(w, errors.Wrap(err, "unable to save link"))
			return
		}
		status = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"status": "OK"}`))
}
