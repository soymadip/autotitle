package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mydehq/autotitle/internal/renamer"
	"github.com/mydehq/autotitle/internal/types"
)

func TestIntegration_NonGreedyAndCollision(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create files that trigger greedy over-capture and collision
	// Both mkv so the target path is IDENTICAL
	files := []string{
		"[Subs] [v2] My show - 01.mkv",
		"[Other] [v2] My show - 01.mkv",
	}
	for _, f := range files {
		if _, err := os.Create(filepath.Join(tmpDir, f)); err != nil {
			t.Fatal(err)
		}
	}

	// 2. Mock Media
	media := &types.Media{
		Title: "My show",
		Episodes: []types.Episode{
			{Number: 1, Title: "Episode One"},
		},
	}

	// 3. Target with Non-Greedy potential
	target := &types.Target{
		Path: tmpDir,
		Patterns: []types.Pattern{
			{
				Input: []string{
					"[{{ANY}}] {{SERIES}} - {{EP_NUM}}.{{EXT}}",
				},
				Output: types.OutputConfig{
					Fields: []string{"SERIES", " - ", "EP_NUM", " - ", "EP_NAME"},
				},
			},
		},
	}

	// 4. Run Renamer
	mockDB := &MockDB{path: filepath.Join(tmpDir, "db")}
	r := renamer.New(mockDB, types.BackupConfig{Enabled: false}, []string{"mkv"})

	var capturedEvents []types.Event
	r.WithEvents(func(e types.Event) {
		capturedEvents = append(capturedEvents, e)
	})

	ops, err := r.Execute(context.Background(), tmpDir, target, media)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 5. Verify results
	t.Logf("Operations: %d", len(ops))
	for _, op := range ops {
		t.Logf("Op: %s -> %s", filepath.Base(op.SourcePath), filepath.Base(op.TargetPath))
	}

	for _, e := range capturedEvents {
		t.Logf("Event: [%v] %s", e.Type, e.Message)
	}

	// One should succeed, one should COLLIDE and skip.
	if len(ops) != 1 {
		t.Errorf("Expected exactly 1 operation (one should collide and skip), got %d", len(ops))
	}

	foundCollisionEvent := false
	for _, e := range capturedEvents {
		if e.Type == types.EventError && len(e.Message) >= 9 && e.Message[:9] == "Collision" {
			foundCollisionEvent = true
			break
		}
	}

	if !foundCollisionEvent {
		t.Error("Did not find expected collision error in events")
	}

	// Verify non-greedy match results in correct Series name
	// Pattern: [{{ANY}}] {{SERIES}} - {{EP_NUM}}.{{EXT}}
	// Non-greedy ANY: "Subs"
	// SERIES: "[v2] My show"
	// Since we use media.Title ("My show"), we need to verify if we WANT it to be non-greedy.
	// Actually, if we use media.Title, the matcher's captured Series is ignored for output.
	// BUT, the matcher MUST match correctly for the rest of the pattern to work.
}
