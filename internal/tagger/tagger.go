// Package tagger embeds metadata into media files using mkvpropedit (MKV)
// and AtomicParsley (MP4/M4V/M4A).
package tagger

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	mkvBin = "mkvpropedit"
	mp4Bin = "atomicparsley"
)

// TagInfo contains the metadata to embed into a media file.
type TagInfo struct {
	Title       string // Episode title
	Show        string // Series name
	EpisodeID   string // Formatted episode number (e.g. "01")
	EpisodeSort int    // Numeric episode number (for sorting)
	AirDate     string // ISO date string (e.g. "2013-04-07"), optional
}

// IsAvailable returns true if at least one supported tagging tool is in $PATH.
func IsAvailable() bool {
	return IsMKVAvailable() || IsMP4Available()
}

// IsMKVAvailable returns true if mkvpropedit is in $PATH.
func IsMKVAvailable() bool {
	_, err := exec.LookPath(mkvBin)
	return err == nil
}

// IsMP4Available returns true if AtomicParsley is in $PATH.
func IsMP4Available() bool {
	_, err := exec.LookPath(mp4Bin)
	return err == nil
}

// isMKV returns true if the file has an .mkv extension (used in tests).
func isMKV(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".mkv")
}

// isTaggable returns true if the file format is supported for tagging.
func isTaggable(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mkv", ".mp4", ".m4v", ".m4a":
		return true
	}
	return false
}

// TagFile embeds metadata into a media file, dispatching based on file extension:
//   - .mkv          → mkvpropedit
//   - .mp4/.m4v/.m4a → AtomicParsley
//
// Unsupported extensions are silently skipped (returns nil).
// Returns an error if the required tool is not installed for the given format.
func TagFile(ctx context.Context, path string, info TagInfo) error {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".mkv":
		if !IsMKVAvailable() {
			return fmt.Errorf("mkvpropedit not found; cannot tag %s", filepath.Base(path))
		}
		return tagMKV(ctx, path, info)

	case ".mp4", ".m4v", ".m4a":
		if !IsMP4Available() {
			return fmt.Errorf("atomicparsley not found; cannot tag %s", filepath.Base(path))
		}
		return tagMP4(ctx, path, info)

	default:
		// Unsupported format — silently skip
		return nil
	}
}

// MKV via mkvpropedit
func tagMKV(ctx context.Context, path string, info TagInfo) error {
	tmpFile, err := os.CreateTemp("", "autotitle-tags-*.xml")
	if err != nil {
		return fmt.Errorf("failed to create temp tag file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if err := writeTagXML(tmpFile, info); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write tag XML: %w", err)
	}
	tmpFile.Close()

	args := []string{
		path,
		"--edit", "info",
		"--set", fmt.Sprintf("title=%s", info.Title),
		"--tags", fmt.Sprintf("all:%s", tmpFile.Name()),
	}

	cmd := exec.CommandContext(ctx, mkvBin, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mkvpropedit failed: %w\noutput: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// tagXMLTemplate is the Matroska global tag XML format.
const tagXMLTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE Tags SYSTEM "matroskatags.dtd">
<Tags>
  <Tag>
    <Targets>
      <TargetTypeValue>50</TargetTypeValue>
      <TargetType>SHOW</TargetType>
    </Targets>
    <Simple>
      <Name>TITLE</Name>
      <String>{{.Show}}</String>
    </Simple>
  </Tag>
  <Tag>
    <Targets>
      <TargetTypeValue>30</TargetTypeValue>
      <TargetType>CHAPTER</TargetType>
    </Targets>
    <Simple>
      <Name>TITLE</Name>
      <String>{{.Title}}</String>
    </Simple>{{if .EpisodeID}}
    <Simple>
      <Name>PART_NUMBER</Name>
      <String>{{.EpisodeID}}</String>
    </Simple>{{end}}{{if .AirDate}}
    <Simple>
      <Name>DATE_RELEASED</Name>
      <String>{{.AirDate}}</String>
    </Simple>{{end}}
  </Tag>
</Tags>
`

var tagTmpl = template.Must(template.New("tags").Parse(tagXMLTemplate))

func writeTagXML(f *os.File, info TagInfo) error {
	return tagTmpl.Execute(f, info)
}

// MP4/M4V/M4A via AtomicParsley
func tagMP4(ctx context.Context, path string, info TagInfo) error {
	args := []string{path, "--overWrite"}

	if info.Title != "" {
		args = append(args, "--title", info.Title)
	}
	if info.Show != "" {
		args = append(args, "--TVShowName", info.Show)
	}
	if info.EpisodeID != "" {
		args = append(args, "--TVEpisode", info.EpisodeID)
	}
	if info.EpisodeSort > 0 {
		args = append(args, "--TVEpisodeNum", fmt.Sprintf("%d", info.EpisodeSort))
	}
	if info.AirDate != "" {
		// AtomicParsley --year accepts full ISO dates or just a year
		args = append(args, "--year", info.AirDate)
	}

	cmd := exec.CommandContext(ctx, mp4Bin, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("AtomicParsley failed: %w\noutput: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
