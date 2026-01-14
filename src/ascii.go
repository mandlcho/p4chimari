package main

import "fmt"

func printPachimari() {
	fmt.Println(`
           _______________
          /               \
         /    _________    \
        |    /         \    |
        |   |  O     O  |   |
        |   |           |   |
         \  |     v     |  /
          \ |   \___/   | /
           \|           |/
         ___\_         _/___
        /    \_________/    \
       /  __/           \__  \
      /  /                 \  \
     /__/                   \__\

          P 4 C H I M A R I
`)
}

func printHeader() {
	printPachimari()
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println("                    p4chimari - Perforce Workspace Helper")
	fmt.Println("═══════════════════════════════════════════════════════════════════════════════")
	fmt.Println()
}
