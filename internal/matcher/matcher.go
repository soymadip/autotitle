package matcher

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	FieldGlue           = "+"
	PlaceholderSeries   = "{{SERIES}}"
	PlaceholderSeriesEn = "{{SERIES_EN}}"
	PlaceholderSeriesJp = "{{SERIES_JP}}"
	PlaceholderEpNum    = "{{EP_NUM}}"
	PlaceholderEpName   = "{{EP_NAME}}"
	PlaceholderFiller   = "{{FILLER}}"
	PlaceholderRes      = "{{RES}}"
	PlaceholderExt      = "{{EXT}}"
	PlaceholderAny      = "{{ANY}}"
)

var (
	// placeholderRegexMap maps placeholder base names to their regex definitions
	placeholderRegexMap = map[string]string{
		"SERIES":    ".+?",
		"SERIES_EN": ".+?",
		"SERIES_JP": ".+?",
		"EP_NUM":    `\d+`,
		"EP_NAME":   ".+?",
		"FILLER":    ".*?",
		"RES":       `\d{3,4}p|\d{3,4}x\d{3,4}`,
		"ANY":       ".*?",
	}
)

type TemplateVars struct {
	Series   string
	SeriesEn string
	SeriesJp string
	EpNum    string
	EpName   string
	Filler   string
	Res      string
	Ext      string
}

// MatchResult contains extracted values from a filename match
type MatchResult struct {
	EpisodeNum int
	Resolution string
	Extension  string
}

type Pattern struct {
	raw      string
	regex    *regexp.Regexp
	idxEpNum int
	idxRes   int
}

func (p *Pattern) String() string {
	return p.regex.String()
}

// Compile compiles a template string into a regex pattern.
// Supports multiple occurrences of the same placeholder by generating
// unique named capture groups (e.g., Any_1, Any_2).
func Compile(template string) (*Pattern, error) {
	templateBase := strings.ReplaceAll(template, "."+PlaceholderExt, "")
	templateBase = strings.ReplaceAll(templateBase, PlaceholderExt, "")

	regexStr := regexp.QuoteMeta(templateBase)

	// Replace placeholders in a single pass using unique group names.
	rePlaceholderFinder := regexp.MustCompile(`\\{\\{([A-Z_]+)\\}\\}`)

	placeholderCounts := make(map[string]int)

	allMatches := rePlaceholderFinder.FindAllStringSubmatch(regexStr, -1)
	totals := make(map[string]int)
	for _, m := range allMatches {
		totals[m[1]]++
	}

	resultRegex := rePlaceholderFinder.ReplaceAllStringFunc(regexStr, func(m string) string {
		match := rePlaceholderFinder.FindStringSubmatch(m)
		baseName := match[1]
		placeholderRegex, ok := placeholderRegexMap[baseName]
		if !ok {
			// Unknown placeholder, treat as literal
			return m
		}

		placeholderCounts[baseName]++
		count := placeholderCounts[baseName]

		groupName := formatGroupName(baseName)

		// Only add suffix if there are multiple occurrences of this specific placeholder
		if totals[baseName] > 1 {
			groupName = fmt.Sprintf("%s_%d", groupName, count)
		}

		return fmt.Sprintf("(?P<%s>%s)", groupName, placeholderRegex)
	})

	resultRegex = "^" + resultRegex + "$"

	re, err := regexp.Compile(resultRegex)
	if err != nil {
		return nil, fmt.Errorf("failed to compile pattern %q: %w (regex: %s)", template, err, resultRegex)
	}

	return &Pattern{
		raw:      template,
		regex:    re,
		idxEpNum: getFirstSubexpIndex(re, "EpNum"),
		idxRes:   getFirstSubexpIndex(re, "Res"),
	}, nil
}

func formatGroupName(baseName string) string {
	parts := strings.Split(baseName, "_")
	var groupName string
	for _, p := range parts {
		if len(p) > 0 {
			groupName += strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return groupName
}

func getFirstSubexpIndex(re *regexp.Regexp, baseName string) int {
	if idx := re.SubexpIndex(baseName); idx >= 0 {
		return idx
	}
	return re.SubexpIndex(baseName + "_1")
}

// Match attempts to match a filename against the compiled pattern
func (p *Pattern) Match(filename string) map[string]string {
	ext := filepath.Ext(filename)
	if ext == "" {
		return nil
	}

	nameWithoutExt := strings.TrimSuffix(filename, ext)
	match := p.regex.FindStringSubmatch(nameWithoutExt)

	if match == nil {
		return nil
	}

	result := make(map[string]string)
	for i, name := range p.regex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	result["Ext"] = strings.TrimPrefix(ext, ".")
	return result
}

// MatchTyped attempts to match a filename and returns a structured result
func (p *Pattern) MatchTyped(filename string) (*MatchResult, bool) {
	ext := filepath.Ext(filename)
	if ext == "" {
		return nil, false
	}

	nameWithoutExt := strings.TrimSuffix(filename, ext)
	match := p.regex.FindStringSubmatch(nameWithoutExt)

	if match == nil {
		return nil, false
	}

	var epNum int
	if p.idxEpNum >= 0 && p.idxEpNum < len(match) {
		valStr := match[p.idxEpNum]
		if val, err := strconv.Atoi(valStr); err == nil {
			epNum = val
		}
	}

	var res string
	if p.idxRes >= 0 && p.idxRes < len(match) {
		res = match[p.idxRes]
	}

	return &MatchResult{
		EpisodeNum: epNum,
		Resolution: res,
		Extension:  strings.TrimPrefix(ext, "."),
	}, true
}

// GenerateFilenameFromFields builds filename from field list
func GenerateFilenameFromFields(fields []string, separator string, vars TemplateVars, padding int) (string, error) {
	if padding <= 0 {
		padding = 3
	}

	var builder strings.Builder
	first := true
	suppressNextSep := false

	for _, field := range fields {

		// Handle Glue Operator
		if field == FieldGlue {
			suppressNextSep = true
			continue
		}

		value, err := resolveField(field, vars, padding)
		if err != nil {
			return "", err
		}

		if value == "" {
			continue
		}

		if !first && !suppressNextSep {
			builder.WriteString(separator)
		}

		builder.WriteString(value)
		first = false
		suppressNextSep = false
	}

	builder.WriteString(".")
	builder.WriteString(vars.Ext)

	return builder.String(), nil
}

func resolveField(field string, vars TemplateVars, padding int) (string, error) {
	switch field {
	case "SERIES":
		return vars.Series, nil
	case "SERIES_EN":
		return vars.SeriesEn, nil
	case "SERIES_JP":
		return vars.SeriesJp, nil
	case "EP_NUM":
		return padNumber(vars.EpNum, padding), nil
	case "EP_NAME":
		return vars.EpName, nil
	case "FILLER":
		return vars.Filler, nil
	case "RES":
		return vars.Res, nil
	}

	// Check if it's explicitly quoted (to allow using "SERIES" as a literal)
	if len(field) >= 2 && field[0] == '"' && field[len(field)-1] == '"' {
		return field[1 : len(field)-1], nil
	}

	return field, nil
}

// padNumber pads a number string with zeros to width
func padNumber(s string, width int) string {

	if s == "" {
		return ""
	}

	if len(s) >= width {
		return s
	}

	return strings.Repeat("0", width-len(s)) + s
}
