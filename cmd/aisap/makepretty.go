// Print everything in pretty-colored, valid YAML.
// Why YAML? No practical reason, I just think it looks nice and I guess if you
// really wanted to it would be easier to parse than unformatted text

package main

import (
	"fmt"
	"strings"
	"path/filepath"

//	helpers "github.com/mgord9518/aisap/helpers"
	check "github.com/mgord9518/aisap/spooky"
	xdg   "github.com/adrg/xdg"
	clr   "github.com/gookit/color"
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
//func prettyListFiles(a ...interface{}) {
//	for i := range(a) {
//		if i == 0 {
//			clr.Printf("  <cyan>%s</>:", a[0])
//			continue
//		} else if i == len(a)-1 {
//			break	
//		}
//
//		str := a[0].(string)
//		if len(str) == 0 {
//			return
//		}
//
//		clr.Printf("  <cyan>%s</>:", str)
//
//		n := a[len(a)-1].(int)
//		for i := len(str); i < n; i++ {
//			fmt.Print(" ")
//		}
//
//		fmt.Print("[")
//		for i := range(a) {
//			a[i] = strings.Replace(a[i].(string), xdg.Home, "~", 1)
//
//			if i > 0 {
//				fmt.Printf(", ")
//			}
//
//			if check.IsSpooky(a[i].(string)) {
//				clr.Printf("<yellow>%s</>", a[i])
//			} else {
//				clr.Printf("<green>%s</>", a[i])
//			}
//		}
//		fmt.Print("]")
//
//		fmt.Println()
//	}
//}
