package autolinkapp

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/utils/httputils"
	"github.com/pkg/errors"
)

const createOptionValue = "_create"

type FormValues struct {
	Link       *apps.SelectOption `json:"link"`
	Name       string             `json:"name"`
	Template   string             `json:"template"`
	Pattern    string             `json:"pattern"`
	TestInput  string             `json:"test_input"`
	TestOutput string             `json:"test_output"`

	KeepModalOpen bool               `json:"keep_modal_open"`
	SubmitButtons *apps.SelectOption `json:"submit_buttons"`
}

func (a *app) handleFormFetch(w http.ResponseWriter, r *http.Request) {
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

	f := a.getForm(values)
	resp := apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: f,
	}

	httputils.WriteJSON(w, resp)
}

func (a *app) handleFormSubmit(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write(manifestData)
}

func (a *app) getForm(values FormValues) *apps.Form {
	linkOption := values.Link
	if linkOption == nil {
		linkOption = &apps.SelectOption{
			Label: "Create New",
			Value: createOptionValue,
		}
	}

	creating := linkOption.Value == createOptionValue

	return &apps.Form{
		Title:         "Edit Autolinks",
		SubmitButtons: "submit_buttons",
		Call:          apps.NewCall("/edit-links"),
		Fields: []*apps.Field{
			{
				Name:       "link",
				ModalLabel: "Link",
				Type:       apps.FieldTypeStaticSelect,
				Value:      linkOption,
			},
			{
				Name:       "name",
				ModalLabel: "Name",
				Type:       apps.FieldTypeText,
				Value:      values.Name,
				ReadOnly:   !creating,
			},
			{
				Name:       "template",
				ModalLabel: "Template",
				Type:       apps.FieldTypeText,
				Value:      values.Template,
			},
			{
				Name:       "pattern",
				ModalLabel: "Pattern",
				Type:       apps.FieldTypeText,
				Value:      values.Pattern,
			},
			{
				Name:        "test_input",
				ModalLabel:  "Test Input",
				Type:        apps.FieldTypeText,
				TextSubtype: apps.TextFieldSubtypeTextarea,
				Value:       values.TestInput,
			},
			{
				Name:        "test_output",
				ModalLabel:  "Test Output",
				Type:        apps.FieldTypeText,
				TextSubtype: apps.TextFieldSubtypeTextarea,
				Value:       values.TestInput,
				ReadOnly:    true,
			},
			{
				Name:       "keep_open",
				ModalLabel: "Keep Modal Open",
				Type:       apps.FieldTypeBool,
				Value:      values.KeepModalOpen,
			},
			{
				Name: "submit_buttons",
				Type: apps.FieldTypeStaticSelect,
				SelectStaticOptions: []apps.SelectOption{
					{
						Label: "Test",
						Value: "test",
					},
					{
						Label: "Delete",
						Value: "delete",
					},
					{
						Label: "Save",
						Value: "save",
					},
				},
			},
		},
	}
}

func (a *app) getEmptyForm() *apps.Form {
	return a.getForm(FormValues{})
}

func extractFormValues(values map[string]interface{}) (fv FormValues, err error) {
	b, err := json.Marshal(values)
	if err != nil {
		return fv, err
	}

	err = json.Unmarshal(b, &fv)
	return fv, err
}

var createNewLinkOption = &apps.SelectOption{
	Label: "Create New",
	Value: "_create",
}
