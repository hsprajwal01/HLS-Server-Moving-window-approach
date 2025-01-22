package main

import (
	"log"

	"app/cmd" // Adjust this import to match your project structure
)

func main() {
	// Execute the root command
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
