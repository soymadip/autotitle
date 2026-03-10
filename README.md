<div align="center">

<img src="./src/img/icon.svg" alt="Autotitle" width="90">

# Autotitle

A CLI tool & Go library for automatically renaming media files (anime, TV shows) with proper titles and detecting filler episodes.


</div>


## Features

- 🎯 **Automatic Episode Renaming** - Pattern-based filename matching and generation
- 🎨 **Flexible Pattern Matching** - Support for multiple filename formats with `{{TEMPLATE}}` variables
- 🔖 **Filler Detection** - Automatically marks filler episodes with `[F]` tag
- 📚 **Episode Database** - Caches episode data from MyAnimeList and AnimeFillerList
- 🧠 **Smart Updates** - Auto-updates database when new episodes air
- 💾 **Smart Backups** - Automatic backup before renaming with restore capability
- 🏷️ **Metadata Tagging** - Embeds episode/series info into `.mkv` (mkvpropedit) and `.mp4`/`.m4v` (atomicparsley) files
- 📦 **Library & CLI** - Use as standalone tool or import as Go package

## Installation

### CLI Tool

```bash
# From my-repo (Arch Linux)
curl -sL https://mydehq.github.io/my-repo/install | bash
sudo pacman -S autotitle

# Or from AUR
paru -S autotitle

# In Windows, from winget
winget install mydehq.autotitle

# Using Nix
nix build github:mydehq/autotitle  # To build
nix run github:mydehq/autotitle    # To run directly
nix develop github:mydehq/autotitle # For development env

# Or build manually
git clone https://github.com/mydehq/autotitle.git && cd autotitle
mise install-bin
```

### As Library

```bash
go get github.com/mydehq/autotitle
```

## Quick Start

```bash
cd /path/to/anime/videos

# Initialize with URLs directly
autotitle init . -u "https://myanimelist.net/anime/XXXXX"

# Or create template config to edit manually
autotitle init .

# Edit _autotitle.yml, preview & add changes

# Perform rename
autotitle .

# Tag already-renamed files without re-renaming
autotitle tag .

# Rename without tagging
autotitle --no-tag .

# Restore if needed
autotitle undo .
```

## Basic Configuration

Running `autotitle init` creates `_autotitle.yml`:

```yaml
targets:
  - path: "."
    url: "https://myanimelist.net/anime/XXXXX/Series_Name"
    filler_url: "https://www.animefillerlist.com/shows/series-name"
    patterns:
      - input:
          - "Episode {{EP_NUM}} {{RES}}"
        output:
          fields: [SERIES, EP_NUM, FILLER, EP_NAME]
          offset: 0 # Optional: Offset local episode numbers (e.g. 1 -> 11)
```

## Documentation

📚 **[Full Documentation](https://mydehq.github.io/docs/autotitle)** — Complete guides, commands, flags, configuration reference, and [library API](https://mydehq.github.io/docs/autotitle/library)

## Data Sources

Currently Below Data sources are implemented:

### Episode Data

|                 Source                 | Type  |
| :------------------------------------: | :---: |
| [MyAnimeList](https://myanimelist.net) | Anime |

### Filler Info

|                       Source                        | Type  |
| :-------------------------------------------------: | :---: |
| [AnimeFillerList](https://www.animefillerlist.com/) | Anime |
