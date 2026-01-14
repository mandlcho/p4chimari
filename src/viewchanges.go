package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ChangelistCategory struct {
	Name  string
	Count int
	Files []string
}

func showViewChanges(p4Info *P4Info) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Clear screen and show header
		clearScreen()
		printHeader()

		// Get data
		changelists := getChangelists()
		unsavedAssets := getUnsavedAssets()
		uncontrolled := getUncontrolledFiles()

		// Calculate panel dimensions
		width := 80
		leftWidth := 35
		rightWidth := width - leftWidth - 3

		// Section header
		fmt.Printf("VIEW CHANGES - %s\n", p4Info.ClientName)
		fmt.Println("─────────────────────────────────────")
		fmt.Println()

		// Menu options
		categories := []ChangelistCategory{
			{Name: "Default Changelist", Count: len(changelists["default"]), Files: changelists["default"]},
			{Name: "Unsaved Assets", Count: len(unsavedAssets), Files: unsavedAssets},
			{Name: "Uncontrolled Changelists", Count: len(uncontrolled), Files: uncontrolled},
		}

		// Add numbered changelists
		for clNum, files := range changelists {
			if clNum != "default" {
				categories = append(categories, ChangelistCategory{
					Name:  fmt.Sprintf("Changelist %s", clNum),
					Count: len(files),
					Files: files,
				})
			}
		}

		// Display left and right panels
		fmt.Printf("┌─%-*s─┬─%-*s─┐\n", leftWidth, strings.Repeat("─", leftWidth), rightWidth, strings.Repeat("─", rightWidth))
		fmt.Printf("│ %-*s │ %-*s │\n", leftWidth, "CATEGORIES", rightWidth, "FILES")
		fmt.Printf("├─%-*s─┼─%-*s─┤\n", leftWidth, strings.Repeat("─", leftWidth), rightWidth, strings.Repeat("─", rightWidth))

		// Print categories and placeholder for files
		maxLines := 20
		for i := 0; i < maxLines; i++ {
			leftContent := ""
			if i < len(categories) {
				cat := categories[i]
				leftContent = fmt.Sprintf("%d. %s (%d)", i+1, cat.Name, cat.Count)
			}

			rightContent := ""

			fmt.Printf("│ %-*s │ %-*s │\n", leftWidth, truncate(leftContent, leftWidth), rightWidth, truncate(rightContent, rightWidth))
		}

		fmt.Printf("└─%-*s─┴─%-*s─┘\n", leftWidth, strings.Repeat("─", leftWidth), rightWidth, strings.Repeat("─", rightWidth))

		// Commands
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  [1-9] - Select category to view files")
		fmt.Println("  [r]   - Refresh")
		fmt.Println("  [c]   - Checkout unsaved assets")
		fmt.Println("  [o]   - Reconcile uncontrolled")
		fmt.Println("  [q]   - Quit")
		fmt.Print("\nEnter command: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "q" {
			return
		} else if input == "r" {
			continue
		} else if input == "c" {
			if len(unsavedAssets) > 0 {
				checkoutFilesList(unsavedAssets)
				fmt.Print("\nPress Enter to continue...")
				reader.ReadString('\n')
			}
		} else if input == "o" {
			reconcileFiles()
			fmt.Print("\nPress Enter to continue...")
			reader.ReadString('\n')
		} else if input >= "1" && input <= "9" {
			idx := int(input[0] - '1')
			if idx < len(categories) {
				showCategoryFiles(categories[idx], reader)
			}
		}
	}
}

func showCategoryFiles(category ChangelistCategory, reader *bufio.Reader) {
	// Clear screen and show header
	clearScreen()
	printHeader()

	fmt.Printf("%s - %d file(s)\n", category.Name, category.Count)
	fmt.Println("─────────────────────────────────────")
	fmt.Println()

	if len(category.Files) == 0 {
		fmt.Println("  No files in this category.")
	} else {
		for i, file := range category.Files {
			fmt.Printf("  %d. %s\n", i+1, file)
		}
	}

	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  [b] - Back to main view")
	fmt.Println("  [c] - Checkout selected files (unsaved assets only)")
	fmt.Print("\nEnter command: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "c" && category.Name == "Unsaved Assets" {
		checkoutFilesList(category.Files)
		fmt.Print("\nPress Enter to continue...")
		reader.ReadString('\n')
	}
}

func getChangelists() map[string][]string {
	result := make(map[string][]string)

	cmd := exec.Command("p4", "opened", "-C")
	output, err := cmd.Output()
	if err != nil {
		return result
	}

	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse: //depot/path#1 - edit change 12345 (text)
		if strings.Contains(line, " - ") {
			parts := strings.Fields(line)
			changeNum := "default"

			for i, part := range parts {
				if part == "change" && i+1 < len(parts) {
					changeNum = parts[i+1]
					break
				}
			}

			if strings.Contains(line, "default change") {
				changeNum = "default"
			}

			filePath := strings.Split(line, "#")[0]
			result[changeNum] = append(result[changeNum], filePath)
		}
	}

	return result
}

func getUnsavedAssets() []string {
	// Don't scan - this is too slow for the view changes screen
	// Users can use option 2 to scan for modified files
	return []string{}
}

func getUncontrolledFiles() []string {
	// For now, return empty - this would need custom logic
	// based on what "uncontrolled" means in your workflow
	return []string{}
}

func checkoutFilesList(files []string) {
	fmt.Println("\nChecking out files...")
	for _, file := range files {
		fmt.Printf("  p4 edit %s\n", file)
		cmd := exec.Command("p4", "edit", file)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("    Error: %v\n", err)
		} else {
			fmt.Printf("    %s\n", strings.TrimSpace(string(output)))
		}
	}
	fmt.Println("\nDone!")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
