package store

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
)

type Store interface {
	GetLinks() []autolink.Autolink
	SaveLinks([]autolink.Autolink) error
}

func SetLink(s Store, newLink autolink.Autolink) (bool, error) {
	links := s.GetLinks()
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
		links = append(s.GetLinks(), newLink)
		changed = true
	}
	if changed {
		if err := s.SaveLinks(links); err != nil {
			return false, errors.Wrap(err, "unable to save link")
		}
		return true, nil
	}

	return false, nil
}
