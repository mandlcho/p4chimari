# p4chimari

A simple Windows tool to manage modified but not checked-out files in your Perforce workspace.

## Features

- Connects to your local P4 workspace automatically
- Finds files that have been modified but are not checked out
- Lets you choose to either:
  - **Check out** files (`p4 edit`)
  - **Reconcile** files (`p4 reconcile`)

## Prerequisites

- Go 1.21 or higher (for building)
- Perforce CLI (`p4`) installed and in your PATH
- Active P4 workspace configured

## Building

```bash
go build -o p4chimari.exe
```

## Usage

1. Navigate to your P4 workspace directory
2. Run `p4chimari.exe`
3. The tool will scan for modified files and present options

```
p4chimari - Perforce Workspace Helper
=====================================

Connected to P4 workspace: your-workspace

Scanning for modified files not checked out...

Found 3 modified file(s) not checked out:
  1. C:\workspace\src\main.cpp
  2. C:\workspace\config\settings.ini
  3. C:\workspace\README.md

What would you like to do?
  1. Check out files (p4 edit)
  2. Reconcile files (p4 reconcile)
  3. Cancel

Enter choice (1-3):
```

## How It Works

1. Verifies P4 is available and you're in a workspace
2. Runs `p4 reconcile -n ...` to preview changes (non-destructive)
3. Presents the list of modified files
4. Executes your choice:
   - **Check out**: Runs `p4 edit` on each file
   - **Reconcile**: Runs `p4 reconcile ...` to sync workspace state

## Notes

- The tool operates on the current directory and subdirectories (`...`)
- No files are modified until you make a choice
- Reconcile will also detect new and deleted files, not just edits
