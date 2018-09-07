package main

import (
	"errors"
	"regexp"
)

// AutoLinker helper for replace regex with links
type AutoLinker struct {
	link    *Link
	pattern *regexp.Regexp
}

// NewAutoLinker create and initialize a AutoLinker
func NewAutoLinker(link *Link) (*AutoLinker, error) {
	if link == nil || len(link.Pattern) == 0 || len(link.Template) == 0 {
		return nil, errors.New("Pattern or template was empty")
	}

	if !link.DisableNonWordPrefix {
		link.Pattern = "(?P<MMDisableNonWordPrefix>^|\\s)" + link.Pattern
		link.Template = "${MMDisableNonWordPrefix}" + link.Template
	}

	if !link.DisableNonWordSuffix {
		link.Pattern = link.Pattern + "(?P<DisableNonWordSuffix>$|\\s|\\.|\\!|\\?|\\,|\\))"
		link.Template = link.Template + "${DisableNonWordSuffix}"
	}

	p, err := regexp.Compile(link.Pattern)
	if err != nil {
		return nil, err
	}

	return &AutoLinker{
		link:    link,
		pattern: p,
	}, nil
}

// Replace will subsitute the regex's with the supplied links
func (l *AutoLinker) Replace(message string) string {
	if l.pattern == nil {
		return message
	}

	// beacuse MMDisableNonWordPrefix DisableNonWordSuffix are greedy then
	// two matches back to back won't get found.  So we need to run the
	// replace all twice
	if !l.link.DisableNonWordPrefix && !l.link.DisableNonWordSuffix {
		message = string(l.pattern.ReplaceAll([]byte(message), []byte(l.link.Template)))
	}

	return string(l.pattern.ReplaceAll([]byte(message), []byte(l.link.Template)))
}
