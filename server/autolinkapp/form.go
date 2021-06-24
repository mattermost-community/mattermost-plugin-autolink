package autolinkapp

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/utils/httputils"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-autolink/server/autolink"
)

const (
	createOptionValue = "_create"

	fieldLink                 = "link"
	fieldName                 = "name"
	fieldEnabled              = "enabled"
	fieldTemplate             = "template"
	fieldPattern              = "pattern"
	fieldScope                = "scope"
	fieldWordMatch            = "word_match"
	fieldDisableNonWordPrefix = "disable_non_word_prefix"
	fieldDisableNonWordSuffix = "disable_non_word_suffix"

	fieldTestInput     = "test_input"
	fieldTestOutput    = "test_output"
	fieldSubmitButtons = "submit_buttons"

	submitButtonSave   = "save"
	submitButtonDelete = "delete"
	submitButtonTest   = "test"
)

type FormValues struct {
	Link                 apps.SelectOption `json:"link"`
	Name                 string            `json:"name"`
	Enabled              bool              `json:"enabled"`
	Template             string            `json:"template"`
	Pattern              string            `json:"pattern"`
	Scope                string            `json:"scope"`
	WordMatch            bool              `json:"word_match"`
	DisableNonWordPrefix bool              `json:"disable_non_word_prefix"`
	DisableNonWordSuffix bool              `json:"disable_non_word_suffix"`

	TestInput  string `json:"test_input"`
	TestOutput string `json:"test_output"`

	SubmitButtons string `json:"submit_buttons"`
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

	if call.SelectedField == fieldLink {
		if values.Link.Value == createOptionValue || values.Link.Value == "" {
			httputils.WriteJSON(w, a.getEmptyFormResponse())
			return
		}

		links := a.store.GetLinks()
		for _, link := range links {
			if link.Name == values.Link.Value {
				values = formValuesFromAutolink(link)
				break
			}
		}
	}

	f := a.getForm(values)
	resp := apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: f,
	}

	httputils.WriteJSON(w, resp)
}

func (a *app) getForm(values FormValues) *apps.Form {
	createNewOption := apps.SelectOption{
		Label: "(create new)",
		Value: createOptionValue,
	}

	links := a.store.GetLinks()
	sort.Slice(links, func(i, j int) bool {
		return strings.Compare(strings.ToLower(links[i].Name), strings.ToLower(links[j].Name)) < 0
	})

	linkOptions := []apps.SelectOption{createNewOption}
	for _, link := range links {
		linkOptions = append(linkOptions, apps.SelectOption{
			Label: link.Name,
			Value: link.Name,
		})
	}

	linkOption := createNewOption
	if values.Link.Value != "" {
		linkOption = values.Link
	}

	return &apps.Form{
		Title:         "Edit Autolinks",
		SubmitButtons: fieldSubmitButtons,
		Call:          apps.NewCall("/edit-links"),
		Fields: []*apps.Field{
			{
				Name:                fieldLink,
				ModalLabel:          "Link",
				Type:                apps.FieldTypeStaticSelect,
				Value:               linkOption,
				SelectStaticOptions: linkOptions,
				SelectRefresh:       true,
			},
			{
				Name:       fieldName,
				ModalLabel: "Name",
				Type:       apps.FieldTypeText,
				Value:      values.Name,
				IsRequired: true,
			},
			{
				Name:       fieldEnabled,
				ModalLabel: "Enabled",
				Type:       apps.FieldTypeBool,
				Value:      values.Enabled,
			},
			{
				Name:        fieldWordMatch,
				ModalLabel:  "Word Match",
				Type:        apps.FieldTypeBool,
				Value:       values.WordMatch,
				Description: "If true uses the [word boundaries](https://www.regular-expressions.info/wordboundaries.html)",
			},
			{
				Name:       fieldDisableNonWordPrefix,
				ModalLabel: "Disable Non-word Prefix",
				Type:       apps.FieldTypeBool,
				Value:      values.DisableNonWordPrefix,
			},
			{
				Name:       fieldDisableNonWordSuffix,
				ModalLabel: "Disable Non-word Suffix",
				Type:       apps.FieldTypeBool,
				Value:      values.DisableNonWordSuffix,
			},
			{
				Name:       fieldScope,
				ModalLabel: "Scope",
				Type:       apps.FieldTypeText,
				Value:      values.Scope,
			},
			{
				Name:       fieldPattern,
				ModalLabel: "Pattern",
				Type:       apps.FieldTypeText,
				Value:      values.Pattern,
				IsRequired: true,
			},
			{
				Name:       fieldTemplate,
				ModalLabel: "Template",
				Type:       apps.FieldTypeText,
				Value:      values.Template,
				IsRequired: true,
			},
			{
				Name:        fieldTestInput,
				ModalLabel:  "Test Input",
				Type:        apps.FieldTypeText,
				TextSubtype: apps.TextFieldSubtypeTextarea,
				Value:       values.TestInput,
			},
			{
				Name:        fieldTestOutput,
				ModalLabel:  "Test Output",
				Type:        apps.FieldTypeText,
				TextSubtype: apps.TextFieldSubtypeTextarea,
				Value:       values.TestOutput,
				ReadOnly:    true,
			},
			{
				Name: fieldSubmitButtons,
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
	return a.getForm(FormValues{Enabled: true})
}

func (a *app) getEmptyFormResponse() *apps.CallResponse {
	return &apps.CallResponse{
		Type: apps.CallResponseTypeForm,
		Form: a.getEmptyForm(),
	}
}

func extractFormValues(values map[string]interface{}) (fv FormValues, err error) {
	b, err := json.Marshal(values)
	if err != nil {
		return fv, err
	}

	err = json.Unmarshal(b, &fv)
	return fv, err
}

func autolinkFromFormValues(fv FormValues) autolink.Autolink {
	scope := []string{}
	if len(fv.Scope) > 0 {
		scope = strings.Split(strings.Trim(fv.Scope, " "), " ")
	}

	return autolink.Autolink{
		Name:                 fv.Name,
		Pattern:              fv.Pattern,
		Template:             fv.Template,
		Scope:                scope,
		Disabled:             !fv.Enabled,
		WordMatch:            fv.WordMatch,
		DisableNonWordPrefix: fv.DisableNonWordPrefix,
		DisableNonWordSuffix: fv.DisableNonWordPrefix,
	}
}

func formValuesFromAutolink(link autolink.Autolink) FormValues {
	scope := ""
	if len(link.Scope) > 0 {
		scope = strings.Join(link.Scope, " ")
	}

	return FormValues{
		Name:                 link.Name,
		Enabled:              !link.Disabled,
		WordMatch:            link.WordMatch,
		DisableNonWordPrefix: link.DisableNonWordPrefix,
		DisableNonWordSuffix: link.DisableNonWordSuffix,
		Pattern:              link.Pattern,
		Template:             link.Template,
		Scope:                scope,
		Link: apps.SelectOption{
			Label: link.Name,
			Value: link.Name,
		},
	}
}
