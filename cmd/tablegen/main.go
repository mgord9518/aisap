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
	profileKeys := make([]string, 0, len(profiles.Profiles()))

	for key, _ := range profiles.Profiles() {
		profileKeys = append(profileKeys, key)
	}

	sort.Strings(profileKeys)

	// Table header
	fmt.Printf("## Current supported applications (%d)\n", len(profileKeys))
	fmt.Println("|name|level|devices|sockets|filesystem|")
	fmt.Println("|-|-|-|-|-|")

	for i := range profileKeys {
		profile := profiles.Profiles()[profileKeys[i]]

		// Name
		fmt.Printf("|")

		for i, name := range profile.Names {
			fmt.Printf("%s", name)

			if i < len(profile.Names) - 1 {
				fmt.Printf(", ")
			}
		}

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
