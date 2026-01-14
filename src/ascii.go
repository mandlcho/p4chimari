package main

import (
	"fmt"
	"os/exec"
	"os"
)

func clearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func printHeader() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   ██████████████████                                                          ║")
	fmt.Println("║  ███████████████████       P4CHIMARI - Perforce Workspace Helper             ║")
	fmt.Println("║  ██████████████████        ────────────────────────────────────              ║")
	fmt.Println("║  ███████████████           Manage modified files and reconcile changes       ║")
	fmt.Println("║  █████████                 with your Perforce workspace                       ║")
	fmt.Println("║  ████████                                                                     ║")
	fmt.Println("║  ██████                    Version: 1.0                                       ║")
	fmt.Println("║  █████   ███    ███        Commands: checkout | reconcile | revert           ║")
	fmt.Println("║  █████    █      █                                                            ║")
	fmt.Println("║  ██████           ██                                                          ║")
	fmt.Println("║  ███████         ███                                                          ║")
	fmt.Println("║  ████████       ████                                                          ║")
	fmt.Println("║  ███████             ███                                                      ║")
	fmt.Println("║  ███████                                                                      ║")
	fmt.Println("║  █████████                                                                    ║")
	fmt.Println("║  ████████████████████                                                         ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func redrawWithHeader(content string) {
	clearScreen()
	printHeader()
	fmt.Print(content)
}
