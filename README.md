# big

big is a terminal utility for finding the largest immediate entries inside a folder.

It scans one directory, ranks its direct files and folders by size, and opens an interactive terminal view for navigation and cleanup.

## Features

- Shows direct files and folders under the scan root.
- Counts folder size recursively without listing nested files.
- Sorts entries under 1 MiB alphabetically and shows them without a size.
- Groups entries into views such as folders, common file extensions, and other entries.
- Opens or reveals the selected entry from the terminal.
- Moves selected entries to the operating system trash.
- Leaves symlinks and special entries unscanned.

## Usage

```sh
big [path]
```

If `path` is omitted, big scans the current directory.

⚠️ Large folders can take time to scan. Press `ctrl+c` to stop the scan.

Use ←/→ to switch between file groups.

## Controls

- `up` / `down`: move selection
- `left` / `right`: switch views
- `pgup` / `pgdown`: scroll by page
- `home` / `end`: jump to first or last entry
- `enter`: open selected entry
- `e`: reveal selected entry in the file manager
- `delete`: move selected entry to trash
- `q`, `esc`, or `ctrl+c`: quit

## Development

Build the CLI:

```sh
go build ./cmd/big
```

Run all tests:

```sh
go test ./...
```
