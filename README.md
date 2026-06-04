# big

big is a terminal utility for finding the largest immediate entries inside a folder.

It scans directory and ranks its direct files and folders by size. You can move items to the trash.

> [!IMPORTANT]  
> Windows-only. I'm sorry. Please write the code for your OS yourself. You might want to take a look at the `_windows.go` files.

<img width="500" alt="big" src="https://github.com/user-attachments/assets/d344b1f1-b64f-4773-a6cf-e92328dd8ac3" />

## Views

- Groups entries into views such as folders, common file extensions, and other entries.
- Sorts entries under 1 MiB alphabetically and shows them without a size.
- Use ←/→ to switch between groups.

## Usage

```sh
big [flags] [path]
```

If `path` is omitted, big scans the current directory.

Flags:

- `--age <days>`: minimum root entry age in whole days; `0` includes entries of any age. Positive values use the file modified date and folder creation date.

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

```sh
go build ./cmd/big
go test ./...
```

## Related

- [shed](https://github.com/pongo/shed) — CLI that moves stale root items from a selected folder into ~\Shed. `Go`
- If you need a more detailed scan, try ncdu or TreeSize.
