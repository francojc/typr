# Changelog

All notable changes to typr are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> typr is a fork of [tt](https://github.com/lemnos/tt) by Aetnaeus.
> Version history below reflects changes made after the fork baseline.

---

## [Unreleased]

## [1.0.0] - 2026-04-21

### Changed

- `typr -quotes` now defaults to ZenQuotes API (with zenlog fallback) -- no `-quotefile` flag needed
- `-quotefile zen` now reads exclusively from local zenlog cache (offline, no network)
- Clear error when zenlog is empty, prompting user to run `typr -quotes` first to populate it

### Added

- `NOTICE` file with attribution to upstream project (`tt` by Aetnaeus)
- GoReleaser configuration for cross-compiled binary releases
- GitHub Actions workflow for automated releases on tag push

---

## [0.9.0] - 2026-04-21

### Added

- Local zenlog offline cache: quotes fetched from ZenQuotes API are persisted to
  `~/.local/share/typr/quotes/zenlog.json` for offline fallback
- Deduplication logic: only unique quotes are stored in the zenlog
- `getQuoteWithFallback()`: tiered fallback -- API → in-memory cache → zenlog → hardcoded default
- Unit tests for zenlog load, append, deduplication, and fallback behavior

---

## [0.8.0] - 2026-01-27

### Added

- ZenQuotes API integration: `-quotefile zen` fetches a random inspirational quote
  from [zenquotes.io](https://zenquotes.io/) with a 5-second timeout
- In-memory quote cache (last 10 quotes) for session-level fallback
- Hardcoded fallback quote when API and cache are both unavailable

---

## [0.7.0] - 2026-01-09

### Added

- `visualize` subcommand: ASCII graph of typing speed progress over time
- Shows min, mean, and max WPM aggregated by day (default: last 30 days)
- Accepts a filename or full path to a `-csv` stats file
- Uses [asciigraph](https://github.com/guptarohit/asciigraph) library

---

## [0.6.0] - 2025-12-10

### Added

- YAML configuration file support (`~/.config/typr/config.yaml`)
- Automatic config file creation with commented defaults on first run
- `file` field added to CSV stats output (`timestamp,wpm,cpm,accuracy,file,n`)
- `n` field added to CSV stats to track test group size

### Changed

- Config format changed from JSON to YAML (JSON configs will be ignored)
- Config location: `~/.config/typr/config.yaml` (previously `config.json`)
- CSV output writes to files instead of stdout
- Stats: `~/.local/share/typr/results/{mode}-stats.csv`
- Errors: `~/.local/share/typr/results/{mode}-errors.csv`
- Flag precedence clarified: override > config > hardcoded default

### Fixed

- Config defaults now correctly applied across all flags

---

## [0.5.0] - 2025-12-07

### Changed

- `Esc` exits the application (previously restarted test)
- `Tab` restarts the current test during typing; starts a new test on results screen
- Improved `-quotes` vs `-quotefile` distinction and error messaging
- Keyboard shortcut documentation updated

---

## [0.4.0] - 2025-12-06

### Added

- Makefile `clean` target

### Changed

- Modernized codebase: replaced deprecated `io/ioutil` with `io` and `os` equivalents
- Removed `rand.Seed()` (auto-initialized since Go 1.20)
- Install prefix updated to user-local directory in Makefile

### Removed

- Binary file removed from repository

---

## [0.1.0] - 2025-12-06

### Added

- Fork baseline from [tt](https://github.com/lemnos/tt) by Aetnaeus
- Retained all upstream features: word mode, quote mode, themes, CSV output,
  multi-mode, XDG Base Directory compliance, `Ctrl-W` word deletion

---

[Unreleased]: https://github.com/francojc/typr/compare/v1.0.0...main
[1.0.0]: https://github.com/francojc/typr/compare/v0.9.0...v1.0.0
[0.9.0]: https://github.com/francojc/typr/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/francojc/typr/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/francojc/typr/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/francojc/typr/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/francojc/typr/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/francojc/typr/compare/v0.1.0...v0.4.0
[0.1.0]: https://github.com/francojc/typr/releases/tag/v0.1.0
