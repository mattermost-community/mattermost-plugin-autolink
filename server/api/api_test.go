package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
	"github.com/stretchr/testify/require"
)

type authorizeAll struct{}

func (authorizeAll) IsAuthorizedAdmin(string) (bool, error) {
	return true, nil
}

type linkStore struct {
	prev       []autolink.Autolink
	saveCalled *bool
	saved      *[]autolink.Autolink
}

func (s *linkStore) GetLinks() []autolink.Autolink {
	return s.prev
}

func (s *linkStore) SaveLinks(links []autolink.Autolink) error {
	*s.saved = links
	*s.saveCalled = true
	return nil
}

func TestSetLink(t *testing.T) {
	for _, tc := range []struct {
		name             string
		method           string
		prevLinks        []autolink.Autolink
		link             autolink.Autolink
		expectSaveCalled bool
		expectSaved      []autolink.Autolink
		expectStatus     int
	}{
		{
			name: "happy simple",
			link: autolink.Autolink{
				Name: "test",
			},
			expectStatus:     http.StatusOK,
			expectSaveCalled: true,
			expectSaved: []autolink.Autolink{
				autolink.Autolink{
					Name: "test",
				},
			},
		},
		{
			name: "add new link",
			link: autolink.Autolink{
				Name:     "test1",
				Pattern:  ".*1",
				Template: "test1",
			},
			prevLinks: []autolink.Autolink{
				autolink.Autolink{
					Name:     "test2",
					Pattern:  ".*2",
					Template: "test2",
				},
			},
			expectStatus:     http.StatusOK,
			expectSaveCalled: true,
			expectSaved: []autolink.Autolink{
				autolink.Autolink{
					Name:     "test2",
					Pattern:  ".*2",
					Template: "test2",
				},
				autolink.Autolink{
					Name:     "test1",
					Pattern:  ".*1",
					Template: "test1",
				},
			},
		}, {
			name: "replace link",
			link: autolink.Autolink{
				Name:     "test2",
				Pattern:  ".*2",
				Template: "new template",
			},
			prevLinks: []autolink.Autolink{
				autolink.Autolink{
					Name:     "test1",
					Pattern:  ".*1",
					Template: "test1",
				},
				autolink.Autolink{
					Name:     "test2",
					Pattern:  ".*2",
					Template: "test2",
				},
				autolink.Autolink{
					Name:     "test3",
					Pattern:  ".*3",
					Template: "test3",
				},
			},
			expectStatus:     http.StatusOK,
			expectSaveCalled: true,
			expectSaved: []autolink.Autolink{
				autolink.Autolink{
					Name:     "test1",
					Pattern:  ".*1",
					Template: "test1",
				},
				autolink.Autolink{
					Name:     "test2",
					Pattern:  ".*2",
					Template: "new template",
				},
				autolink.Autolink{
					Name:     "test3",
					Pattern:  ".*3",
					Template: "test3",
				},
			},
		},
		{
			name: "no change",
			link: autolink.Autolink{
				Name:     "test2",
				Pattern:  ".*2",
				Template: "test2",
			},
			prevLinks: []autolink.Autolink{
				autolink.Autolink{
					Name:     "test1",
					Pattern:  ".*1",
					Template: "test1",
				},
				autolink.Autolink{
					Name:     "test2",
					Pattern:  ".*2",
					Template: "test2",
				},
			},
			expectStatus:     http.StatusNotModified,
			expectSaveCalled: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var saved []autolink.Autolink
			var saveCalled bool

			h := NewHandler(
				&linkStore{
					prev:       tc.prevLinks,
					saveCalled: &saveCalled,
					saved:      &saved,
				},
				authorizeAll{},
			)

			body, err := json.Marshal(tc.link)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			method := "POST"
			if tc.method != "" {
				method = tc.method
			}
			r, err := http.NewRequest(method, "/api/v1/link", bytes.NewReader(body))
			require.NoError(t, err)

			r.Header.Set("Mattermost-Plugin-ID", "testfrom")
			r.Header.Set("Mattermost-User-ID", "testuser")

			h.ServeHTTP(w, r)
			require.Equal(t, tc.expectStatus, w.Code)
			require.Equal(t, tc.expectSaveCalled, saveCalled)
			require.Equal(t, tc.expectSaved, saved)
		})
	}
}
