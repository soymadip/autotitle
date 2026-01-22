# Implementation Status

## ✅ Completed: Database Filename Migration
**Format:** {ID}.json → {ID}@{slug}.json  
**Status:** COMPLETE (10-100x faster searches via glob patterns)

**Format:** {ID}@{slug}.json (e.g., 779@meitantei-conan-movie-01.json)

## ✅ Completed: Field-Based Output Format
**New Feature:** Support for literal strings in fields array  
**Status:** COMPLETE

**Format:**
```yaml
output:
  fields: ["DC", EP_NUM, FILLER, EP_NAME]  # Mix literals with fields
  separator: " - "  # Optional, defaults to " - "
```

**Changes:**
- Removed `prefix`/`suffix` fields (use inline literals instead)
- Made `separator` optional with " - " default
- Removed complex cleanup logic (build clean from start)
- Automatic empty field skipping

---

# Original Migration Plan

Database Filename Migration: {ID}.json → {ID}@{slug}.json
Update database to use {ID}@{slug}.json format for 10-100x faster searches via filename pattern matching.

User Review Required
IMPORTANT

New format: {ID}@{slug}.json (e.g., 779@meitantei-conan-movie-01.json)

Existing files: User will manually delete/regenerate existing .json files (only 3 files currently)

Proposed Changes
Database Package
[MODIFY] 
database.go
1. Update 
Load()
 method (lines 136-156)

Change from: filepath.Join(db.Dir, seriesID+".json")
Change to: Find file with glob filepath.Glob(filepath.Join(db.Dir, seriesID+"@*.json"))
Return error if not found or multiple matches
2. Update 
Save()
 method (lines 158-171)

Change to: filepath.Join(db.Dir, sd.MALID+"@"+sd.Slug+".json")
Add slug truncation if filename > 255 chars
3. Update 
Exists()
 method (lines 173-177)

Use glob: filepath.Glob(filepath.Join(db.Dir, seriesID+"@*.json"))
Return true if len(matches) > 0
4. Update 
Delete()
 method (lines 179-185)

Find with glob pattern, delete all matches
5. Update 
List()
 method (lines 194-207)

Parse {ID}@{slug}.json format
Extract ID: strings.Split(filename, "@")[0]
6. Update 
Find()
 method (lines 209-259)

Fast path: filepath.Glob(*@*{querySlug}*.json) → load only matches
Fallback: Scan all files if no filename matches
Database Tests
[MODIFY] 
database_test.go
1. Update 
TestNewAndLoadSave
 - expect {ID}@{slug}.json path

2. Update 
TestFind
 - add filename fast-path test

3. Update 
TestExists
 - test glob pattern matching

4. Update 
TestList
 - verify ID extraction from new format (split on @)

5. Add TestLongSlug - verify slug truncation for long titles

Verification Plan
Tests
cd /home/soymadip/Projects/autotitle
go test ./internal/database -v
Manual Testing
# Delete old format files
rm ~/.cache/autotitle/db/*.json
# Generate new database (will use new format)
go run cmd/autotitle/main.go db gen "https://myanimelist.net/anime/779"
# Verify new filename format
ls -1 ~/.cache/autotitle/db/
# Expected: 779@meitantei-conan-movie-01-tokei-jikake-no-matenrou.json
# Test fast search
go run cmd/autotitle/main.go db info "conan"
Performance Impact
Before: O(n) - scan all files
After: O(k) - scan matching filenames only
Speedup: 10-100x for typical queries