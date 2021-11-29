package main

import (
	"strings"
	"path/filepath"

	aisap "github.com/mgord9518/aisap"
	xdg   "github.com/adrg/xdg"
)

func makeDevPretty(str string) string {
	str = filepath.Clean(str)

	if len(str) > 5 && str[0:5] == "/dev/" {
		str = strings.Replace(str, "/dev/", "", 1)
	}

	return str
}

// Convert xdg and full directories into their shortened counterparts
func makePretty(str string) string {
	s  := strings.Split(str, ":")
	str = aisap.ExpandDir(str)
	ex := ":" + s[len(s)-1]

	// Pretty it up by replacing `/home/$USERNAME` with `~`
	str = strings.Replace(str, xdg.Home, "~", 1)


	return str + ex
}

// Check if a file or directory is spooky (sandbox escape vector) so that the
// user can be warned that their sandbox is insecure
// TODO: expand this list! There are a lot of files that can be used to escape
// the sandbox, while it's impossible to cover all bases, we should try to get
// as close as possible
func spooky(str string) bool {
	// These files/ directories are specifically escape vectors on their own
	spookyFiles := []string{
		"~",
		"/home",
		"~/Apps",
		"~/Applications",
		"~/AppImages",
		"~/.profile",
		"~/.bashrc",
		"~/.zshrc",
	}

	// If the sandbox requests these directories at all, it is a potential threat
	spookyDirs := []string{
		"~/.ssh",
		"~/.local",
		"~/.config",
	}

	// Split the string into its actual directory and whether it's read only or
	// read write
	slice := strings.Split(str, ":")
	s1 := strings.Join(slice[:len(slice)-1], ":")
	s2 := ":"+strings.Split(str, ":")[len(slice)-1]

	for _, val := range(spookyFiles) {
		if s1 == val && s2 == ":rw" {
			return true
		}
	}

	for _, val := range(spookyDirs) {
		if len(s1) >= len(val) {
			if s1[:len(val)] == val {
				return true
			}
		}
	}

	return false
}
