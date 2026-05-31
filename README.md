# big

big is a terminal utility for finding the largest immediate entries inside a folder.

It scans directory and ranks its direct files and folders by size.

<img width="500" alt="big" src="https://github.com/user-attachments/assets/d344b1f1-b64f-4773-a6cf-e92328dd8ac3" />

## Views

- Groups entries into views such as folders, common file extensions, and other entries.
- Sorts entries under 1 MiB alphabetically and shows them without a size.
- Use ←/→ to switch between groups.

## Usage

```sh
big [path]
```

If `path` is omitted, big scans the current directory.

⚠️ Large folders can take time to scan. Press `ctrl+c` to stop the scan.

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
