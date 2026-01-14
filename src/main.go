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

type DirtyFile struct {
	Path   string
	Action string
}

type P4Info struct {
	UserName     string
	ClientName   string
	ClientHost   string
	ClientRoot   string
	ServerAddr   string
	ServerUptime string
	CurrentDir   string
}

func main() {
	defer func() {
		fmt.Print("\nPress Enter to exit...")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}()

	// Clear screen and show header
	clearScreen()
	printHeader()

	// Check if p4 is available
	if !isP4Available() {
		fmt.Println("❌ Status: NOT CONNECTED")
		fmt.Println("Error: p4 command not found. Please ensure Perforce CLI is installed and in PATH.")
		return
	}

	// Get detailed P4 connection info
	p4Info, err := getP4Info()
	if err != nil {
		fmt.Println("❌ Status: NOT CONNECTED")
		fmt.Printf("Error: Unable to connect to P4.\n%v\n", err)
		return
	}

	// Check if current directory is under workspace root
	if !strings.HasPrefix(strings.ToLower(p4Info.CurrentDir), strings.ToLower(p4Info.ClientRoot)) {
		fmt.Println("⚠ WARNING: Not in workspace directory!")
		fmt.Printf("  Current Dir: %s\n", p4Info.CurrentDir)
		fmt.Printf("  Workspace Root: %s\n", p4Info.ClientRoot)
		fmt.Println("\nChanging to workspace root...")
		err = os.Chdir(p4Info.ClientRoot)
		if err != nil {
			fmt.Printf("Error: Could not change to workspace root: %v\n", err)
			return
		}
		p4Info.CurrentDir, _ = os.Getwd()
	}

	// Display connection status
	fmt.Println("✓ Status: CONNECTED")
	fmt.Println("\nConnection Details:")
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("  User:            %s\n", p4Info.UserName)
	fmt.Printf("  Workspace:       %s\n", p4Info.ClientName)
	fmt.Printf("  Host:            %s\n", p4Info.ClientHost)
	fmt.Printf("  Root:            %s\n", p4Info.ClientRoot)
	fmt.Printf("  Server:          %s\n", p4Info.ServerAddr)
	if p4Info.ServerUptime != "" {
		fmt.Printf("  Server Uptime:   %s\n", p4Info.ServerUptime)
	}
	fmt.Printf("  Current Dir:     %s\n", p4Info.CurrentDir)
	fmt.Println("─────────────────────────────────────")

	// Load config
	config, _ := loadConfig()

	// Single-level main menu
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n─────────────────────────────────────")
		fmt.Println("MAIN MENU")
		fmt.Println("─────────────────────────────────────")
		fmt.Println("  1. View Changes (UE-style view)")
		fmt.Println("  2. Scan & show modified files (choose folders)")
		fmt.Println("  3. Reconcile all files in Project folder")
		fmt.Println("  4. 🎯 Show hijacked files - See which opened files have NO changes")
		fmt.Println("  5. 🧹 Auto-revert unchanged files - Clean up hijacked files")
		fmt.Println("  6. Exit")
		fmt.Print("\nEnter choice (1-6): ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			showViewChanges(p4Info)
			fmt.Print("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "2":
			// Show folder picker
			fmt.Println("\nSelect folder(s) to scan:")
			selectedFolders, err := showFolderPicker(p4Info, config)
			if err != nil {
				fmt.Printf("Cancelled: %v\n", err)
				continue
			}

			// Scan and show results
			scanAndShowFiles(selectedFolders, reader)
		case "3":
			projectPath := filepath.Join(p4Info.ClientRoot, "Project")
			reconcileFilesInFolders([]string{projectPath})
			fmt.Print("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "4":
			err := showHijackedStatus()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			fmt.Print("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "5":
			err := revertHijackedFiles()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			fmt.Print("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "6":
			fmt.Println("Exiting.")
			return
		default:
			fmt.Println("Invalid choice.")
		}
	}
}

func scanAndShowFiles(selectedFolders []string, reader *bufio.Reader) {
	// Scan workspace for changes
	fmt.Println("\nScanning workspace for changes...")
	fmt.Println("─────────────────────────────────────")

	fmt.Printf("→ Checking pending changes (p4 opened)...\n")
	pendingFiles, _ := getPendingFiles()
	fmt.Printf("  ✓ Found %d file(s) already checked out\n", len(pendingFiles))

	fmt.Printf("→ Scanning for modified files in selected folders...\n")
	for _, folder := range selectedFolders {
		fmt.Printf("  - %s\n", folder)
	}
	dirtyFiles, err := findDirtyFilesInFolders(selectedFolders, true)
	if err != nil {
		fmt.Printf("  ✗ Error: %v\n", err)
	} else {
		fmt.Printf("  ✓ Found %d file(s) modified but not checked out\n", len(dirtyFiles))
	}
	fmt.Println("─────────────────────────────────────")

	// Show results
	fmt.Println("\n📋 Pending Changes (Already Checked Out):")
	if len(pendingFiles) == 0 {
		fmt.Println("  ✓ None")
	} else {
		fmt.Printf("  %d file(s) checked out\n", len(pendingFiles))
		for i, file := range pendingFiles {
			if i < 10 {
				fmt.Printf("    • %s\n", file)
			}
		}
		if len(pendingFiles) > 10 {
			fmt.Printf("    ... and %d more\n", len(pendingFiles)-10)
		}
	}

	fmt.Println("\n⚠ Files Modified But Not Checked Out:")
	if len(dirtyFiles) == 0 {
		fmt.Println("  ✓ None - All changes are tracked!")
	} else {
		fmt.Printf("  %d file(s) need attention\n", len(dirtyFiles))
		for i, file := range dirtyFiles {
			if i < 10 {
				fmt.Printf("    • [%s] %s\n", file.Action, file.Path)
			}
		}
		if len(dirtyFiles) > 10 {
			fmt.Printf("    ... and %d more\n", len(dirtyFiles)-10)
		}
	}

	// Actions menu
	fmt.Println("\n─────────────────────────────────────")
	fmt.Println("Actions:")
	fmt.Println("  1. Filter by action (add/edit/delete)")
	fmt.Println("  2. Checkout selected files")
	fmt.Println("  3. Reconcile all in these folders")
	fmt.Println("  4. Revert files")
	fmt.Println("  5. Back to main menu")
	fmt.Print("\nEnter choice (1-5): ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		filterByAction(dirtyFiles, reader)
	case "2":
		if len(dirtyFiles) > 0 {
			selectAndCheckoutFiles(dirtyFiles)
		} else {
			fmt.Println("No files to checkout.")
		}
	case "3":
		reconcileFilesInFolders(selectedFolders)
	case "4":
		if len(dirtyFiles) > 0 {
			revertFiles(dirtyFiles, reader)
		} else {
			fmt.Println("No files to revert.")
		}
	case "5":
		return
	default:
		fmt.Println("Invalid choice.")
	}

	fmt.Print("\nPress Enter to continue...")
	reader.ReadString('\n')
}

func isP4Available() bool {
	cmd := exec.Command("p4", "info")
	err := cmd.Run()
	return err == nil
}

func getP4Info() (*P4Info, error) {
	cmd := exec.Command("p4", "info")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	info := &P4Info{}
	cwd, _ := os.Getwd()
	info.CurrentDir = cwd

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "User name:") {
			info.UserName = extractValue(line)
		} else if strings.HasPrefix(line, "Client name:") {
			info.ClientName = extractValue(line)
		} else if strings.HasPrefix(line, "Client host:") {
			info.ClientHost = extractValue(line)
		} else if strings.HasPrefix(line, "Client root:") {
			info.ClientRoot = extractValue(line)
		} else if strings.HasPrefix(line, "Server address:") {
			info.ServerAddr = extractValue(line)
		} else if strings.HasPrefix(line, "Server uptime:") {
			info.ServerUptime = extractValue(line)
		}
	}

	if info.ClientName == "" {
		return nil, fmt.Errorf("could not determine client name")
	}

	return info, nil
}

func extractValue(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func getPendingFiles() ([]string, error) {
	cmd := exec.Command("p4", "opened")
	output, err := cmd.Output()

	// If no files are opened, p4 returns an error
	if err != nil {
		if len(output) == 0 {
			return []string{}, nil
		}
		return nil, err
	}

	var files []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Parse lines like: "//depot/path/file.txt#1 - edit default change (text)"
		parts := strings.Split(line, "#")
		if len(parts) >= 1 {
			files = append(files, strings.TrimSpace(parts[0]))
		}
	}

	return files, nil
}

func findDirtyFiles() ([]DirtyFile, error) {
	p4Info, _ := getP4Info()
	contentPath := filepath.Join(p4Info.ClientRoot, "Project", "Content")
	return findDirtyFilesInFolders([]string{contentPath}, false)
}

func findDirtyFilesInFolders(folders []string, verbose bool) ([]DirtyFile, error) {
	var allDirtyFiles []DirtyFile

	for _, folder := range folders {
		files, err := scanFolder(folder, verbose)
		if err != nil {
			return nil, err
		}
		allDirtyFiles = append(allDirtyFiles, files...)
	}

	return allDirtyFiles, nil
}

func scanFolder(folder string, verbose bool) ([]DirtyFile, error) {
	// Run p4 reconcile in preview mode
	if verbose {
		fmt.Printf("  Executing: p4 reconcile -n %s/...\n", folder)
	}

	cmd := exec.Command("p4", "reconcile", "-n", filepath.Join(folder, "..."))
	cmd.Dir = folder

	// Create a channel to track progress
	done := make(chan bool)
	if verbose {
		go func() {
			spinner := []string{"|", "/", "-", "\\"}
			i := 0
			startTime := time.Now()
			for {
				select {
				case <-done:
					fmt.Printf("\r  ✓ Complete (took %.1f seconds)\n", time.Since(startTime).Seconds())
					return
				default:
					elapsed := time.Since(startTime).Seconds()
					fmt.Printf("\r  %s Scanning... (%.0fs) [Ctrl+C to cancel]", spinner[i%len(spinner)], elapsed)
					i++
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}

	output, _ := cmd.CombinedOutput()

	if verbose {
		done <- true
	}

	// Note: p4 reconcile -n can return various exit codes
	outputStr := string(output)

	if verbose {
		fmt.Printf("  Processing results...\n")
	}

	var dirtyFiles []DirtyFile
	fileCount := 0

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse different p4 reconcile -n output formats:
		// "//depot/path/file.txt#1 - opened for edit"
		// "//depot/path/file.txt#1 - reconcile to edit"
		// "//depot/path/file.txt - reconcile to add"
		if strings.Contains(line, "//") && (strings.Contains(line, " - opened for ") ||
		   strings.Contains(line, " - reconcile to ") || strings.Contains(line, "- currently opened for")) {

			// Extract depot path
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
				// If conversion fails, use depot path
				localPath = depotPath
			}

			// Determine action
			action := "edit"
			if strings.Contains(line, "add") {
				action = "add"
			} else if strings.Contains(line, "delete") {
				action = "delete"
			}

			dirtyFiles = append(dirtyFiles, DirtyFile{
				Path:   localPath,
				Action: action,
			})

			fileCount++
			if verbose && fileCount%10 == 0 {
				fmt.Printf("  ... found %d files so far\n", fileCount)
			}
		}
	}

	return dirtyFiles, nil
}

func depotToLocalPath(depotPath string) (string, error) {
	cmd := exec.Command("p4", "where", depotPath)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// p4 where output: "depot-path client-path local-path"
	parts := strings.Fields(string(output))
	if len(parts) >= 3 {
		return filepath.Clean(parts[2]), nil
	}

	return "", fmt.Errorf("could not convert depot path to local path")
}

func checkoutFiles(files []DirtyFile) {
	fmt.Println("\nChecking out files...")

	for _, file := range files {
		fmt.Printf("  p4 edit %s\n", file.Path)
		cmd := exec.Command("p4", "edit", file.Path)
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("    Error: %v\n", err)
			if len(output) > 0 {
				fmt.Printf("    %s\n", string(output))
			}
		} else {
			fmt.Printf("    %s\n", strings.TrimSpace(string(output)))
		}
	}

	fmt.Println("\nDone!")
}

func reconcileFiles() {
	fmt.Println("\nReconciling files...")
	fmt.Println("This will open files for add, edit, or delete to match your workspace (Project/Content folder only).")

	p4Info, err := getP4Info()
	if err != nil {
		fmt.Printf("Error getting workspace info: %v\n", err)
		return
	}

	// Only reconcile the Content folder
	contentPath := filepath.Join(p4Info.ClientRoot, "Project", "Content")

	cmd := exec.Command("p4", "reconcile", filepath.Join(contentPath, "..."))
	cmd.Dir = contentPath
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	if err != nil && len(outputStr) == 0 {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println(outputStr)

	if strings.Contains(outputStr, "opened for") {
		fmt.Println("\n✓ Files have been opened for change!")
		fmt.Println("Use 'p4 opened' to see all opened files.")
	} else if strings.Contains(outputStr, "no file(s) to reconcile") {
		fmt.Println("No changes to reconcile.")
	}

	fmt.Println("\nDone!")
}

func showReconcilePreview() {
	fmt.Println("\nReconcile Preview (what would happen in Project/Content folder):")
	fmt.Println("─────────────────────────────────────")

	p4Info, err := getP4Info()
	if err != nil {
		fmt.Printf("Error getting workspace info: %v\n", err)
		return
	}

	// Only reconcile the Content folder
	contentPath := filepath.Join(p4Info.ClientRoot, "Project", "Content")

	cmd := exec.Command("p4", "reconcile", "-n", filepath.Join(contentPath, "..."))
	cmd.Dir = contentPath
	output, err := cmd.CombinedOutput()

	outputStr := string(output)

	if len(outputStr) == 0 {
		fmt.Println("No files would be reconciled.")
		return
	}

	fmt.Println(outputStr)
	fmt.Println("─────────────────────────────────────")
	fmt.Println("To apply these changes, choose option 2 (Reconcile all files)")
}

func showAllFiles(files []DirtyFile) {
	fmt.Println("\nAll Modified Files:")
	fmt.Println("─────────────────────────────────────")
	for i, file := range files {
		fmt.Printf("  %d. [%s] %s\n", i+1, file.Action, file.Path)
	}
	fmt.Println("─────────────────────────────────────")
}

func selectAndCheckoutFiles(files []DirtyFile) {
	fmt.Println("\nSelect files to checkout (comma-separated numbers, or 'all'):")
	showAllFiles(files)

	fmt.Print("\nEnter selection: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("Cancelled.")
		return
	}

	var selectedFiles []DirtyFile

	if input == "all" {
		selectedFiles = files
	} else {
		selections := strings.Split(input, ",")
		for _, sel := range selections {
			sel = strings.TrimSpace(sel)
			idx := 0
			fmt.Sscanf(sel, "%d", &idx)
			if idx > 0 && idx <= len(files) {
				selectedFiles = append(selectedFiles, files[idx-1])
			}
		}
	}

	if len(selectedFiles) == 0 {
		fmt.Println("No valid files selected.")
		return
	}

	fmt.Printf("\nChecking out %d file(s)...\n", len(selectedFiles))
	checkoutFiles(selectedFiles)
}

func filterByAction(files []DirtyFile, reader *bufio.Reader) {
	fmt.Println("\n─────────────────────────────────────")
	fmt.Println("Filter by action type:")
	fmt.Println("  1. Show only Edits")
	fmt.Println("  2. Show only Adds")
	fmt.Println("  3. Show only Deletes")
	fmt.Println("  4. Show All")
	fmt.Print("\nEnter choice (1-4): ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var filtered []DirtyFile
	filterType := "All"

	switch input {
	case "1":
		filterType = "Edits"
		for _, file := range files {
			if file.Action == "edit" {
				filtered = append(filtered, file)
			}
		}
	case "2":
		filterType = "Adds"
		for _, file := range files {
			if file.Action == "add" {
				filtered = append(filtered, file)
			}
		}
	case "3":
		filterType = "Deletes"
		for _, file := range files {
			if file.Action == "delete" {
				filtered = append(filtered, file)
			}
		}
	case "4":
		filtered = files
	default:
		fmt.Println("Invalid choice.")
		return
	}

	fmt.Printf("\n%s: %d file(s)\n", filterType, len(filtered))
	showAllFiles(filtered)

	fmt.Println("\nOptions:")
	fmt.Println("  1. Checkout these files")
	fmt.Println("  2. Revert these files")
	fmt.Println("  3. Back")
	fmt.Print("\nEnter choice: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		selectAndCheckoutFiles(filtered)
	case "2":
		revertFiles(filtered, reader)
	case "3":
		return
	}
}

func revertFiles(files []DirtyFile, reader *bufio.Reader) {
	fmt.Println("\n⚠️  WARNING: REVERT FILES ⚠️")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println("This will PERMANENTLY DELETE your local changes and restore files from P4!")
	fmt.Println("This action CANNOT be undone!")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")

	showAllFiles(files)

	fmt.Println("\nSelect files to revert (comma-separated numbers, 'all', or 'cancel'):")
	fmt.Print("Enter selection: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" || input == "cancel" {
		fmt.Println("Cancelled.")
		return
	}

	var selectedFiles []DirtyFile

	if input == "all" {
		selectedFiles = files
	} else {
		selections := strings.Split(input, ",")
		for _, sel := range selections {
			sel = strings.TrimSpace(sel)
			idx := 0
			fmt.Sscanf(sel, "%d", &idx)
			if idx > 0 && idx <= len(files) {
				selectedFiles = append(selectedFiles, files[idx-1])
			}
		}
	}

	if len(selectedFiles) == 0 {
		fmt.Println("No valid files selected.")
		return
	}

	// Final confirmation
	fmt.Printf("\n⚠️  FINAL CONFIRMATION: Revert %d file(s)?\n", len(selectedFiles))
	fmt.Print("Type 'YES' to confirm: ")

	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm != "YES" {
		fmt.Println("Cancelled.")
		return
	}

	fmt.Println("\nReverting files...")
	for _, file := range selectedFiles {
		fmt.Printf("  p4 sync -f %s\n", file.Path)
		cmd := exec.Command("p4", "sync", "-f", file.Path)
		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("    Error: %v\n", err)
			if len(output) > 0 {
				fmt.Printf("    %s\n", string(output))
			}
		} else {
			fmt.Printf("    ✓ %s\n", strings.TrimSpace(string(output)))
		}
	}

	fmt.Println("\n✓ Done! Files have been reverted to P4 versions.")
}

func reconcileFilesInFolders(folders []string) {
	fmt.Println("\nReconciling files in selected folders...")
	fmt.Println("This will open files for add, edit, or delete to match your workspace.")

	for _, folder := range folders {
		fmt.Printf("\nReconciling: %s\n", folder)
		cmd := exec.Command("p4", "reconcile", filepath.Join(folder, "..."))
		cmd.Dir = folder
		output, err := cmd.CombinedOutput()

		outputStr := string(output)

		if err != nil && len(outputStr) == 0 {
			fmt.Printf("  Error: %v\n", err)
			continue
		}

		fmt.Println(outputStr)

		if strings.Contains(outputStr, "opened for") {
			fmt.Println("  ✓ Files have been opened for change!")
		} else if strings.Contains(outputStr, "no file(s) to reconcile") {
			fmt.Println("  No changes to reconcile.")
		}
	}

	fmt.Println("\n✓ Done!")
}
