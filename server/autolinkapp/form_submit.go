package autolinkapp

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/md"
	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
	"github.com/pkg/errors"
)

func (a *app) handleFormSubmit(w http.ResponseWriter, r *http.Request) {
	call := &apps.CallRequest{}
	err := json.NewDecoder(r.Body).Decode(call)
	if err != nil {
		httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.Wrap(err, "failed to decode call request body")))
		return
	}

	values, err := extractFormValues(call.Values)
	if err != nil {
		httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.Wrap(err, "error extracting form values")))
		return
	}

	var res *apps.CallResponse
	switch values.SubmitButtons {
	case submitButtonSave:
		res = a.handleSaveLink(call)
	case submitButtonDelete:
		res = a.handleDeleteLink(call)
	case submitButtonTest:
		res = a.handleTestLink(call)
	default:
		httputils.WriteJSON(w, apps.NewErrorCallResponse(errors.Errorf("invalid submit button provided: '%s'", values.SubmitButtons)))
		return
	}

	httputils.WriteJSON(w, res)
}

func (a *app) handleSaveLink(call *apps.CallRequest) *apps.CallResponse {
	values, err := extractFormValues(call.Values)
	if err != nil {
		return apps.NewErrorCallResponse(errors.Wrap(err, "error extracting form values"))
	}

	newLink := autolinkFromFormValues(values)

	oldLinks := a.store.GetLinks()
	newLinks := []autolink.Autolink{}

	if values.Name == createOptionValue {
		return apps.NewErrorCallResponse(errors.Errorf("invalid name '%s'", values.Name))
	}

	for _, link := range oldLinks {
		if link.Name == values.Name {
			if link.Name != values.Link.Value {
				return apps.NewErrorCallResponse(errors.Errorf("there is already a link named '%s'", values.Name))
			}
			newLinks = append(newLinks, newLink) // Overwrite existing link
		} else if link.Name == values.Link.Value {
			newLinks = append(newLinks, newLink) // Rename existing link
		} else {
			newLinks = append(newLinks, link)
		}
	}

	if values.Link.Value == createOptionValue {
		newLinks = append(newLinks, newLink) // Create a new link
	}

	err = a.store.SaveLinks(newLinks)
	if err != nil {
		return apps.NewErrorCallResponse(errors.Wrap(err, "failed to save auto link"))
	}

	return &apps.CallResponse{
		Markdown: md.Markdownf("Saved auto link '%s'", values.Name),
	}
}

func (a *app) handleDeleteLink(call *apps.CallRequest) *apps.CallResponse {
	values, err := extractFormValues(call.Values)
	if err != nil {
		return apps.NewErrorCallResponse(errors.Wrap(err, "error extracting form values"))
	}

	links := a.store.GetLinks()
	out := []autolink.Autolink{}

	for _, link := range links {
		if link.Name != values.Link.Value {
			out = append(out, link)
		}
	}

	err = a.store.SaveLinks(out)
	if err != nil {
		return apps.NewErrorCallResponse(errors.Wrap(err, "error saving links"))
	}

	return &apps.CallResponse{
		Markdown: md.Markdownf("Deleted auto link '%s'", values.Link.Value),
	}
}

func (a *app) handleTestLink(call *apps.CallRequest) *apps.CallResponse {
	fv, err := extractFormValues(call.Values)
	if err != nil {
		return apps.NewErrorCallResponse(errors.Wrap(err, "error extracting form values"))
	}

	l := autolinkFromFormValues(fv)
	err = l.Compile()
	if err != nil {
		return apps.NewErrorCallResponse(errors.Wrap(err, "error compiling link"))
	}

	out := l.Replace(fv.TestInput)
	fv.TestOutput = out
	f := a.getForm(fv)

	resp := &apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: f,
	}

	return resp
}
