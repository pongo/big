# Viewport Navigation And Help

Status: ready-for-agent

## Parent

.scratch/v0.1/plan.md

## What to build

Complete the terminal interaction model so the list behaves like a focused fullscreen viewer. The rows should use a custom `viewport` renderer, remain scrollable without pagination, keep the selected row visible, and expose only real keyboard actions through the Charm help footer.

## Acceptance criteria

- [ ] The TUI uses `charm.land/bubbles/v2/viewport` plus custom row rendering rather than `bubbles/list`.
- [ ] The list fills the available height between header and footer.
- [ ] The UI does not show pagination or the number of entries.
- [ ] `up/down` move the selected row.
- [ ] `pgup/pgdn` scroll through the list.
- [ ] `home/end` jump to the first or last row.
- [ ] `q`, `esc`, and `ctrl+c` quit.
- [ ] `enter` has no effect.
- [ ] The selected row remains visible while moving through long lists.
- [ ] Empty scan roots show `No entries` while keeping the footer available.
- [ ] The footer uses `charm.land/bubbles/v2/help`.
- [ ] Mouse support is not added.

## Blocked by

- .scratch/v0.1/issues/01-core-scan-to-tui.md

## Comments

