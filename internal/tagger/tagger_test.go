package tagger

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// renderTagXML is a test helper that renders the tag XML template to a string.
func renderTagXML(t *testing.T, info TagInfo) string {
	t.Helper()
	f, err := os.CreateTemp("", "tagger-test-*.xml")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	if err := writeTagXML(f, info); err != nil {
		f.Close()
		t.Fatalf("writeTagXML: %v", err)
	}
	f.Close()

	data, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return string(data)
}

func TestWriteTagXML_AllFields(t *testing.T) {
	info := TagInfo{
		Title:     "To You, in 2000 Years",
		Show:      "Attack on Titan",
		EpisodeID: "01",
		AirDate:   "2013-04-07",
	}

	xml := renderTagXML(t, info)
	assertContains(t, xml, "Attack on Titan")
	assertContains(t, xml, "To You, in 2000 Years")
	assertContains(t, xml, "PART_NUMBER")
	assertContains(t, xml, "01")
	assertContains(t, xml, "DATE_RELEASED")
	assertContains(t, xml, "2013-04-07")
	assertContains(t, xml, "TargetTypeValue")
}

func TestWriteTagXML_NoAirDate(t *testing.T) {
	info := TagInfo{
		Title:     "Episode Title",
		Show:      "My Anime",
		EpisodeID: "05",
	}

	xml := renderTagXML(t, info)
	if strings.Contains(xml, "DATE_RELEASED") {
		t.Error("Should not include DATE_RELEASED when AirDate is empty")
	}
	assertContains(t, xml, "Episode Title")
	assertContains(t, xml, "My Anime")
	assertContains(t, xml, "05")
}

func TestIsMKV(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/path/to/file.mkv", true},
		{"/path/to/file.MKV", true},
		{"/path/to/file.mp4", false},
		{"/path/to/file.avi", false},
		{"/path/to/file", false},
	}
	for _, c := range cases {
		got := isMKV(c.path)
		if got != c.want {
			t.Errorf("isMKV(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

func TestIsTaggable(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/path/to/file.mkv", true},
		{"/path/to/file.MKV", true},
		{"/path/to/file.mp4", true},
		{"/path/to/file.MP4", true},
		{"/path/to/file.m4v", true},
		{"/path/to/file.m4a", true},
		{"/path/to/file.avi", false},
		{"/path/to/file.webm", false},
		{"/path/to/file.ts", false},
		{"/path/to/file", false},
	}
	for _, c := range cases {
		got := isTaggable(c.path)
		if got != c.want {
			t.Errorf("isTaggable(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}

// Verify the template is valid XML (basic sanity)
func TestWriteTagXML_ValidXML(t *testing.T) {
	info := TagInfo{Title: "Test", Show: "Series"}
	xml := renderTagXML(t, info)
	if !strings.HasPrefix(strings.TrimSpace(xml), "<?xml") {
		t.Errorf("Expected XML declaration, got: %s", xml[:min(50, len(xml))])
	}
	assertContains(t, xml, "<Tags>")
	assertContains(t, xml, "</Tags>")
}

// TestTagFile_MKV_Integration creates a real MKV with ffmpeg, tags it, and verifies with mkvinfo.
func TestTagFile_MKV_Integration(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found; skipping MKV integration test")
	}
	if !IsMKVAvailable() {
		t.Skip("mkvpropedit not found; skipping MKV integration test")
	}

	tmpDir := t.TempDir()
	mkvPath := tmpDir + "/ep01.mkv"
	ffmpegArgs := []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1",
		"-f", "lavfi", "-i", "color=c=black:s=64x64:d=1",
		"-map", "0:a", "-map", "1:v",
		"-c:v", "libx264", "-c:a", "aac", "-shortest",
		mkvPath, "-y", "-loglevel", "quiet",
	}
	if out, err := exec.Command("ffmpeg", ffmpegArgs...).CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg failed to create test MKV: %v\n%s", err, out)
	}

	info := TagInfo{
		Title:     "To You, in 2000 Years",
		Show:      "Attack on Titan",
		EpisodeID: "01",
		AirDate:   "2013-04-07",
	}

	if err := TagFile(context.Background(), mkvPath, info); err != nil {
		t.Fatalf("TagFile (MKV) failed: %v", err)
	}

	// mkvinfo --all is required to show the Tags section
	out, err := exec.Command("mkvinfo", "--all", mkvPath).CombinedOutput()
	if err != nil {
		t.Fatalf("mkvinfo failed: %v\n%s", err, out)
	}
	outStr := string(out)

	for _, want := range []string{"Attack on Titan", "To You, in 2000 Years", "2013-04-07"} {
		if !strings.Contains(outStr, want) {
			t.Errorf("mkvinfo output missing %q\nFull output:\n%s", want, outStr)
		}
	}
	t.Logf("✓ MKV tags verified: title=%q show=%q date=%q", info.Title, info.Show, info.AirDate)
}

// TestTagFile_MP4_Integration creates a real MP4 with ffmpeg, tags it, and verifies with AtomicParsley.
func TestTagFile_MP4_Integration(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found; skipping MP4 integration test")
	}
	if !IsMP4Available() {
		t.Skip("AtomicParsley not found; skipping MP4 integration test")
	}

	tmpDir := t.TempDir()
	mp4Path := tmpDir + "/ep01.mp4"
	ffmpegArgs := []string{
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1",
		"-f", "lavfi", "-i", "color=c=black:s=64x64:d=1",
		"-map", "0:a", "-map", "1:v",
		"-c:v", "libx264", "-c:a", "aac", "-shortest",
		mp4Path, "-y", "-loglevel", "quiet",
	}
	if out, err := exec.Command("ffmpeg", ffmpegArgs...).CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg failed to create test MP4: %v\n%s", err, out)
	}

	info := TagInfo{
		Title:       "To You, in 2000 Years",
		Show:        "Attack on Titan",
		EpisodeID:   "01",
		EpisodeSort: 1,
		AirDate:     "2013-04-07",
	}

	if err := TagFile(context.Background(), mp4Path, info); err != nil {
		t.Fatalf("TagFile (MP4) failed: %v", err)
	}

	// Verify with atomicparsley -t (print all tags)
	out, err := exec.Command(mp4Bin, mp4Path, "-t").CombinedOutput()
	if err != nil {
		t.Fatalf("AtomicParsley -t failed: %v\n%s", err, out)
	}
	outStr := string(out)

	for _, want := range []string{"Attack on Titan", "To You, in 2000 Years"} {
		if !strings.Contains(outStr, want) {
			t.Errorf("AtomicParsley output missing %q\nFull output:\n%s", want, outStr)
		}
	}
	t.Logf("✓ MP4 tags verified: title=%q show=%q", info.Title, info.Show)
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("Expected output to contain %q\nGot:\n%s", needle, haystack)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
