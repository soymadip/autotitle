// Package autotitle provides high-level functions for renaming anime episodes
// with proper titles and filler detection.
//
// This package mirrors the CLI functionality and provides a compatible API
// for integrating autotitle into other Go applications.

package autotitle

import "github.com/soymadip/autotitle/internal/api"

// Re-export all types from internal/api
type (
	Option       = api.Option
	Options      = api.Options
	DBGenOptions = api.DBGenOptions
	Pattern      = api.Pattern
	TemplateVars = api.TemplateVars
	EpisodeData  = api.EpisodeData
	SeriesData   = api.SeriesData
)

// Re-export all option constructors
var (
	WithDryRun   = api.WithDryRun
	WithNoBackup = api.WithNoBackup
	WithVerbose  = api.WithVerbose
	WithQuiet    = api.WithQuiet
	WithConfig   = api.WithConfig
	WithForce    = api.WithForce
)

// Re-export all core functions
var (
	Rename         = api.Rename
	Undo           = api.Undo
	Clean          = api.Clean
	DBGen          = api.DBGen
	Init           = api.Init
	DBPath         = api.DBPath
	DBList         = api.DBList
	DBInfo         = api.DBInfo
	DBRm           = api.DBRm
	CompilePattern = api.CompilePattern
	ExtractMALID   = api.ExtractMALID
)
