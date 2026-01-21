package matcher

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	PlaceholderSeries = "{{SERIES}}"
	PlaceholderEpNum  = "{{EP_NUM}}"
	PlaceholderEpName = "{{EP_NAME}}"
	PlaceholderFiller = "{{FILLER}}"
	PlaceholderRes    = "{{RES}}"
	PlaceholderExt    = "{{EXT}}"
	PlaceholderAny    = "{{ANY}}"
)

type TemplateVars struct {
	Series string
	EpNum  string
	EpName string
	Filler string
	Res    string
	Ext    string
}

type Pattern struct {
	raw   string
	regex *regexp.Regexp
}

// Compile compiles a template string into a regex pattern
func Compile(template string) (*Pattern, error) {
	regexStr := regexp.QuoteMeta(template)

	// Replace placeholders with named capture groups
	regexStr = strings.ReplaceAll(regexStr, regexp.QuoteMeta(PlaceholderSeries), "(?P<Series>.+)")
	regexStr = strings.ReplaceAll(regexStr, regexp.QuoteMeta(PlaceholderEpNum), "(?P<EpNum>\\d+)")
	regexStr = strings.ReplaceAll(regexStr, regexp.QuoteMeta(PlaceholderEpName), "(?P<EpName>.+)")
	regexStr = strings.ReplaceAll(regexStr, regexp.QuoteMeta(PlaceholderFiller), "(?P<Filler>.*)")
	regexStr = strings.ReplaceAll(regexStr, regexp.QuoteMeta(PlaceholderRes), "(?P<Res>\\d{3,4}p)")
	regexStr = strings.ReplaceAll(regexStr, regexp.QuoteMeta(PlaceholderExt), "(?P<Ext>[a-zA-Z0-9]+)")
	regexStr = strings.ReplaceAll(regexStr, regexp.QuoteMeta(PlaceholderAny), "(?P<Any>.*)")

	// Anchor the regex to match full string
	regexStr = "^" + regexStr + "$"

	re, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile pattern %q: %w", template, err)
	}

	return &Pattern{
		raw:   template,
		regex: re,
	}, nil
}

// Match attempts to match a filename against the compiled pattern
func (p *Pattern) Match(filename string) map[string]string {
	match := p.regex.FindStringSubmatch(filename)
	if match == nil {
		return nil
	}

	result := make(map[string]string)
	for i, name := range p.regex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}

// GenerateFilename creates a new filename using the output template and variables
func GenerateFilename(outputTemplate string, vars TemplateVars) string {
	res := outputTemplate
	res = strings.ReplaceAll(res, PlaceholderSeries, vars.Series)
	res = strings.ReplaceAll(res, PlaceholderEpNum, padNumber(vars.EpNum, 3))
	res = strings.ReplaceAll(res, PlaceholderEpName, vars.EpName)

	// Filler logic: if vars.Filler is not empty (e.g. "[F]"), use it
	res = strings.ReplaceAll(res, PlaceholderFiller, vars.Filler)

	res = strings.ReplaceAll(res, PlaceholderRes, vars.Res)
	res = strings.ReplaceAll(res, PlaceholderExt, vars.Ext)

	// Clean up double spaces that might occur if a placeholder is empty
	res = strings.ReplaceAll(res, "  ", " ")
	res = strings.TrimSpace(res)

	return res
}

// padNumber pads a number string with zeros to width
func padNumber(s string, width int) string {
	if s == "" {
		return ""
	}
	if len(s) < width {
		return strings.Repeat("0", width-len(s)) + s
	}
	return s
}
