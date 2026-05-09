package main

import (
	"escrowd/internal/bot"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "bot" {
		fmt.Println("starting escrowd bot...")
		bot.Start()
		return
	}
	fmt.Println("usage: escrowd bot")
	fmt.Println("  starts the Discord bot")
	os.Exit(1)
}