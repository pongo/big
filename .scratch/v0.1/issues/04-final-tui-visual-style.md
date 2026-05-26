# Final TUI Visual Style

Status: ready-for-agent

## Parent

.scratch/v0.1/plan.md

## What to build

Apply the final visual treatment for `big` so the fullscreen terminal interface is calm, readable, and polished on Windows terminals. This slice should keep the layout dense and useful: no decorative full-screen frame, no oversized typography, and no extra explanatory text beyond the list, header, and help footer.

## Acceptance criteria

- [ ] The TUI uses a calm dark theme by default without background detection.
- [ ] The header shows the **Scan Root** basename and is bright but not oversized.
- [ ] Drive roots and root-like paths use the cleaned path when basename would be empty.
- [ ] The selected row uses a contrasting background and a bold name.
- [ ] Folder rows are slightly more accented than file rows.
- [ ] File rows remain visually neutral.
- [ ] Symlink and junction rows use muted cyan/blue styling and keep the size column blank.
- [ ] The size column is right-aligned and monospaced.
- [ ] The footer/help styling is muted.
- [ ] There is no decorative frame around the full screen.
- [ ] The visual styling preserves the list's available height.
- [ ] Manual verification confirms the layout is readable for empty, short, and long lists.

## Blocked by

- .scratch/v0.1/issues/02-links-and-unreadable-contents.md
- .scratch/v0.1/issues/03-viewport-navigation-and-help.md

## Comments

