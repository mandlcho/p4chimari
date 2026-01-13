package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	fmt.Println("p4chimari - Perforce Workspace Helper")
	fmt.Println("=====================================\n")

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

	// Show quick summary
	pendingFiles, _ := getPendingFiles()
	dirtyFiles, _ := findDirtyFiles()

	fmt.Println("\nQuick Summary:")
	fmt.Printf("  Pending Changes: %d file(s)\n", len(pendingFiles))
	fmt.Printf("  Unsaved Assets:  %d file(s)\n", len(dirtyFiles))

	// Main menu
	fmt.Println("\n─────────────────────────────────────")
	fmt.Println("What would you like to do?")
	fmt.Println("  1. View Changes (UE-style view)")
	fmt.Println("  2. Quick checkout all unsaved")
	fmt.Println("  3. Quick reconcile all")
	fmt.Println("  4. Exit")
	fmt.Print("\nEnter choice (1-4): ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		showViewChanges(p4Info)
	case "2":
		if len(dirtyFiles) > 0 {
			checkoutFiles(dirtyFiles)
		} else {
			fmt.Println("No unsaved files to checkout.")
		}
	case "3":
		reconcileFiles()
	case "4":
		fmt.Println("Exiting.")
	default:
		fmt.Println("Invalid choice.")
	}
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
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Run p4 reconcile in preview mode to find modified files
	cmd := exec.Command("p4", "reconcile", "-n", "...")
	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()

	// Note: p4 reconcile -n can return various exit codes
	outputStr := string(output)

	var dirtyFiles []DirtyFile

	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Debug: show raw output
		// fmt.Printf("DEBUG: %s\n", line)

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
	fmt.Println("This will open files for add, edit, or delete to match your workspace.")

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	cmd := exec.Command("p4", "reconcile", "...")
	cmd.Dir = cwd
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
	fmt.Println("\nReconcile Preview (what would happen):")
	fmt.Println("─────────────────────────────────────")

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}

	cmd := exec.Command("p4", "reconcile", "-n", "...")
	cmd.Dir = cwd
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
