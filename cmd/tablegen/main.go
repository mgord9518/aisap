// Simple CLI tool to generate a markdown table of all profiles
// Sends profile database directly to stdout

package main

import (
	"fmt"

	profiles "github.com/mgord9518/aisap/profiles"
)

// Process flags
func main() {
	// Table header
	fmt.Printf("## Current supported applications (%d)\n", len(profiles.RawProfiles))
	fmt.Println("|name|level|devices|sockets|filesystem|")
	fmt.Println("|-|-|-|-|-|")

	for _, profile := range profiles.RawProfiles {
		// Name
		fmt.Printf("|%s", profile.Names[0])

		// Level
		fmt.Printf("|%d", profile.Level)

		// Devices
		fmt.Printf("|")
		for i, device := range profile.Devices {
			if i > 0 {
				fmt.Printf(", ")
			}

			fmt.Printf("%s", device)
		}

		// Sockets
		fmt.Printf("|")
		for i, socket := range profile.Sockets {
			if i > 0 {
				fmt.Printf(", ")
			}

			fmt.Printf("%s", socket)
		}

		// Filesystem
		fmt.Printf("|")
		for i, file := range profile.Files {
			if i > 0 {
				fmt.Printf(", ")
			}

			fmt.Printf("%s", file)
		}

		fmt.Println("|")
	}
}
