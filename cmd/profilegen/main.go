// Simple CLI tool to generate JSON formatted profiles from aisap
// Eventually the main database may be switched to JSON, but for now it's
// just being provided for export

// Sends profile database directly to stdout

package main

import (
	"fmt"
	"encoding/json"

	profiles    "github.com/mgord9518/aisap/profiles"
	permissions "github.com/mgord9518/aisap/permissions"
)

type ProfileDatabase struct {
	Profiles map[string]permissions.AppImagePerms `json:"profiles"`
}

// Process flags
func main() {
	database := &ProfileDatabase{}
	database.Profiles = profiles.Profiles()

	jsonData, _ := json.Marshal(database)

	fmt.Println(string(jsonData))
}
