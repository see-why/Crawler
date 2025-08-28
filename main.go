package main

import (
	"fmt"
	"os"
)

func main() {
	// Get command line arguments (excluding program name)
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	if len(args) > 1 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}

	// Exactly one argument - the BASE_URL
	baseURL := args[0]
	fmt.Printf("starting crawl of: %s\n", baseURL)
}
