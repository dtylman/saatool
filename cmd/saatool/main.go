package main

import (
	"log"

	"github.com/dtylman/saatool/ui"
)

func main() {
	ui.MemoryLog.Init() // Initialize the memory log buffer
	log.Printf("Starting SaaTool application...")
	err := ui.NewMainWindow()
	if err != nil {
		log.Fatalf("Failed to create main window: %v", err)
	}
	ui.Main.ShowAndRun()
}
