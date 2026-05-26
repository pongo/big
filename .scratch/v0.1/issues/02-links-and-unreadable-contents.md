# Links And Unreadable Contents

Status: ready-for-agent

## Parent

.scratch/v0.1/plan.md

## What to build

Teach `big` to handle symlinks, junctions, and unreadable nested contents according to the product semantics. Symlinks and junctions remain visible as **Root Entries**, but they have no **Entry Size**, are not followed, and appear after all sized entries. Nested access errors are tolerated without logs or visible warnings.

## Acceptance criteria

- [ ] Symlink and junction root entries are shown in the list.
- [ ] Symlink and junction root entries do not display a size.
- [ ] Symlink and junction root entries are never followed while calculating size.
- [ ] Symlink and junction root entries appear after every sized root entry.
- [ ] Symlink and junction root entries are sorted by name within their no-size group.
- [ ] When a link target can be read, the row renders as `name -> target`; otherwise it renders as `name`.
- [ ] Access errors while scanning nested folder contents do not crash the app.
- [ ] Unreadable nested contents do not contribute to folder **Entry Size** and are not surfaced in the TUI.
- [ ] Tests cover no-follow ranking and unreadable nested contents with a fake filesystem where practical.
- [ ] Any real Windows symlink/junction test skips gracefully when the OS does not allow creating the required link.

## Blocked by

- .scratch/v0.1/issues/01-core-scan-to-tui.md

## Comments

