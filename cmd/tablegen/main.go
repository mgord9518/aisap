// Simple CLI tool to generate a markdown table of all profiles
// Sends profile database directly to stdout

package main

import (
	"fmt"
	"sort"

	profiles "github.com/mgord9518/aisap/profiles"
)

// Process flags
func main() {
	profileMap := profiles.Profiles()

	// Table header
	fmt.Printf("## Current supported applications (%d)\n", len(profileMap))
	fmt.Println("|name|level|devices|sockets|filesystem|")
	fmt.Println("|-|-|-|-|-|")

	names := make([]string, 0, len(profileMap))

	for name, _ := range profileMap {
		names = append(names, name)
	}

	sort.Strings(names)

	for _, name := range names {
		profile := profileMap[name]

		// Name
		fmt.Printf("|%s", name)

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
