package main

import (
	"escrowd/cmd/escrowd"
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
	escrowd.Run()
}
