package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-plugin-autolink/server/link"
)

const PluginIDContextValue = "pluginid"

type LinkStore interface {
	GetLinks() []link.Link
	SaveLinks([]link.Link) error
}

type Authorization interface {
	IsAuthorizedAdmin(userID string) (bool, error)
}

type Handler struct {
	root          *mux.Router
	linkStore     LinkStore
	authorization Authorization
}

func NewHandler(linkStore LinkStore, authorization Authorization) *Handler {
	h := &Handler{
		linkStore:     linkStore,
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

func (h *Handler) handleErrorWithCode(w http.ResponseWriter, code int, errTitle string, err error) {
	w.WriteHeader(code)
	b, _ := json.Marshal(struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}{
		Error:   errTitle,
		Details: err.Error(),
	})
	_, _ = w.Write(b)
}

func (h *Handler) adminOrPluginRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID != "" {
			authorized, err := h.authorization.IsAuthorizedAdmin(userID)
			if err != nil {
				http.Error(w, "Not authorized", http.StatusUnauthorized)
				return
			}
			if authorized {
				next.ServeHTTP(w, r)
				return
			}
		}

		ifPluginId := r.Context().Value(PluginIDContextValue)
		pluginId, ok := ifPluginId.(string)
		if ok && pluginId != "" {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Not authorized", http.StatusUnauthorized)
	})
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, fromPluginID string) {
	h.root.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), PluginIDContextValue, fromPluginID)))
}

func (h *Handler) setLink(w http.ResponseWriter, r *http.Request) {
	var newLink link.Link
	if err := json.NewDecoder(r.Body).Decode(&newLink); err != nil {
		h.handleError(w, fmt.Errorf("Unable to decode body: %w", err))
		return
	}

	links := h.linkStore.GetLinks()
	found := false
	for i := range links {
		if links[i].Name == newLink.Name || links[i].Pattern == newLink.Pattern {
			links[i] = newLink
			found = true
			break
		}
	}
	if !found {
		links = append(h.linkStore.GetLinks(), newLink)
	}

	if err := h.linkStore.SaveLinks(links); err != nil {
		h.handleError(w, fmt.Errorf("Unable to save link: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "OK"}`))
}
