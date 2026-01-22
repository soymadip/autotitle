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

// GenerateFilenameFromFields builds filename from field list
// Fields can be field names (uppercase) or literal strings
// Only adds separator between non-empty values - no cleanup needed
func GenerateFilenameFromFields(fields []string, separator string, vars TemplateVars) string {
	if separator == "" {
		separator = " - "
	}
	
	fieldValues := map[string]string{
		"SERIES":  vars.Series,
		"EP_NUM":  padNumber(vars.EpNum, 3),
		"EP_NAME": vars.EpName,
		"FILLER":  vars.Filler,
		"RES":     vars.Res,
	}
	
	var parts []string
	for _, field := range fields {
		// Check if it's a field name (uppercase)
		if value, ok := fieldValues[field]; ok {
			if value != "" {
				parts = append(parts, value)
			}
		} else {
			// It's a literal string - always include
			parts = append(parts, field)
		}
	}
	
	return strings.Join(parts, separator) + "." + vars.Ext
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
