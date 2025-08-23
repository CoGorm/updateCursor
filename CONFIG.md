# Configuration Guide

updateCursor supports flexible configuration through a YAML file.

## Configuration File Location

The configuration file is located at:
```
~/.config/updateCursor/config.yaml
```

If the file doesn't exist, updateCursor will create it with default values on first run.

## Configuration Options

| Setting | Description | Default Value |
|---------|-------------|---------------|
| `download_dir` | Directory where Cursor AppImage files are stored | `~/Downloads/Cursor` |
| `file_name_pattern` | Pattern for downloaded filenames (use `<version>` placeholder) | `Cursor-<version>-x86_64.AppImage` |
| `latest_symlink` | Path to symlink pointing to current version | `~/Downloads/Cursor/Cursor.AppImage` |
| `ledger_path` | Path to update history log file | `~/.config/updateCursor/cursor-versions.log` |

## Example Configurations

### Default Configuration
```yaml
download_dir: "~/Downloads/Cursor"
file_name_pattern: "Cursor-<version>-x86_64.AppImage"
latest_symlink: "~/Downloads/Cursor/Cursor.AppImage"
ledger_path: "~/.config/updateCursor/cursor-versions.log"
```

### Custom User Setup
```yaml
download_dir: "~/Applications/Cursor"
file_name_pattern: "Cursor_<version>.AppImage"
latest_symlink: "~/.local/bin/Cursor.AppImage"
ledger_path: "~/.config/updateCursor/cursor-versions.log"
```

### System-wide Installation
```yaml
download_dir: "/opt/cursor/versions"
file_name_pattern: "cursor-<version>-linux-x64.AppImage"
latest_symlink: "/usr/local/bin/cursor"
ledger_path: "/var/log/cursor-updates.log"
```

### Portable Setup
```yaml
download_dir: "./cursor-versions"
file_name_pattern: "Cursor-<version>.AppImage"
latest_symlink: "./cursor"
ledger_path: "./cursor-versions.log"
```

### Development Environment
```yaml
download_dir: "~/Development/tools/cursor"
file_name_pattern: "cursor-dev-<version>.AppImage"
latest_symlink: "~/bin/cursor-dev"
ledger_path: "~/Development/tools/cursor/updates.log"
```

## Getting Started

1. **Use defaults**: Just run `updatecursor` - it will create a default config automatically
2. **Customize**: Copy `config.example.yaml` to `~/.config/updateCursor/config.yaml` and edit
3. **Validate**: Run `updatecursor check` to verify your configuration works

## Path Expansion

- `~` is automatically expanded to your home directory
- Relative paths are supported
- All directories will be created automatically if they don't exist

## Filename Pattern

The `file_name_pattern` must contain `<version>` which will be replaced with the actual Cursor version (e.g., `1.4.5`).

Examples:
- `Cursor-<version>-x86_64.AppImage` → `Cursor-1.4.5-x86_64.AppImage`
- `Cursor_<version>.AppImage` → `Cursor_1.4.5.AppImage`
- `cursor-<version>-linux.AppImage` → `cursor-1.4.5-linux.AppImage`
