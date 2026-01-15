package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ModifiedFile represents a file with actual modifications
type ModifiedFile struct {
	Path       string
	Action     string
	HasChanges bool
	IsOpened   bool
}

// ScanResult contains the results of scanning for modified files
type ScanResult struct {
	OpenedWithChanges    []ModifiedFile
	OpenedWithoutChanges []ModifiedFile
	NotOpenedButModified []ModifiedFile
	TotalScanned         int
	ScanDuration         time.Duration
}

// ============================================================================
// MAIN SCAN FUNCTION (All others call this)
// ============================================================================

// ScanForModifiedFilesScoped scans with custom selection and optional path scope
func ScanForModifiedFilesScoped(folders []string, scopePath string, verbose bool, scanOpenedWithChanges bool, scanOpenedWithoutChanges bool, scanNotOpened bool) (*ScanResult, error) {
	startTime := time.Now()
	result := &ScanResult{}

	if verbose {
		fmt.Println("\n🔍 Scanning workspace...")
		fmt.Println("─────────────────────────────────────")
	}

	// 1. Get files opened for edit with actual changes (p4 diff -se)
	if scanOpenedWithChanges {
		if verbose {
			fmt.Println("→ Finding opened files with changes...")
		}
		openedWithChanges, err := getOpenedFilesWithChangesScoped(scopePath, verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to get opened files with changes: %v", err)
		}
		result.OpenedWithChanges = openedWithChanges
		if verbose {
			fmt.Printf("  ✓ Found %d file(s) opened with changes\n", len(openedWithChanges))
		}
	} else {
		result.OpenedWithChanges = []ModifiedFile{}
	}

	// 2. Get files opened for edit without changes (p4 diff -sr)
	if scanOpenedWithoutChanges {
		if verbose {
			fmt.Println("→ Finding opened files without changes (hijacked)...")
		}
		openedWithoutChanges, err := getOpenedFilesWithoutChangesScoped(scopePath, verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to get hijacked files: %v", err)
		}
		result.OpenedWithoutChanges = openedWithoutChanges
		if verbose {
			fmt.Printf("  ✓ Found %d hijacked file(s)\n", len(openedWithoutChanges))
		}
	} else {
		result.OpenedWithoutChanges = []ModifiedFile{}
	}

	// 3. Get files modified but not opened (p4 reconcile -n)
	if scanNotOpened {
		if verbose {
			fmt.Println("→ Finding modified files not yet opened...")
		}
		notOpenedButModified, err := getModifiedFilesNotOpened(folders, verbose)
		if err != nil {
			return nil, fmt.Errorf("failed to get modified files: %v", err)
		}
		result.NotOpenedButModified = notOpenedButModified
		if verbose {
			fmt.Printf("  ✓ Found %d modified file(s) not opened\n", len(notOpenedButModified))
		}
	} else {
		result.NotOpenedButModified = []ModifiedFile{}
	}

	result.TotalScanned = len(result.OpenedWithChanges) + len(result.OpenedWithoutChanges) + len(result.NotOpenedButModified)
	result.ScanDuration = time.Since(startTime)

	if verbose {
		fmt.Println("─────────────────────────────────────")
		fmt.Printf("✓ Scan complete (took %.1fs)\n", result.ScanDuration.Seconds())
	}

	return result, nil
}

// ============================================================================
// WRAPPER FUNCTIONS (for backward compatibility)
// ============================================================================

// ScanForModifiedFiles scans for all types of modified files (full scan)
func ScanForModifiedFiles(folders []string, verbose bool) (*ScanResult, error) {
	return ScanForModifiedFilesScoped(folders, "", verbose, true, true, true)
}

// ============================================================================
// P4 COMMAND FUNCTIONS
// ============================================================================

// getOpenedFilesWithChangesScoped returns files that are opened for edit and have actual changes in a specific scope
func getOpenedFilesWithChangesScoped(scopePath string, verbose bool) ([]ModifiedFile, error) {
	var cmd *exec.Cmd
	if scopePath == "" {
		cmd = exec.Command("p4", "diff", "-se")
		if verbose {
			fmt.Println("  Executing: p4 diff -se")
		}
	} else {
		cmd = exec.Command("p4", "diff", "-se", scopePath)
		if verbose {
			fmt.Printf("  Executing: p4 diff -se %s\n", scopePath)
		}
	}

	// Show progress spinner
	done := make(chan bool)
	if verbose {
		go showSpinner(done, "  ")
	}

	output, _ := cmd.Output()
	outputStr := string(output)

	if verbose {
		done <- true
		time.Sleep(50 * time.Millisecond) // Let spinner clean up
	}

	if len(outputStr) == 0 {
		return []ModifiedFile{}, nil
	}

	var files []ModifiedFile
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		files = append(files, ModifiedFile{
			Path:       line,
			Action:     "edit",
			HasChanges: true,
			IsOpened:   true,
		})
	}

	return files, nil
}

// getOpenedFilesWithoutChangesScoped returns files that are opened but have no changes (hijacked) in a specific scope
func getOpenedFilesWithoutChangesScoped(scopePath string, verbose bool) ([]ModifiedFile, error) {
	var cmd *exec.Cmd
	if scopePath == "" {
		cmd = exec.Command("p4", "diff", "-sr")
		if verbose {
			fmt.Println("  Executing: p4 diff -sr")
		}
	} else {
		cmd = exec.Command("p4", "diff", "-sr", scopePath)
		if verbose {
			fmt.Printf("  Executing: p4 diff -sr %s\n", scopePath)
		}
	}

	// Show progress spinner
	done := make(chan bool)
	if verbose {
		go showSpinner(done, "  ")
	}

	output, _ := cmd.Output()
	outputStr := string(output)

	if verbose {
		done <- true
		time.Sleep(50 * time.Millisecond) // Let spinner clean up
	}

	if len(outputStr) == 0 {
		return []ModifiedFile{}, nil
	}

	var files []ModifiedFile
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		files = append(files, ModifiedFile{
			Path:       line,
			Action:     "edit",
			HasChanges: false,
			IsOpened:   true,
		})
	}

	return files, nil
}

// getModifiedFilesNotOpened returns files that have been modified but not opened for edit
func getModifiedFilesNotOpened(folders []string, verbose bool) ([]ModifiedFile, error) {
	if len(folders) == 0 {
		folders = []string{"."}
	}

	var allFiles []ModifiedFile

	for _, folder := range folders {
		if verbose {
			fmt.Printf("  Scanning folder: %s\n", folder)
		}

		files, err := scanFolderForUnopened(folder, verbose)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, files...)
	}

	return allFiles, nil
}

// scanFolderForUnopened scans a folder for modified but unopened files
func scanFolderForUnopened(folder string, verbose bool) ([]ModifiedFile, error) {
	cmd := exec.Command("p4", "reconcile", "-n", folder+"/...")

	if verbose {
		fmt.Printf("  Executing: p4 reconcile -n %s/...\n", folder)
	}

	// Show progress spinner
	done := make(chan bool)
	if verbose {
		go showSpinner(done, "  ")
	}

	output, _ := cmd.CombinedOutput()
	outputStr := string(output)

	if verbose {
		done <- true
		time.Sleep(50 * time.Millisecond) // Let spinner clean up
	}

	if len(outputStr) == 0 {
		return []ModifiedFile{}, nil
	}

	var files []ModifiedFile
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse p4 reconcile -n output
		if strings.Contains(line, "//") && strings.Contains(line, " - reconcile to ") {
			depotPath := ""
			if strings.Contains(line, "#") {
				parts := strings.Split(line, "#")
				depotPath = strings.TrimSpace(parts[0])
			} else {
				parts := strings.Split(line, " - ")
				if len(parts) >= 1 {
					depotPath = strings.TrimSpace(parts[0])
				}
			}

			if depotPath == "" {
				continue
			}

			// Convert depot path to local path
			localPath, err := depotToLocalPath(depotPath)
			if err != nil {
				localPath = depotPath
			}

			// Determine action
			action := "edit"
			if strings.Contains(line, "add") {
				action = "add"
			} else if strings.Contains(line, "delete") {
				action = "delete"
			}

			files = append(files, ModifiedFile{
				Path:       localPath,
				Action:     action,
				HasChanges: true,
				IsOpened:   false,
			})
		}
	}

	return files, nil
}

// ============================================================================
// DISPLAY FUNCTIONS
// ============================================================================

// PrintScanResults prints the scan results in a formatted way
func PrintScanResults(result *ScanResult) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         WORKSPACE MODIFICATION SCAN RESULTS                ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  Total files found:              %d\n", result.TotalScanned)
	fmt.Printf("  ✓ Opened with changes:          %d\n", len(result.OpenedWithChanges))
	fmt.Printf("  ⚠ Opened without changes:        %d (hijacked)\n", len(result.OpenedWithoutChanges))
	fmt.Printf("  📝 Modified but not opened:      %d\n", len(result.NotOpenedButModified))
	fmt.Printf("  ⏱ Scan duration:                 %.1fs\n", result.ScanDuration.Seconds())
	fmt.Println()

	if len(result.OpenedWithChanges) > 0 {
		fmt.Println("─────────────────────────────────────")
		fmt.Printf("✓ Files Opened with REAL Changes (%d):\n", len(result.OpenedWithChanges))
		for i, file := range result.OpenedWithChanges {
			if i < 15 {
				fmt.Printf("  [%s] %s\n", strings.ToUpper(file.Action), file.Path)
			}
		}
		if len(result.OpenedWithChanges) > 15 {
			fmt.Printf("  ... and %d more\n", len(result.OpenedWithChanges)-15)
		}
	}

	if len(result.OpenedWithoutChanges) > 0 {
		fmt.Println("\n─────────────────────────────────────")
		fmt.Printf("⚠ Hijacked Files (Opened but NO changes) (%d):\n", len(result.OpenedWithoutChanges))
		for i, file := range result.OpenedWithoutChanges {
			if i < 15 {
				fmt.Printf("  %s\n", file.Path)
			}
		}
		if len(result.OpenedWithoutChanges) > 15 {
			fmt.Printf("  ... and %d more\n", len(result.OpenedWithoutChanges)-15)
		}
	}

	if len(result.NotOpenedButModified) > 0 {
		fmt.Println("\n─────────────────────────────────────")
		fmt.Printf("📝 Modified Files NOT Opened (%d):\n", len(result.NotOpenedButModified))
		for i, file := range result.NotOpenedButModified {
			if i < 15 {
				fmt.Printf("  [%s] %s\n", strings.ToUpper(file.Action), file.Path)
			}
		}
		if len(result.NotOpenedButModified) > 15 {
			fmt.Printf("  ... and %d more\n", len(result.NotOpenedButModified)-15)
		}
	}

	if result.TotalScanned == 0 {
		fmt.Println("✨ All clean! No modifications found in your workspace.")
	}

	fmt.Println()
}

// ============================================================================
// MAIN UI FUNCTION
// ============================================================================

// ShowModifiedFilesAndRevert shows all modified files and allows selective force sync
func ShowModifiedFilesAndRevert(p4Info *P4Info, reader *bufio.Reader) {
	// First, ask for directory/scope to limit the scan
	fmt.Println("\n─────────────────────────────────────")
	fmt.Println("SELECT SCAN SCOPE")
	fmt.Println("─────────────────────────────────────")
	fmt.Println("⚠️  Large workspaces: Limit scope for faster scans!")
	fmt.Println()
	fmt.Println("  1. Entire workspace (all opened files)")
	fmt.Println("  2. Project folder only")
	fmt.Println("  3. Current directory only")
	fmt.Println("  4. Browse and select directory")
	fmt.Println("  5. Custom path (type manually)")
	fmt.Println("  6. Cancel")
	fmt.Print("\nEnter choice (1-6): ")

	scopeChoice, _ := reader.ReadString('\n')
	scopeChoice = strings.TrimSpace(scopeChoice)

	if scopeChoice == "6" {
		fmt.Println("Cancelled.")
		return
	}

	var scanPath string
	switch scopeChoice {
	case "1":
		scanPath = "" // Empty = all files
		fmt.Println("\n📂 Scope: Entire workspace")
	case "2":
		scanPath = filepath.Join(p4Info.ClientRoot, "Project") + "/..."
		fmt.Printf("\n📂 Scope: Project folder (%s)\n", scanPath)
	case "3":
		currentDir, _ := os.Getwd()
		scanPath = currentDir + "/..."
		fmt.Printf("\n📂 Scope: Current directory (%s)\n", scanPath)
	case "4":
		selectedDir := browseWorkspaceDirectories(p4Info, reader)
		if selectedDir == "" {
			fmt.Println("Cancelled.")
			return
		}
		scanPath = selectedDir + "/..."
		fmt.Printf("\n📂 Scope: %s\n", scanPath)
	case "5":
		fmt.Print("\nEnter directory path (relative to workspace root): ")
		customPath, _ := reader.ReadString('\n')
		customPath = strings.TrimSpace(customPath)
		if customPath == "" {
			fmt.Println("Cancelled.")
			return
		}
		fullPath := filepath.Join(p4Info.ClientRoot, customPath)
		scanPath = fullPath + "/..."
		fmt.Printf("\n📂 Scope: %s\n", scanPath)
	default:
		fmt.Println("Invalid choice.")
		return
	}

	// Ask user what to scan for
	fmt.Println("\n─────────────────────────────────────")
	fmt.Println("SELECT WHAT TO SCAN FOR")
	fmt.Println("─────────────────────────────────────")
	fmt.Println("Choose what to include in the scan:")
	fmt.Println("  1. Opened files with changes")
	fmt.Println("  2. Opened files without changes (hijacked)")
	fmt.Println("  3. Modified files not opened yet")
	fmt.Println()
	fmt.Println("Enter selection (comma-separated, e.g., '1,2' or '1,2,3'):")
	fmt.Println("  Quick:  1,2")
	fmt.Println("  Full:   1,2,3")
	fmt.Print("\nYour choice: ")

	scanChoice, _ := reader.ReadString('\n')
	scanChoice = strings.TrimSpace(scanChoice)

	if scanChoice == "" {
		fmt.Println("Cancelled.")
		return
	}

	// Parse selections
	selections := strings.Split(scanChoice, ",")
	scanOpenedWithChanges := false
	scanOpenedWithoutChanges := false
	scanNotOpened := false

	for _, sel := range selections {
		sel = strings.TrimSpace(sel)
		switch sel {
		case "1":
			scanOpenedWithChanges = true
		case "2":
			scanOpenedWithoutChanges = true
		case "3":
			scanNotOpened = true
		default:
			fmt.Printf("Invalid selection: %s\n", sel)
			return
		}
	}

	if !scanOpenedWithChanges && !scanOpenedWithoutChanges && !scanNotOpened {
		fmt.Println("No valid selections made.")
		return
	}

	// Ask user which directory to scan (only if scanning unopened files)
	var folders []string
	if scanNotOpened {
		fmt.Println("\n─────────────────────────────────────")
		fmt.Println("SELECT SCAN DIRECTORY")
		fmt.Println("─────────────────────────────────────")
		fmt.Println("  1. Entire workspace (all files)")
		fmt.Println("  2. Project folder only")
		fmt.Println("  3. Current directory")
		fmt.Println("  4. Browse workspace directories")
		fmt.Println("  5. Custom path (type manually)")
		fmt.Println("  6. Cancel")
		fmt.Print("\nEnter choice (1-6): ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			folders = []string{p4Info.ClientRoot}
			fmt.Printf("\n📂 Scanning: Entire workspace (%s)\n", p4Info.ClientRoot)
		case "2":
			projectPath := p4Info.ClientRoot + "/Project"
			folders = []string{projectPath}
			fmt.Printf("\n📂 Scanning: Project folder (%s)\n", projectPath)
		case "3":
			currentDir, _ := depotToLocalPath(".")
			if currentDir == "" {
				currentDir = "."
			}
			folders = []string{currentDir}
			fmt.Printf("\n📂 Scanning: Current directory (%s)\n", currentDir)
		case "4":
			selectedDir := browseWorkspaceDirectories(p4Info, reader)
			if selectedDir == "" {
				fmt.Println("Cancelled.")
				return
			}
			folders = []string{selectedDir}
			fmt.Printf("\n📂 Scanning: %s\n", selectedDir)
		case "5":
			fmt.Print("\nEnter directory path (relative to workspace root): ")
			customPath, _ := reader.ReadString('\n')
			customPath = strings.TrimSpace(customPath)
			if customPath == "" {
				fmt.Println("Cancelled.")
				return
			}
			fullPath := filepath.Join(p4Info.ClientRoot, customPath)
			folders = []string{fullPath}
			fmt.Printf("\n📂 Scanning: Custom directory (%s)\n", fullPath)
		case "6":
			fmt.Println("Cancelled.")
			return
		default:
			fmt.Println("Invalid choice.")
			return
		}
	} else {
		// Not scanning unopened files - no directory selection needed
		folders = []string{} // Empty means skip unopened files scan
	}

	// Show what we're scanning
	fmt.Println()
	if scanOpenedWithChanges && scanOpenedWithoutChanges && scanNotOpened {
		fmt.Println("🔍 Running FULL scan (all files)...")
	} else if scanOpenedWithChanges && scanOpenedWithoutChanges {
		fmt.Println("🚀 Running QUICK scan (opened files only)...")
	} else {
		fmt.Print("🔍 Scanning for: ")
		parts := []string{}
		if scanOpenedWithChanges {
			parts = append(parts, "opened with changes")
		}
		if scanOpenedWithoutChanges {
			parts = append(parts, "hijacked files")
		}
		if scanNotOpened {
			parts = append(parts, "unopened modified files")
		}
		fmt.Println(strings.Join(parts, ", ") + "...")
	}

	// Scan for modified files with path scope
	result, err := ScanForModifiedFilesScoped(folders, scanPath, true, scanOpenedWithChanges, scanOpenedWithoutChanges, scanNotOpened)
	if err != nil {
		fmt.Printf("Error scanning: %v\n", err)
		return
	}

	// Print results
	PrintScanResults(result)

	if result.TotalScanned == 0 {
		return
	}

	// Combine all modified files into a single list
	var allModifiedFiles []ModifiedFile
	allModifiedFiles = append(allModifiedFiles, result.OpenedWithChanges...)
	allModifiedFiles = append(allModifiedFiles, result.NotOpenedButModified...)

	if len(allModifiedFiles) == 0 {
		fmt.Println("No files with actual modifications to revert.")
		return
	}

	// Show selection menu
	fmt.Println("\n─────────────────────────────────────")
	fmt.Println("FORCE GET REVISION (p4 sync -f)")
	fmt.Println("─────────────────────────────────────")
	fmt.Println("⚠️  WARNING: This will DISCARD your local changes!")
	fmt.Println()
	fmt.Println("Select files to force sync (restore from P4):")
	fmt.Println()

	// Display all files with numbers
	for i, file := range allModifiedFiles {
		status := "MODIFIED"
		if file.IsOpened {
			status = "OPENED"
		}
		fmt.Printf("  %d. [%s] [%s] %s\n", i+1, status, strings.ToUpper(file.Action), file.Path)
	}

	fmt.Println()
	fmt.Println("Enter selection:")
	fmt.Println("  - Single: 1")
	fmt.Println("  - Multiple: 1,3,5")
	fmt.Println("  - Range: 1-5")
	fmt.Println("  - All: all")
	fmt.Println("  - Cancel: cancel or press Enter")
	fmt.Print("\nYour choice: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" || input == "cancel" {
		fmt.Println("Cancelled.")
		return
	}

	// Parse selection
	selectedFiles := parseFileSelection(input, allModifiedFiles)

	if len(selectedFiles) == 0 {
		fmt.Println("No valid files selected.")
		return
	}

	// Show what will be force synced
	fmt.Printf("\n⚠️  FINAL CONFIRMATION: Force sync %d file(s)?\n", len(selectedFiles))
	fmt.Println("This will DISCARD your local changes and restore from P4!")
	fmt.Println()
	fmt.Println("Files to be force synced:")
	for i, file := range selectedFiles {
		if i < 10 {
			fmt.Printf("  • %s\n", file.Path)
		}
	}
	if len(selectedFiles) > 10 {
		fmt.Printf("  ... and %d more\n", len(selectedFiles)-10)
	}
	fmt.Print("\nType 'YES' to confirm: ")

	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "YES" {
		fmt.Println("Cancelled.")
		return
	}

	// Force sync the selected files
	fmt.Println("\nForce syncing files...")
	for _, file := range selectedFiles {
		fmt.Printf("  p4 sync -f %s\n", file.Path)
		cmd := exec.Command("p4", "sync", "-f", file.Path)
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("    ✗ Error: %v\n", err)
			if len(output) > 0 {
				fmt.Printf("    %s\n", string(output))
			}
		} else {
			fmt.Printf("    ✓ %s\n", strings.TrimSpace(string(output)))
		}
	}

	fmt.Println("\n✓ Done! Files have been force synced from P4.")
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// showSpinner displays an animated spinner with elapsed time
func showSpinner(done chan bool, prefix string) {
	spinner := []string{"|", "/", "-", "\\"}
	i := 0
	startTime := time.Now()
	for {
		select {
		case <-done:
			elapsed := time.Since(startTime).Seconds()
			fmt.Printf("\r%s✓ Complete (%.1fs)       \n", prefix, elapsed)
			return
		default:
			elapsed := time.Since(startTime).Seconds()
			fmt.Printf("\r%s%s Scanning... (%.0fs)", prefix, spinner[i%len(spinner)], elapsed)
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// parseFileSelection parses user input and returns selected files
func parseFileSelection(input string, files []ModifiedFile) []ModifiedFile {
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "all" {
		return files
	}

	var selected []ModifiedFile
	parts := strings.Split(input, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check for range (e.g., "1-5")
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				var start, end int
				fmt.Sscanf(strings.TrimSpace(rangeParts[0]), "%d", &start)
				fmt.Sscanf(strings.TrimSpace(rangeParts[1]), "%d", &end)

				if start > 0 && end <= len(files) && start <= end {
					for i := start; i <= end; i++ {
						selected = append(selected, files[i-1])
					}
				}
			}
		} else {
			// Single number
			var num int
			fmt.Sscanf(part, "%d", &num)
			if num > 0 && num <= len(files) {
				selected = append(selected, files[num-1])
			}
		}
	}

	return selected
}

// browseWorkspaceDirectories allows interactive browsing of workspace directories
func browseWorkspaceDirectories(p4Info *P4Info, reader *bufio.Reader) string {
	currentPath := p4Info.ClientRoot

	for {
		// List subdirectories
		dirs, err := getSubdirectories(currentPath)
		if err != nil {
			fmt.Printf("Error reading directory: %v\n", err)
			return ""
		}

		// Display current location and directories
		fmt.Println("\n─────────────────────────────────────")
		relPath := strings.TrimPrefix(currentPath, p4Info.ClientRoot)
		if relPath == "" {
			relPath = "/"
		}
		fmt.Printf("Current: %s\n", relPath)
		fmt.Println("─────────────────────────────────────")

		if len(dirs) == 0 {
			fmt.Println("  (No subdirectories)")
		} else {
			for i, dir := range dirs {
				if i < 20 {
					fmt.Printf("  %d. %s\n", i+1, dir)
				}
			}
			if len(dirs) > 20 {
				fmt.Printf("  ... and %d more\n", len(dirs)-20)
			}
		}

		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  [1-20]  - Enter subdirectory")
		fmt.Println("  [s]     - Select this directory")
		fmt.Println("  [u]     - Go up one level")
		fmt.Println("  [c]     - Cancel")
		fmt.Print("\nEnter command: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "s" {
			return currentPath
		} else if input == "c" {
			return ""
		} else if input == "u" {
			// Go up one level
			if currentPath != p4Info.ClientRoot {
				currentPath = filepath.Dir(currentPath)
			}
		} else {
			// Try to parse as number
			var num int
			_, err := fmt.Sscanf(input, "%d", &num)
			if err == nil && num > 0 && num <= len(dirs) && num <= 20 {
				currentPath = filepath.Join(currentPath, dirs[num-1])
			} else {
				fmt.Println("Invalid choice.")
			}
		}
	}
}

// getSubdirectories returns a list of subdirectories in the given path
func getSubdirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Skip hidden directories and common non-code directories
			name := entry.Name()
			if !strings.HasPrefix(name, ".") && name != "node_modules" && name != "bin" && name != "obj" {
				dirs = append(dirs, name)
			}
		}
	}

	return dirs, nil
}
