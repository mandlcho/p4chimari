package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Baseline stores the snapshot of files checked out when UE was opened
type Baseline struct {
	Timestamp time.Time `json:"timestamp"`
	Files     []string  `json:"files"`
}

// getBaselinePath returns the path to the baseline file
func getBaselinePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".p4chimari_baseline.json"), nil
}

// captureBaseline saves current opened files as baseline (hijacked files)
func captureBaseline() error {
	fmt.Println("\n📸 Capturing baseline (hijacked files)...")
	fmt.Println("─────────────────────────────────────")

	// Get currently opened files
	openedFiles, err := getPendingFiles()
	if err != nil {
		return fmt.Errorf("failed to get opened files: %v", err)
	}

	baseline := Baseline{
		Timestamp: time.Now(),
		Files:     openedFiles,
	}

	// Save to file
	baselinePath, err := getBaselinePath()
	if err != nil {
		return fmt.Errorf("failed to get baseline path: %v", err)
	}

	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %v", err)
	}

	err = os.WriteFile(baselinePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write baseline: %v", err)
	}

	fmt.Printf("✓ Captured %d hijacked file(s)\n", len(openedFiles))
	fmt.Printf("  Timestamp: %s\n", baseline.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Saved to: %s\n", baselinePath)

	if len(openedFiles) > 0 {
		fmt.Println("\nHijacked files:")
		for i, file := range openedFiles {
			if i < 10 {
				fmt.Printf("  • %s\n", file)
			}
		}
		if len(openedFiles) > 10 {
			fmt.Printf("  ... and %d more\n", len(openedFiles)-10)
		}
	}

	return nil
}

// loadBaseline loads the saved baseline
func loadBaseline() (*Baseline, error) {
	baselinePath, err := getBaselinePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(baselinePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no baseline found - run 'Capture Baseline' first")
		}
		return nil, err
	}

	var baseline Baseline
	err = json.Unmarshal(data, &baseline)
	if err != nil {
		return nil, fmt.Errorf("failed to parse baseline: %v", err)
	}

	return &baseline, nil
}

// clearBaseline removes the saved baseline
func clearBaseline() error {
	baselinePath, err := getBaselinePath()
	if err != nil {
		return err
	}

	err = os.Remove(baselinePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	fmt.Println("✓ Baseline cleared")
	return nil
}

// getUnchangedFiles returns files that are opened but have no actual changes
// These are likely hijacked files
func getUnchangedFiles() ([]string, error) {
	cmd := exec.Command("p4", "diff", "-sr")
	output, _ := cmd.Output()

	// p4 diff -sr returns opened files with no changes
	// It may return exit code 1 if there are differences, which is expected
	outputStr := string(output)

	if len(outputStr) == 0 {
		return []string{}, nil
	}

	var unchangedFiles []string
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		unchangedFiles = append(unchangedFiles, line)
	}

	return unchangedFiles, nil
}

// getRealChanges returns files that have actual changes (not hijacked)
func getRealChanges() ([]string, []string, error) {
	// Get all opened files
	openedFiles, err := getPendingFiles()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get opened files: %v", err)
	}

	// Get unchanged files (hijacked)
	unchangedFiles, err := getUnchangedFiles()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get unchanged files: %v", err)
	}

	// Convert to maps for easy lookup
	unchangedMap := make(map[string]bool)
	for _, file := range unchangedFiles {
		unchangedMap[file] = true
	}

	// Separate real changes from hijacked
	var realChanges []string
	var hijacked []string

	for _, file := range openedFiles {
		if unchangedMap[file] {
			hijacked = append(hijacked, file)
		} else {
			realChanges = append(realChanges, file)
		}
	}

	return realChanges, hijacked, nil
}

// revertHijackedFiles reverts files that were hijacked (unchanged)
func revertHijackedFiles() error {
	fmt.Println("\n🔄 Finding hijacked files (opened but unchanged)...")
	fmt.Println("─────────────────────────────────────")

	hijackedFiles, err := getUnchangedFiles()
	if err != nil {
		return fmt.Errorf("failed to find hijacked files: %v", err)
	}

	if len(hijackedFiles) == 0 {
		fmt.Println("✓ No hijacked files found - all opened files have changes!")
		return nil
	}

	fmt.Printf("Found %d hijacked file(s):\n", len(hijackedFiles))
	for i, file := range hijackedFiles {
		if i < 20 {
			fmt.Printf("  • %s\n", file)
		}
	}
	if len(hijackedFiles) > 20 {
		fmt.Printf("  ... and %d more\n", len(hijackedFiles)-20)
	}

	fmt.Printf("\n⚠️  This will revert %d unchanged file(s)\n", len(hijackedFiles))
	fmt.Print("Proceed? (yes/no): ")

	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "yes" && response != "y" {
		fmt.Println("Cancelled.")
		return nil
	}

	fmt.Println("\nReverting hijacked files...")

	// Use p4 revert -a to revert unchanged files
	cmd := exec.Command("p4", "revert", "-a")
	output, err := cmd.CombinedOutput()

	if err != nil && len(output) == 0 {
		return fmt.Errorf("failed to revert files: %v", err)
	}

	outputStr := string(output)
	fmt.Println(outputStr)

	fmt.Println("\n✓ Done! Hijacked files have been reverted.")
	fmt.Println("  Your real changes remain checked out.")

	return nil
}

// showHijackedStatus shows comparison of hijacked vs real changes
func showHijackedStatus() error {
	fmt.Println("\n📊 Hijacked Files Analysis")
	fmt.Println("─────────────────────────────────────")

	realChanges, hijacked, err := getRealChanges()
	if err != nil {
		return err
	}

	total := len(realChanges) + len(hijacked)

	fmt.Printf("Total opened files:     %d\n", total)
	fmt.Printf("  Real changes:         %d (%.0f%%)\n", len(realChanges), float64(len(realChanges))/float64(total)*100)
	fmt.Printf("  Hijacked (unchanged): %d (%.0f%%)\n", len(hijacked), float64(len(hijacked))/float64(total)*100)
	fmt.Println("─────────────────────────────────────")

	if len(realChanges) > 0 {
		fmt.Println("\n✓ Real Changes:")
		for i, file := range realChanges {
			if i < 10 {
				fmt.Printf("  • %s\n", file)
			}
		}
		if len(realChanges) > 10 {
			fmt.Printf("  ... and %d more\n", len(realChanges)-10)
		}
	}

	if len(hijacked) > 0 {
		fmt.Println("\n⚠️  Hijacked Files (unchanged):")
		for i, file := range hijacked {
			if i < 10 {
				fmt.Printf("  • %s\n", file)
			}
		}
		if len(hijacked) > 10 {
			fmt.Printf("  ... and %d more\n", len(hijacked)-10)
		}
	}

	return nil
}

// showHijackedMenu displays the quick hijacked file management menu
func showHijackedMenu(reader *bufio.Reader) {
	for {
		fmt.Println("\n🎯 HIJACKED FILE MANAGEMENT")
		fmt.Println("─────────────────────────────────────")
		fmt.Println("What would you like to do?")
		fmt.Println("  1. Show hijacked files status")
		fmt.Println("  2. Revert hijacked files (auto-cleanup)")
		fmt.Println("  3. Back to main menu")
		fmt.Print("\nEnter choice (1-3): ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			err := showHijackedStatus()
			if err != nil {
				fmt.Printf("\nError: %v\n", err)
			}
			fmt.Print("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "2":
			err := revertHijackedFiles()
			if err != nil {
				fmt.Printf("\nError: %v\n", err)
			}
			fmt.Print("\nPress Enter to continue...")
			reader.ReadString('\n')
		case "3":
			return
		default:
			fmt.Println("Invalid choice.")
		}
	}
}
