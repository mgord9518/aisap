package spooky

import (
	//"encoding/json"
	"strings"

	_ "embed"
)

type Database struct {
    Paths []string `json:"paths"`
    Trees []string `json:"trees"`
}

//go:embed spooky_database.json
var jsonDatabase []byte

// TODO: finish (need state machine to parse env variables out of the JSON)
func isSpookyNew(path string) bool {
    return true;
}

// Check if a file or directory is spooky (sandbox escape vector or possibly
// suspiscious files to request) so that the user can be warned that their
// sandbox may be insecure or leak information from other applications
// TODO: expand this list! There are a lot of files that can be used to escape
// the sandbox, while it's impossible to cover all bases, we should try to get
// as close as possible
func IsSpooky(str string) bool {
	// An app with access to these specific paths may be able to use them for escape
	spookyFiles := []string{
		"~",
		"/",
		"/etc",
		"/home",
		"~/.profile",
		"~/.bashrc",
		"~/.zshrc",
	}

	// Requesting these, or any file under them at all could be a potential
	// escape route or leak personal information
	spookyDirs := []string{
		"~/Apps",
		"~/AppImages",
		"~/Applications",
		"~/go",
		"~/.cache",
		"~/.ssh",
		"~/.vim",
		"~/.gnupg",
		"~/.firefox",
		"~/.mozilla",
		"~/.local",
		"~/.config",
		"/tmp",
		"/run",
	}

	// Split the string into its actual directory and whether it's read only or
	// read write
	slice := strings.Split(str, ":")
	s1 := strings.Join(slice[:len(slice)-1], ":")

	for _, val := range spookyFiles {
		if s1 == val {
			return true
		}
	}

	for _, val := range spookyDirs {
		if len(s1) < len(val) {
			continue
		}

		if len(s1) == len(val) {
			if s1[:len(val)] == val {
				return true
			} else {
				continue
			}
		}
	}

	return false
}
