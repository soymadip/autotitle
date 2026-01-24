package renamer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mydehq/autotitle/internal/config"
	"github.com/mydehq/autotitle/internal/database"
	"github.com/mydehq/autotitle/internal/fetcher"
	"github.com/mydehq/autotitle/internal/logger"
	"github.com/mydehq/autotitle/internal/matcher"
)

// Renamer orchestrates the file renaming process
type Renamer struct {
	Config    *config.GlobalConfig
	MapConfig *config.MapConfig
	DB        *database.DB
	DryRun    bool
	NoBackup  bool
	Verbose   bool
	Quiet     bool
}

type FileResult struct {
	Original  string
	NewName   string
	Error     error
	IsRenamed bool
}

// Execute performs the renaming on the target directory
func (r *Renamer) Execute(targetPath string) error {
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}

	targetConfig, err := r.MapConfig.ResolveTarget(".")
	if err != nil {
		return fmt.Errorf("failed to resolve target config: %w", err)
	}

	if targetConfig.MALURL == "" {
		return fmt.Errorf("mal_url is missing in configuration")
	}

	malID := fetcher.ExtractMALID(targetConfig.MALURL)
	if malID == 0 {
		return fmt.Errorf("invalid mal_url in config: could not extract ID")
	}

	seriesSlug := fmt.Sprintf("%d", malID)

	seriesData, err := r.DB.Load(seriesSlug)
	if err != nil {
		return fmt.Errorf("failed to load database for '%s' (ID: %d): %w", seriesSlug, malID, err)
	}
	if seriesData == nil {
		return fmt.Errorf("database not found for '%s'. Run 'autotitle db gen' first.", seriesSlug)
	}

	files, err := r.scanFiles(absPath)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		logger.Warn("No video files found in %s", absPath)
		return nil
	}

	var plans []FileResult

	patterns := targetConfig.Patterns
	if len(patterns) == 0 {
		patterns = r.Config.Patterns
	}

	type PatternPair struct {
		Matcher      *matcher.Pattern
		OutputConfig config.OutputConfig
	}

	var pairs []PatternPair

	for _, p := range patterns {
		if len(p.Output.Fields) == 0 {
			continue
		}

		for _, inputTmpl := range p.Input {
			cp, err := matcher.Compile(inputTmpl)

			if err != nil {
				logger.Error("Pattern error: %v", err)
				continue
			}

			pairs = append(pairs, PatternPair{
				Matcher:      cp,
				OutputConfig: p.GetOutputConfig(),
			})
		}
	}

	// Process files
	for _, file := range files {
		var bestMatch map[string]string
		var matchedPair *PatternPair

		matched := false
		for i, pair := range pairs {
			vars := pair.Matcher.Match(filepath.Base(file))
			if vars != nil {
				bestMatch = vars
				matchedPair = &pairs[i]
				matched = true
				break
			}
		}

		if !matched {
			if r.Verbose {
				logger.Info("Skipping non-matching file: %s", filepath.Base(file))
			}
			continue
		}

		epNumStr := bestMatch["EpNum"]
		epNum, _ := parseEpNum(epNumStr) // Handles "01", "1" etc

		epData, ok := seriesData.Episodes[epNum]
		if !ok {
			if r.Verbose {
				logger.Warn("Episode %d not found in database for file %s", epNum, filepath.Base(file))
			}
		}

		// Use anime title from database
		seriesName := seriesData.Title
		if seriesName == "" {
			seriesName = "Unknown"
		}

		// Construct TemplateVars
		tv := matcher.TemplateVars{
			Series:   seriesName,
			SeriesEn: seriesData.TitleEnglish,
			SeriesJp: seriesData.TitleJapanese,
			EpNum:    epNumStr,
			EpName:   epData.Title,
			Res:      bestMatch["Res"],
			Ext:      bestMatch["Ext"],
		}
		if epData.Filler {
			tv.Filler = "[F]"
		}

		newName := matcher.GenerateFilenameFromFields(
			matchedPair.OutputConfig.Fields,
			matchedPair.OutputConfig.Separator,
			tv,
		)

		if newName == filepath.Base(file) {
			continue
		}

		plans = append(plans, FileResult{
			Original: file,
			NewName:  filepath.Join(filepath.Dir(file), newName),
		})
	}

	if len(plans) == 0 {
		if !r.Quiet {
			logger.Info("No files to rename.")
		}
		return nil
	}

	backupDir := filepath.Join(absPath, r.Config.Backup.DirName)
	if !r.DryRun && !r.NoBackup {
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return fmt.Errorf("failed to create backup dir: %w", err)
		}
	}
	undoLog := make(map[string]string)

	for _, plan := range plans {
		dest := plan.NewName

		if r.DryRun {
			if !r.Quiet {
				relOld, _ := filepath.Rel(absPath, plan.Original)
				relNew, _ := filepath.Rel(absPath, dest)
				fmt.Printf("%s -> %s\n", relOld, relNew)
			}
			continue
		}

		if !r.NoBackup {
			backupPath := filepath.Join(backupDir, filepath.Base(plan.Original))
			if err := copyFile(plan.Original, backupPath); err != nil {
				if !r.Quiet {
					logger.Error("Backup failed for %s: %v", filepath.Base(plan.Original), err)
				}
				continue
			}
		}

		if err := os.Rename(plan.Original, dest); err != nil {
			if !r.Quiet {
				logger.Error("Rename failed: %v", err)
			}
		} else {
			undoLog[filepath.Base(plan.Original)] = dest
		}
	}

	if !r.NoBackup && len(undoLog) > 0 {
		logPath := filepath.Join(backupDir, "undo.json")
		data, err := json.MarshalIndent(undoLog, "", "  ")
		if err == nil {
			_ = os.WriteFile(logPath, data, 0644)
		}
	}

	if !r.Quiet {
		logger.Success("Processed %d files.", len(plans))
	}

	return nil
}

// scanFiles returns list of supported video files
func (r *Renamer) scanFiles(root string) ([]string, error) {
	var files []string
	formats := make(map[string]bool)
	for _, f := range r.Config.Formats {
		formats["."+strings.ToLower(f)] = true
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if formats[ext] {
			files = append(files, filepath.Join(root, entry.Name()))
		}
	}
	return files, nil
}

// Clean removes the backup directory
func (r *Renamer) Clean(targetPath string) error {
	backupDir := filepath.Join(targetPath, r.Config.Backup.DirName)
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("no backup directory found at %s", backupDir)
	}
	return os.RemoveAll(backupDir)
}

// Undo restores files from backup
func (r *Renamer) Undo(targetPath string) error {
	backupDir := filepath.Join(targetPath, r.Config.Backup.DirName)
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("no backup directory found to undo from")
	}

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return err
	}

	undoLog := make(map[string]string)
	logPath := filepath.Join(backupDir, "undo.json")
	if data, err := os.ReadFile(logPath); err == nil {
		_ = json.Unmarshal(data, &undoLog)
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "undo.json" {
			continue
		}

		src := filepath.Join(backupDir, entry.Name())
		dest := filepath.Join(targetPath, entry.Name())

		if renamedFile, ok := undoLog[entry.Name()]; ok {
			if err := os.Remove(renamedFile); err != nil {
				logger.Warn("Failed to remove renamed file %s: %v", renamedFile, err)
			}
		}

		if err := os.Rename(src, dest); err != nil {
			logger.Error("Failed to restore %s: %v", entry.Name(), err)
		} else {
			count++
		}
	}

	os.Remove(logPath)

	os.Remove(backupDir)

	logger.Success("Restored %d files.", count)
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func parseEpNum(s string) (int, error) {
	// Atoi handles "01" as 1
	return strconv.Atoi(s)
}
