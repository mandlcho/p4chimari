# P4CHIMARI

A Windows tool to manage Perforce workspace changes with an Unreal Engine-style workflow (scan → review → reconcile/checkout/revert), including “hijacked file” cleanup for UE projects.

## Problem

UE projects + Perforce can generate noisy local states:
- files modified but not checked out,
- pending edits scattered across Content folders,
- UE “hijacks” (auto-checkouts / unchanged edits) cluttering changelists.

P4CHIMARI provides a fast, guided workflow to get from “workspace messy” to “ready to submit” safely.

## Who it’s for

- Unreal developers/tech artists using Perforce on Windows
- Anyone who wants a structured, low-risk way to reconcile and revert

## Goals

- Reduce time spent manually running `p4 status/edit/reconcile/revert`
- Make dangerous operations explicit (double confirmation)
- Make UE hijacked-file cleanup quick and repeatable

## Success metrics

- Time to clean a noisy workspace: minutes, not hours
- Fewer accidental reverts (guardrails + confirmation)
- Reduced changelist noise (only real diffs remain)

## Scope

- Local workspace scanning and Perforce CLI orchestration
- Folder selection + filtering + action execution
- Hijacked file detection / cleanup via `p4 revert -a` flow

## Non-goals

- Replacing P4V
- Cross-platform support (Windows-first)
- Managing branching/merging workflows

## Constraints / assumptions

- Requires `p4` CLI installed and in PATH
- Requires an active workspace/client configured (`P4PORT`, `P4USER`, etc.)
- Unreal hijack logic depends on Perforce’s view of unchanged files (`revert -a`)

## Quick start

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

## Roadmap

- Add “dry-run” mode for revert/reconcile previews
- Improve large-workspace performance (incremental scanning, caching)
- Optional: export reports for changelist reviews (CSV/Markdown)
