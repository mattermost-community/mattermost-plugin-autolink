package main

import (
	"errors"
	"regexp"
)

// AutoLinker helper for replace regex with links
type AutoLinker struct {
	link *Link
	re   *regexp.Regexp
}

// NewAutoLinker create and initialize a AutoLinker
func NewAutoLinker(link Link) (*AutoLinker, error) {
	if len(link.Pattern) == 0 || len(link.Template) == 0 {
		return nil, errors.New("Pattern or template was empty")
	}

	if !link.DisableNonWordPrefix {
		link.Pattern = `\b` + link.Pattern
	}

	if !link.DisableNonWordSuffix {
		link.Pattern = link.Pattern + `\b`
	}

	re, err := regexp.Compile(link.Pattern)
	if err != nil {
		return nil, err
	}

	return &AutoLinker{
		link: &link,
		re:   re,
	}, nil
}

// Replace will subsitute the regex's with the supplied links
func (l *AutoLinker) Replace(message string) string {
	if l.re == nil {
		return message
	}

	return l.re.ReplaceAllString(message, l.link.Template)
}
