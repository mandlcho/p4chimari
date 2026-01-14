# 🎯 P4CHIMARI

<p align="center">
  <img src="docs/pachimari.jpg" width="200" alt="P4CHIMARI Logo">
</p>

A powerful Windows tool to manage Perforce workspace changes with an Unreal Engine-style interface.

## ✨ Features

### 📁 Smart Folder Selection
- **Interactive folder picker** - Browse and tick folders you want to scan
- **Recent folders** - Quick access to your most-used folders
- **Multi-select support** - Scan multiple folders at once

### 🔍 Advanced Filtering
- **Filter by action type** - View only adds, edits, or deletes
- **Pending changes** - See what's already checked out
- **Unsaved assets** - Find modified files not checked out

### 🛠️ Powerful Actions
- **Checkout files** - Open files for edit (`p4 edit`)
- **Reconcile** - Auto-detect and sync all changes (`p4 reconcile`)
- **Revert files** - Restore files to P4 version (`p4 sync -f`)
- **UE-style view** - Changelist viewer like Unreal Engine

### 🎨 User Experience
- Animated spinner with elapsed time
- ASCII art header
- Progress indicators
- Double confirmation for dangerous operations

---

## 🚀 Quick Start

### Prerequisites
- **Go 1.21+** (for building)
- **Perforce CLI** (`p4`) in your PATH
- Active P4 workspace

### Installation

**Option 1: Easy Install (Recommended)**
1. Double-click `INSTALL.bat`
2. Wait for build to complete
3. Double-click `RUN.bat` to launch

**Option 2: Manual Build**
```bash
cd src
go build -o ../p4chimari.exe
```

---

## 📖 Usage

### Running P4CHIMARI

Double-click `RUN.bat` or run `p4chimari.exe` from anywhere.

### Workflow Example

1. **Launch** → Shows ASCII art and connection status
2. **Select Folders** → Choose which Content subfolders to scan
   - Browse: Interactive folder picker
   - Recent: Quick access to previous selections
3. **Scan** → Displays modified files with action types
4. **Filter** → Show only adds, edits, or deletes
5. **Take Action** → Checkout, reconcile, or revert selected files

---

## 📂 Project Structure

```
p4chimari/
├── 🎯 RUN.bat           ← CLICK THIS TO RUN!
├── 🔨 INSTALL.bat       ← Click this to build first
├── 📖 README.md         ← You are here
├── bin/                 ← Executable (auto-generated)
│   └── p4chimari.exe
├── src/                 ← Source code
│   ├── main.go
│   ├── ascii.go
│   ├── config.go
│   ├── folderpicker.go
│   └── viewchanges.go
└── docs/                ← Documentation & assets
    └── pachimari.jpg
```

**Just double-click RUN.bat** - that's it!

---

## ⚙️ Configuration

P4CHIMARI stores config in `~/.p4chimari.json`:
- Recent folder selections
- Usage statistics
- Last scan locations

---

## 🎯 Key Features Explained

### Filter by Action
```
Filter by action type:
  1. Show only Edits    - Modified files
  2. Show only Adds     - New files
  3. Show only Deletes  - Removed files
  4. Show All           - Everything
```

### Revert Files (⚠️ Dangerous)
Restores files to P4 version, **permanently deleting local changes**.
- Select individual files
- Double confirmation required
- Type "YES" to confirm

### Folder Picker
```
┌─ CATEGORIES ─────────┐
│ [ ] Blueprints       │
│ [✓] Maps             │
│ [✓] Characters       │
│ [ ] Audio            │
└──────────────────────┘
```

---

## 🤝 Contributing

Feel free to open issues or submit pull requests!

---

## 📝 License

MIT License - Feel free to use and modify!

---

<p align="center">
  Made with ❤️ for Unreal Engine + Perforce workflows
</p>
