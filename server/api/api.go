package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
	"github.com/mattermost/mattermost-plugin-autolink/server/autolinkapp"
	"github.com/mattermost/mattermost-plugin-autolink/server/store"
)

type Authorization interface {
	IsAuthorizedAdmin(userID string) (bool, error)
}

type Handler struct {
	root          *mux.Router
	store         store.Store
	authorization Authorization
}

func NewHandler(store store.Store, authorization Authorization, service autolinkapp.Service) *Handler {
	h := &Handler{
		store:         store,
		authorization: authorization,
	}

	root := mux.NewRouter()
	api := root.PathPrefix("/api/v1").Subrouter()
	api.Use(h.adminOrPluginRequired)
	api.HandleFunc("/link", h.setLink).Methods("POST")

	api.Handle("{anything:.*}", http.NotFoundHandler())

	appRouter := root.PathPrefix("/app/v1").Subrouter()
	// appRouter.Use(h.adminOrPluginRequired) // will fetching the manifest fail?
	autolinkapp.RegisterRouter(appRouter, store, service)

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
		h.handleError(w, fmt.Errorf("unable to decode body: %w", err))
		return
	}

	changed, err := store.SetLink(h.store, newLink)
	if err != nil {
		h.handleError(w, fmt.Errorf("unable to save link: %w", err))
		return
	}

	status := http.StatusNotModified
	if changed {
		status = http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`{"status": "OK"}`))
}
