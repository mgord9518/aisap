// Print everything in pretty-colored, valid YAML.
// Why YAML? No practical reason, I just think it looks nice and I guess if you
// really wanted to it would be easier to parse than unformatted text for
// something like a nushell plugin
// TODO: Clean up this file! Maybe split into separate library for my future
// CLI tools?

package main

import (
	"fmt"
	"strings"
	"path/filepath"

	check   "github.com/mgord9518/aisap/spooky"
	clr     "github.com/gookit/color"
//	helpers "github.com/mgord9518/aisap/helpers"
	xdg     "github.com/adrg/xdg"
)

func makeDevPretty(str string) string {
	str = filepath.Clean(str)

	if len(str) > 5 && str[0:5] == "/dev/" {
		str = strings.Replace(str, "/dev/", "", 1)
	}

	return str
}

func prettyList(a ...interface{}) {
	for i := range(a) {
		if i == 0 {
			clr.Printf("  <cyan>%s</>:", a[0])
			continue
		} else if i == len(a)-1 {
			break	
		}

		// pad with spaces until the requested lengh is reached
		n := a[len(a)-1].(int)
		str := a[0].(string)
		for i := len(str); i < n; i++ {
			fmt.Print(" ")
		}

		switch v := a[i].(type) {
		default:
			panic("invalid type!")
		case string:
			clr.Printf(" <green>%s</>\n", a[i])
		case []string:
			fmt.Print("[")
			for i := range(v) {
				if i > 0 {
					fmt.Printf(", ")
				}
				clr.Printf("<green>%s</>", v[i])
			}
			fmt.Println("]")
		case int:
			clr.Printf(" <green>%d</>\n", a[i])
		}
	}

}

// Like `prettyList` but highlights spooky files in orange
func prettyListFiles(a ...interface{}) {
	for i := range(a) {
		if i == 0 {
			clr.Printf("  <cyan>%s</>:", a[0])
			continue
		} else if i == len(a)-1 {
			break	
		}

		// pad with spaces until the requested lengh is reached
		n := a[len(a)-1].(int)
		str := a[0].(string)
		for i := len(str); i < n; i++ {
			fmt.Print(" ")
		}

		switch v := a[i].(type) {
		default:
			panic("invalid type!")
		case []string:
			fmt.Print("[")
			for i := range(v) {
				if i > 0 {
					fmt.Printf(", ")
				}
				v[i] = strings.Replace(v[i], xdg.Home, "~", 1)

				if check.IsSpooky(v[i]) {
					clr.Printf("<yellow>%s</>", v[i])
				} else {
					clr.Printf("<green>%s</>", v[i])
				}
			}
			fmt.Println("]")
		}
	}

}

// Like `prettyList` but highlights dangerous sockets in orange
func prettyListSockets(a ...interface{}) {
	for i := range(a) {
		if i == 0 {
			clr.Printf("  <cyan>%s</>:", a[0])
			continue
		} else if i == len(a)-1 {
			break	
		}

		// pad with spaces until the requested lengh is reached
		n := a[len(a)-1].(int)
		str := a[0].(string)
		for i := len(str); i < n; i++ {
			fmt.Print(" ")
		}

		switch v := a[i].(type) {
		default:
			panic("invalid type!")
		case []string:
			fmt.Print("[")
			for i := range(v) {
				if i > 0 {
					fmt.Printf(", ")
				}
				v[i] = strings.Replace(v[i], xdg.Home, "~", 1)

				if v[i] == "session" || v[i] == "x11" {
					clr.Printf("<yellow>%s</>", v[i])
				} else {
					clr.Printf("<green>%s</>", v[i])
				}
			}
			fmt.Println("]")
		}
	}

}
