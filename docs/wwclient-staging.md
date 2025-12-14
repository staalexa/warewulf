# wwclient File Staging Feature

## Overview

The wwclient now includes a git-like staging feature that allows users to stage files locally and then export them to Warewulf overlays. This is useful for preparing overlay content on client nodes before deploying it.

## Commands

### `wwclient add`

Stage a file for later export to an overlay.

**Usage:**
```bash
wwclient add SOURCE [DEST]
```

**Examples:**
```bash
# Add a file using its original name in root directory
wwclient add /tmp/myfile.txt

# Add a file with a specific destination path in the overlay
wwclient add /tmp/config.txt /etc/myapp/config.txt

# Use custom staging directory
wwclient add /tmp/script.sh /usr/local/bin/script.sh --staging-dir=/custom/staging
```

**Options:**
- `-s, --staging-dir string`: Directory to use for staging (default: `/var/lib/warewulf/staging`)

### `wwclient status`

Show all currently staged files.

**Usage:**
```bash
wwclient status
```

**Example:**
```bash
wwclient status --staging-dir=/tmp/staging
```

**Output:**
```
Staged files (2):

  /etc/myapp/config.txt
    Source: /tmp/config.txt
    Size: 21 bytes
    Added: 2025-12-14 20:28:52

  /usr/local/bin/startup.sh
    Source: /tmp/startup.sh
    Size: 37 bytes
    Added: 2025-12-14 20:29:00
```

### `wwclient export`

Export all staged files to a Warewulf overlay.

**Usage:**
```bash
wwclient export OVERLAY_NAME
```

**Examples:**
```bash
# Export to an overlay
wwclient export my-overlay

# Export and clear staging area
wwclient export my-overlay --clear

# Use custom staging directory
wwclient export my-overlay --staging-dir=/custom/staging
```

**Options:**
- `-c, --clear`: Clear staging area after successful export
- `-s, --staging-dir string`: Directory to use for staging (default: `/var/lib/warewulf/staging`)

## Workflow Example

Here's a complete workflow demonstrating the staging feature:

```bash
# 1. Check current staging status (should be empty)
$ wwclient status
No files staged

# 2. Stage some files
$ wwclient add /tmp/config.txt /etc/myapp/config.txt
Staged file: /tmp/config.txt -> /etc/myapp/config.txt
Added file to staging: /tmp/config.txt
  Destination: /etc/myapp/config.txt

$ wwclient add /tmp/startup.sh /usr/local/bin/startup.sh
Staged file: /tmp/startup.sh -> /usr/local/bin/startup.sh
Added file to staging: /tmp/startup.sh
  Destination: /usr/local/bin/startup.sh

# 3. Check staging status
$ wwclient status
Staged files (2):

  /etc/myapp/config.txt
    Source: /tmp/config.txt
    Size: 21 bytes
    Added: 2025-12-14 20:28:52

  /usr/local/bin/startup.sh
    Source: /tmp/startup.sh
    Size: 37 bytes
    Added: 2025-12-14 20:29:00

# 4. Export to overlay and clear staging
$ wwclient export my-overlay --clear
Exported: /tmp/config.txt -> my-overlay:/etc/myapp/config.txt
Exported: /tmp/startup.sh -> my-overlay:/usr/local/bin/startup.sh

Successfully exported 2 file(s) to overlay 'my-overlay'
Cleared all staged files
Staging area cleared

# 5. Verify staging was cleared
$ wwclient status
No files staged
```

## How It Works

1. **Staging**: When you add a file using `wwclient add`, the file is copied to a staging directory (default: `/var/lib/warewulf/staging/files`) along with metadata stored in a JSON index file.

2. **Persistence**: The staging index is saved to disk, so staged files persist across wwclient invocations.

3. **Export**: When you export staged files, they are copied from the staging directory into the specified overlay's rootfs directory structure. If the overlay doesn't exist, it will be created.

4. **Overlay Management**: The export command automatically handles site vs. distribution overlays, cloning to site overlays when necessary.

## Use Cases

- **Configuration Management**: Stage configuration files on client nodes before deploying them to overlays
- **Batch Operations**: Collect multiple files before exporting them all at once
- **Testing**: Stage files, review them, and then export when ready
- **Workflow Automation**: Script the staging and export process for repeatable deployments

## Notes

- Only regular files are supported (not directories)
- Files are copied (not moved) to the staging area
- The staging directory can be customized per command or set globally
- Exported overlays are created as site overlays if they don't exist
- The `--clear` flag is useful to automatically clean up after export
