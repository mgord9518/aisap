package main

import (
	"fmt"
	"strings"
	"path/filepath"

	aisap "github.com/mgord9518/aisap"
	check "github.com/mgord9518/aisap/spooky"
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

func prettyList(str string, s []string) {
	if len(s) == 0 {
		return
	}

	fmt.Printf("%s - %s%s%s", g, z, str, c)

	for i := range(s) {
		if i > 0 {
			fmt.Printf(", ")
		}

		fmt.Printf(s[i])
	}

	fmt.Println()
}

// Like `prettyList` but highlights spooky files in orange
func prettyListFiles(str string, s []string) {
	if len(s) == 0 {
		return
	}

	fmt.Printf("%s - %s%s", g, z, str)

	for i := range(s) {
//		s[i] = makePretty(s[i])

		if i > 0 {
			fmt.Printf(", ")
		}

		if check.IsSpooky(s[i]) {
			fmt.Printf(y)
		} else {
			fmt.Printf(c)
		}

		fmt.Printf(s[i])

	}

	fmt.Println()
}
