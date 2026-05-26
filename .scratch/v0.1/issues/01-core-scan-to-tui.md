# Core Scan-To-TUI Path

Status: ready-for-agent

## Parent

.scratch/v0.1/plan.md

## What to build

Create the first runnable version of `big`: a user can run the binary with no arguments or with one folder path, the app scans the selected **Scan Root**, ranks sized **Root Entries**, and opens a fullscreen terminal interface that shows the largest entries first.

This slice should establish the project structure, the scan model, the real filesystem adapter, size formatting, CLI argument handling, and a minimal Bubble Tea screen. It does not need final visual polish or advanced scrolling behavior, but the app must run end-to-end.

## Acceptance criteria

- [ ] `cmd/big` exists and `go run ./cmd/big` scans the current folder.
- [ ] `go run ./cmd/big <path>` scans the specified folder.
- [ ] More than one argument prints a short usage error and exits non-zero.
- [ ] Missing paths and non-folder paths print clear errors and exit non-zero.
- [ ] Regular files and folders directly inside the **Scan Root** appear as **Root Entries**.
- [ ] Folder **Entry Size** is calculated recursively from readable contents.
- [ ] Hidden root files and folders are not excluded.
- [ ] Sized entries are sorted by **Entry Size** descending, with name as a stable tie-breaker.
- [ ] Sizes are formatted with base 1024, integer floor values, and units `B`, `KB`, `MB`, `GB`.
- [ ] The TUI starts in alternate screen mode and displays a header plus rows with a right-aligned size column.
- [ ] Unit tests cover size formatting, sized entry sorting, recursive folder sizing, and hidden root files using a fake filesystem where practical.

## Blocked by

None - can start immediately

## Comments

