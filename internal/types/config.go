package types

import (
	"fmt"
	"path/filepath"
)

// Config represents the autotitle configuration file
type Config struct {
	Targets []Target `yaml:"targets"`
	BaseDir string   `yaml:"-"`
}

// Target represents a rename target in the configuration
type Target struct {
	Path      string    `yaml:"path"`
	URL       string    `yaml:"url"`                  // Provider URL (MAL, TMDB, etc.)
	FillerURL string    `yaml:"filler_url,omitempty"` // Optional filler source URL
	Patterns  []Pattern `yaml:"patterns"`
}

// Pattern represents input/output pattern configuration
type Pattern struct {
	Input  []string     `yaml:"input"`
	Output OutputConfig `yaml:"output"`
}

// OutputConfig represents output format configuration
type OutputConfig struct {
	Fields    []string `yaml:"fields,flow"`
	Separator string   `yaml:"separator,omitempty"`
	Offset    int      `yaml:"offset,omitempty"`  // Episode number offset
	Padding   int      `yaml:"padding,omitempty"` // Episode number padding (e.g. 2 -> 01, 3 -> 001)
}

// GlobalConfig represents the global configuration file (~/.config/autotitle/config.yml)
type GlobalConfig struct {
	MapFile  string       `yaml:"map_file"`
	Patterns []Pattern    `yaml:"patterns"`
	Formats  []string     `yaml:"formats"`
	API      APIConfig    `yaml:"api"`
	Backup   BackupConfig `yaml:"backup"`
}

// ResolveTarget finds the target configuration for a given path
func (c *Config) ResolveTarget(path string) (*Target, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	for i := range c.Targets {
		targetPath := c.Targets[i].Path
		if !filepath.IsAbs(targetPath) {
			// Resolve relative to map file location
			targetPath = filepath.Join(c.BaseDir, targetPath)
		}

		// Check if paths resolve to the same location
		tAbs, err := filepath.Abs(targetPath)
		if err == nil && tAbs == absPath {
			return &c.Targets[i], nil
		}

		// Support "." as an exact match for the base directory
		if targetPath == "." && absPath == c.BaseDir {
			return &c.Targets[i], nil
		}
	}

	return nil, fmt.Errorf("no target found for path: %s", path)
}
