package main

import (
	"fmt"
	"regexp"
)

// Link represents a pattern to autolink.
type Link struct {
	Name                 string
	Disabled             bool
	Pattern              string
	Template             string
	Scope                []string
	WordMatch            bool
	DisableNonWordPrefix bool
	DisableNonWordSuffix bool

	template  string
	re        *regexp.Regexp
	multipass bool
}

// DisplayName returns a display name for the link.
func (l Link) DisplayName() string {
	if l.Name != "" {
		return l.Name
	}
	return l.Pattern
}

// Compile compiles the link's regular expression
func (l *Link) Compile() error {
	if l.Disabled || len(l.Pattern) == 0 || len(l.Template) == 0 {
		return nil
	}

	pattern := l.Pattern
	template := l.Template
	multipass := false
	if !l.DisableNonWordPrefix {
		if l.WordMatch {
			pattern = `\b` + pattern
		} else {
			pattern = `(?P<MattermostNonWordPrefix>(^|\s))` + pattern
			template = `${MattermostNonWordPrefix}` + template
			multipass = true
		}
	}
	if !l.DisableNonWordSuffix {
		if l.WordMatch {
			pattern = pattern + `\b`
		} else {
			pattern = pattern + `(?P<MattermostNonWordSuffix>$|[\s\.\!\?\,\)])`
			template = template + `${MattermostNonWordSuffix}`
			multipass = true
		}
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	l.re = re
	l.template = template
	l.multipass = multipass

	return nil
}

// Replace will subsitute the regex's with the supplied links
func (l Link) Replace(message string) string {
	if l.re == nil {
		return message
	}

	// Since they don't consume, `\b`s require no special handling, can just ReplaceAll
	if !l.multipass {
		return l.re.ReplaceAllString(message, l.template)
	}

	// Replace one at a time
	in := []byte(message)
	out := []byte{}
	for {
		submatch := l.re.FindSubmatchIndex(in)
		if submatch == nil {
			break
		}
		out = append(out, in[:submatch[0]]...)
		out = l.re.Expand(out, []byte(l.template), in, submatch)
		in = in[submatch[1]:]
	}
	out = append(out, in...)
	return string(out)
}

// ToMarkdown prints a Link as a markdown list element
func (l Link) ToMarkdown(i int) string {
	text := "- "
	if i > 0 {
		text += fmt.Sprintf("%v: ", i)
	}
	if l.Name != "" {
		if l.Disabled {
			text += fmt.Sprintf("~~%s~~", l.Name)
		} else {
			text += fmt.Sprintf("%s", l.Name)
		}
	}
	if l.Disabled {
		text += fmt.Sprintf(" **Disabled**")
	}
	text += "\n"

	text += fmt.Sprintf("  - Pattern: `%s`\n", l.Pattern)
	text += fmt.Sprintf("  - Template: `%s`\n", l.Template)

	if l.DisableNonWordPrefix {
		text += fmt.Sprintf("  - DisableNonWordPrefix: `%v`\n", l.DisableNonWordPrefix)
	}
	if l.DisableNonWordSuffix {
		text += fmt.Sprintf("  - DisableNonWordSuffix: `%v`\n", l.DisableNonWordSuffix)
	}
	if len(l.Scope) != 0 {
		text += fmt.Sprintf("  - Scope: `%v`\n", l.Scope)
	}
	if l.WordMatch {
		text += fmt.Sprintf("  - WordMatch: `%v`\n", l.WordMatch)
	}
	return text
}

// ToConfig returns a JSON-encodable Link represented solely with map[string]
// interface and []string types, compatible with gob/RPC, to be used in
// SavePluginConfig
func (l Link) ToConfig() map[string]interface{} {
	return map[string]interface{}{
		"Name":                 l.Name,
		"Pattern":              l.Pattern,
		"Template":             l.Template,
		"Scope":                l.Scope,
		"DisableNonWordPrefix": l.DisableNonWordPrefix,
		"DisableNonWordSuffix": l.DisableNonWordSuffix,
		"WordMatch":            l.WordMatch,
		"Disabled":             l.Disabled,
	}
}
