# 🎯 Hijacked Files Guide

## The Problem

When working with Unreal Engine 5 and Perforce, you face two types of unwanted checkouts:

1. **Auto-checkout on open**: UE5 checks out files just by opening the editor
2. **Dependency checkouts**: When you modify a file, UE5 auto-checks out related dependency files

**Result**: Your changelist gets cluttered with files you don't actually want to submit.

## The Solution

P4CHIMARI provides two powerful commands to handle hijacked files:

### Option 6: Show Hijacked Files Status

This command analyzes your currently opened files and separates them into two categories:

```
📊 Hijacked Files Analysis
─────────────────────────────────────
Total opened files:     47
  Real changes:         12 (26%)
  Hijacked (unchanged): 35 (74%)
─────────────────────────────────────

✓ Real Changes:
  • //depot/Project/Content/Maps/MainLevel.umap
  • //depot/Project/Content/Blueprints/BP_Player.uasset
  ... and 10 more

⚠️  Hijacked Files (unchanged):
  • //depot/Project/Config/DefaultEngine.ini
  • //depot/Project/Content/Materials/M_Default.uasset
  ... and 33 more
```

**How it works:**
- Uses `p4 diff -sr` to find files that are opened but have no actual changes
- Shows you exactly which files are safe to revert
- Gives you confidence before cleaning up

### Option 7: Revert Hijacked Files

This command automatically reverts all hijacked files (files with no actual changes):

```
🔄 Finding hijacked files (opened but unchanged)...
─────────────────────────────────────
Found 35 hijacked file(s):
  • //depot/Project/Config/DefaultEngine.ini
  • //depot/Project/Content/Materials/M_Default.uasset
  ...

⚠️  This will revert 35 unchanged file(s)
Proceed? (yes/no): yes

Reverting hijacked files...
✓ Done! Hijacked files have been reverted.
  Your real changes remain checked out.
```

**How it works:**
- Uses `p4 revert -a` to automatically revert unchanged files
- Safe operation - only reverts files with zero modifications
- Keeps all your actual work intact

## Workflow Example

### Before P4CHIMARI
```
You: *Opens UE5*
UE5: *Checks out 40 files automatically*
You: *Makes changes to 3 blueprints*
UE5: *Checks out 7 more dependency files*
You: *Tries to submit*
You: "Wait... which of these 50 files do I actually need??"
You: *Spends 15 minutes cherry-picking files*
```

### After P4CHIMARI
```
You: *Opens UE5, works normally*
You: *Runs p4chimari.exe*
You: *Selects Option 6 to see status*
Tool: "12 real changes, 38 hijacked"
You: *Selects Option 7*
Tool: "Reverted 38 hijacked files"
You: *Submits clean changelist with only 12 files*
You: ✨ "That took 30 seconds!"
```

## Best Practices

### When to Use

1. **After working session**: Before creating a changelist
2. **Before submitting**: Final cleanup to ensure clean commits
3. **Regular checkups**: Run periodically to keep workspace clean

### Safety Tips

- **Always run Option 6 first** to see what will be reverted
- **Review the list** before confirming Option 7
- **Don't worry**: Only unchanged files are reverted - your work is safe

### What Gets Reverted

✅ **WILL be reverted (safe):**
- Files checked out but not modified
- Auto-opened config files
- Dependency files with no changes
- Files opened by accident

❌ **WON'T be reverted (protected):**
- Files with actual code/content changes
- Files you edited
- New files you added
- Modified blueprints/assets

## Technical Details

### p4 diff -sr
Lists files that are opened for edit but have no content differences from the depot version.

### p4 revert -a
Reverts only those files that are opened for edit but have no content, type, or resolved-integration changes.

This is a **safe operation** because P4 will not revert files with actual modifications.

## Troubleshooting

### "No hijacked files found"
This means all your opened files have actual changes. Nothing to clean up!

### "Error: Unable to connect to P4"
Ensure your Perforce connection is active and you're in the workspace root.

### Too many files to review?
Use Option 6 first to generate a summary. The tool shows the first 10-20 files of each category, which is usually enough to verify it's working correctly.

---

**Pro Tip**: Make this part of your daily workflow - run it before every submit to keep your changelists clean and reviewable!
