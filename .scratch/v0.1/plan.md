# Implementation Plan

Goal: implement `big` as a Windows-oriented Go CLI/TUI utility that shows the largest immediate entries in a selected folder.

## 1. Project Structure

- `cmd/big/main.go` - CLI argument parsing, scan startup, TUI startup.
- `internal/scan/` - root entry model, size calculation, sorting, real filesystem adapter.
- `internal/tui/` - Bubble Tea model, viewport rendering, help footer, styles.
- `CONTEXT.md` contains the shared glossary: `Scan Root`, `Root Entry`, `Entry Size`, `Size Ranking`.

## 2. CLI Behavior

- `big` scans the current folder.
- `big <path>` scans the given folder.
- More than one argument prints a short usage error and exits with a non-zero code.
- A missing path or a path that is not a folder prints a clear error and exits with a non-zero code.
- No flags, config files, or logs in the first version.

## 3. Scan Behavior

- Show all immediate root entries: files, folders, hidden entries, symlinks, and junctions.
- For files, use the file size.
- For folders, calculate the recursive total of readable contained files and folders.
- Do not follow symlinks or junctions during size calculation.
- Show symlinks and junctions as root entries without a size.
- Access errors inside folders do not crash the app and are not surfaced in the UI; unreadable contents simply do not contribute to the size.
- Failure to read the scan root itself exits with an error.

## 4. Sorting

- Entries with a size appear first, sorted by size descending.
- Equal-sized entries are sorted by name for stable output.
- Symlinks and junctions without a size appear after all sized entries, sorted by name.

## 5. Size Formatting

- Use base 1024.
- Supported units: `B`, `KB`, `MB`, `GB`.
- Use integer values only, floored rather than rounded.
- Show `0 B` for empty sized entries.
- Values above GB are still displayed in GB.

## 6. TUI

- Use Charm v2 imports:
  - `charm.land/bubbletea/v2`
  - `charm.land/bubbles/v2/help`
  - `charm.land/bubbles/v2/viewport`
  - `charm.land/lipgloss/v2`
- Use alternate screen/fullscreen mode.
- Run scanning synchronously before starting the TUI; no progress bar, cancellation UI, or loading state.
- Header:
  - show the scan root basename;
  - for `C:\` and root-like paths, show the cleaned path;
  - make the header bright but not oversized.
- List:
  - use `viewport` plus custom row rendering, not `bubbles/list`;
  - make the list fill all available height;
  - do not use pagination;
  - do not show the number of entries;
  - show `No entries` for an empty scan root;
  - keep the size column right-aligned and monospaced;
  - preserve the size column width for symlink/junction rows, but leave it blank;
  - render symlink/junction names as `name -> target` when the target is available, otherwise just `name`.
- Visual style:
  - use a calm dark theme by default, without background detection;
  - selected row uses a contrasting background and bold name;
  - folders and files are distinguished softly: folders are slightly more accented, files are neutral;
  - symlinks and junctions use a muted cyan/blue and have no size;
  - footer/help is muted;
  - no decorative frame around the full screen, so the list can use the available height.
- Keyboard:
  - `↑/↓` `up/down` move selection;
  - `pgup/pgdn` scroll;
  - `home/end` jump;
  - `q`, `esc` and `ctrl+c` quit;
  - `enter` does nothing.
- Mouse support is not needed in the first version.

## 7. Testing

- Prefer tests that do not use the real filesystem by introducing a small fakeable filesystem interface in `internal/scan`.
- Unit tests:
  - size formatting;
  - sorting sized entries;
  - symlink/junction entries without size after sized entries;
  - hidden root files are not excluded;
  - access errors inside folders are ignored.
- Use minimal real filesystem tests only for Windows-specific behavior when needed.
- Symlink tests should skip gracefully if Windows does not allow symlink creation.
- Do not add fragile TUI snapshot tests; verify the TUI manually after implementation.

## 8. Documentation

- Keep `CONTEXT.md` as a glossary only, without implementation details.
- Do not create an ADR for Charm v2; the choice comes directly from the product requirement and does not need a separate architectural record.

feat: plan Windows TUI disk usage scanner implementation