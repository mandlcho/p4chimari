package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func showFolderPicker(p4Info *P4Info, config *Config) ([]string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Clear screen
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()

		fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
		fmt.Println("                              SELECT FOLDERS TO SCAN")
		fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
		fmt.Println()

		// Show recent folders first
		recentFolders := config.GetRecentFolders()
		if len(recentFolders) > 0 {
			fmt.Println("📌 Recent Folders:")
			for i, folder := range recentFolders {
				if i < 5 {
					fmt.Printf("  %d. %s (used %d times)\n", i+1, folder.Path, folder.UseCount)
				}
			}
			fmt.Println()
		}

		fmt.Println("Options:")
		fmt.Println("  [b] Browse Content subfolders")
		fmt.Println("  [r] Use recent folder")
		fmt.Println("  [a] Scan all of Content")
		fmt.Println("  [c] Custom path")
		fmt.Println("  [q] Cancel")
		fmt.Print("\nEnter choice: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "b":
			return browseFolders(p4Info, config, reader)
		case "r":
			if len(recentFolders) > 0 {
				return selectRecentFolder(recentFolders, reader)
			} else {
				fmt.Println("No recent folders found.")
				fmt.Print("Press Enter to continue...")
				reader.ReadString('\n')
			}
		case "a":
			contentPath := filepath.Join(p4Info.ClientRoot, "Project", "Content")
			config.AddRecentFolder(contentPath)
			config.Save()
			return []string{contentPath}, nil
		case "c":
			return enterCustomPath(p4Info, config, reader)
		case "q":
			return nil, fmt.Errorf("cancelled")
		}
	}
}

func browseFolders(p4Info *P4Info, config *Config, reader *bufio.Reader) ([]string, error) {
	contentPath := filepath.Join(p4Info.ClientRoot, "Project", "Content")

	// Get subdirectories
	entries, err := os.ReadDir(contentPath)
	if err != nil {
		return nil, err
	}

	var folders []string
	for _, entry := range entries {
		if entry.IsDir() {
			folders = append(folders, entry.Name())
		}
	}

	if len(folders) == 0 {
		fmt.Println("No subfolders found in Content.")
		fmt.Print("Press Enter to continue...")
		reader.ReadString('\n')
		return nil, fmt.Errorf("no folders")
	}

	selected := make(map[int]bool)

	for {
		// Clear screen
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()

		fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
		fmt.Println("                         SELECT FOLDERS (Multi-select)")
		fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
		fmt.Println()

		for i, folder := range folders {
			checkbox := "[ ]"
			if selected[i] {
				checkbox = "[✓]"
			}
			fmt.Printf("  %d. %s %s\n", i+1, checkbox, folder)
		}

		fmt.Println()
		fmt.Printf("Selected: %d folder(s)\n", len(selected))
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  [number]     - Toggle selection")
		fmt.Println("  [a]          - Select all")
		fmt.Println("  [n]          - Clear selection")
		fmt.Println("  [done]       - Confirm selection")
		fmt.Println("  [cancel]     - Go back")
		fmt.Print("\nEnter command: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "done" {
			if len(selected) == 0 {
				fmt.Println("No folders selected!")
				fmt.Print("Press Enter to continue...")
				reader.ReadString('\n')
				continue
			}

			var selectedPaths []string
			for idx := range selected {
				folderPath := filepath.Join(contentPath, folders[idx])
				selectedPaths = append(selectedPaths, folderPath)
				config.AddRecentFolder(folderPath)
			}
			config.Save()
			return selectedPaths, nil
		} else if input == "cancel" {
			return nil, fmt.Errorf("cancelled")
		} else if input == "a" {
			for i := range folders {
				selected[i] = true
			}
		} else if input == "n" {
			selected = make(map[int]bool)
		} else {
			// Try to parse as number
			var idx int
			_, err := fmt.Sscanf(input, "%d", &idx)
			if err == nil && idx > 0 && idx <= len(folders) {
				idx-- // Convert to 0-based
				selected[idx] = !selected[idx]
			}
		}
	}
}

func selectRecentFolder(recentFolders []RecentFolder, reader *bufio.Reader) ([]string, error) {
	fmt.Println("\nSelect a recent folder:")
	for i, folder := range recentFolders {
		if i < 5 {
			fmt.Printf("  %d. %s\n", i+1, folder.Path)
		}
	}
	fmt.Print("\nEnter number: ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	var idx int
	_, err := fmt.Sscanf(input, "%d", &idx)
	if err != nil || idx < 1 || idx > len(recentFolders) {
		return nil, fmt.Errorf("invalid selection")
	}

	return []string{recentFolders[idx-1].Path}, nil
}

func enterCustomPath(p4Info *P4Info, config *Config, reader *bufio.Reader) ([]string, error) {
	fmt.Print("\nEnter custom path (relative to Content): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return nil, fmt.Errorf("cancelled")
	}

	contentPath := filepath.Join(p4Info.ClientRoot, "Project", "Content")
	customPath := filepath.Join(contentPath, input)

	// Check if path exists
	if _, err := os.Stat(customPath); os.IsNotExist(err) {
		fmt.Printf("Path does not exist: %s\n", customPath)
		fmt.Print("Press Enter to continue...")
		reader.ReadString('\n')
		return nil, fmt.Errorf("path not found")
	}

	config.AddRecentFolder(customPath)
	config.Save()
	return []string{customPath}, nil
}
