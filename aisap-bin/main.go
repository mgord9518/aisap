// Copyright © 2021 Mathew Gordon <github.com/mgord9518>
//
// Permission  is hereby  granted,  free of charge,  to any person  obtaining a
// copy of this software  and associated documentation files  (the “Software”),
// to   deal   in   the  Software   without  restriction,   including   without
// limitation the rights  to use, copy, modify, merge,   publish,   distribute,
// sublicense,  and/or sell copies of  the Software, and to  permit  persons to
// whom  the   Software  is  furnished  to  do  so,  subject  to  the following
// conditions:
// 
// The  above  copyright notice  and this permission notice  shall be  included
// in  all  copies  or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY  OF ANY KIND,  EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED  TO  THE WARRANTIES  OF  MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE  AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS  OR COPYRIGHT  HOLDERS  BE  LIABLE FOR ANY CLAIM,  DAMAGES  OR OTHER
// LIABILITY, WHETHER IN  AN  ACTION OF CONTRACT, TORT  OR  OTHERWISE,  ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

// This software is in APLHA,  it isn't recommended to  run it for any  reasons
// besides testing!

package main

import (
	"fmt"
	"path/filepath"
	"os"
	"strconv"
	"strings"

	aisap "github.com/mgord9518/aisap"
	flag  "github.com/spf13/pflag"
	xdg   "github.com/adrg/xdg"
)

var (
	ai   *aisap.AppImage
	err   error
	argv0 string
)

// Process flags
func main() {
	if len(flag.Args()) < 1 {
		flag.Usage()
	}

	ai, err = aisap.NewAppImage(flag.Args()[0])
	if err != nil {
		fmt.Println("Failed to sandbox AppImage:", err)
		cleanExit(1)
	}

	if *permFile != "" {
		err = ai.SetPerms(*permFile)
		if err != nil {
			fmt.Println(err)
			cleanExit(1)
		}
	}

	// Add extra permissions as passed from flags. eg: `--file`
	// Note: If *not* using XDG standard names (eg: `xdg-desktop`) you MUST
	// Provide the full filepath when using `AddFiles`
	ai.AddFiles(addFile)
	ai.AddDevices(addDev)
	ai.AddSockets(addSoc)
	ai.AddShare(addShare)

	// If the `--level` flag is used, set the AppImage to that level
	if *level > -1 && *level <= 3 {
		ai.Perms.Level = *level
	}

	if ai.Perms.Level == -1 {
		fmt.Println("Failed to retrieve AppImage permissions!")
		fmt.Println("Defaulting sandbox level to 3 with no further access")
		fmt.Println("In the case this sandbox does not work properly, use the command line")
		fmt.Println("flags to add the necessary minimum permissions or create a custom profile")
		ai.Perms.Level = 3
	}

	// Give basic info on the permissions the AppImage requests
	if *listPerms && ai.Perms.Level > 0 {
		var spookyBool bool
		fmt.Printf("%sSandbox base level: %s\n", y, strconv.Itoa(ai.Perms.Level))
		if ai.Perms.Level == 1 {
			fmt.Printf(" %s>%s All system files, including machine identifiable information\n", y, z)
			fmt.Printf(" %s>%s For applications that refuse to run with further sandboxing\n", y, z)
		} else if ai.Perms.Level == 2 {
			fmt.Printf(" %s>%s Some system files such as themes\n", g, z)
			fmt.Printf(" %s>%s Most GUI apps should use this\n", g, z)
		} else if ai.Perms.Level == 3 {
			fmt.Printf(" %s>%s Minimal system files\n", g, z)
			fmt.Printf(" %s>%s Console apps and the few GUI apps that work\n", g, z)
		}

		if len(ai.Perms.Files) > 0 {
			fmt.Printf("%sFiles/directories:\n", y)
			for _, v := range(ai.Perms.Files) {
				v = makePretty(v)
				if spooky(v) {
					fmt.Printf("%s", r)
					spookyBool = true
				} else {
					fmt.Printf("%s", g)
				}
				fmt.Printf(" >%s %s\n", z, v)
			}
		}
		if len(ai.Perms.Devices) > 0 {
			fmt.Printf("%sDevice files:\n", y)
			for _, v := range(ai.Perms.Devices) {
				fmt.Printf(" %s>%s %s\n", g, z, v)
			}
		}
		if len(ai.Perms.Sockets) > 0 {
			fmt.Printf("%sSockets:\n", y)
			for _, v := range(ai.Perms.Sockets) {
				fmt.Printf(" %s>%s %s\n", g, z, v)
			}
		}
		if len(ai.Perms.Share) > 0 {
			fmt.Printf("%sShare:\n", y)
			for _, v := range(ai.Perms.Share) {
				fmt.Printf(" %s>%s %s\n", g, z, v)
			}
		}
		if spookyBool {
			fmt.Printf("\n%sWARNING: This AppImage requests files/ directories that can potentially\n", y)
			fmt.Printf("be used to escape the sandbox (shown with red arrow under the file list)\n")
		}
	} else if *listPerms && ai.Perms.Level == 0 {
		fmt.Printf("%sApplication `"+ai.Name+"` requests to be used unsandboxed!%s", y, z)
		fmt.Println("Use the command line flag `--level [1-3]` to try to sandbox it anyway")
	}

	if *listPerms {
		cleanExit(0)
	}

	// Sandbox if level is above 0
	if ai.Perms.Level > 0 {
		err = aisap.Sandbox(ai, flag.Args()[1:])
	} else if ai.Perms.Level == 0 {
		err = aisap.Run(ai, flag.Args()[1:])
	}

	if err != nil {
		fmt.Println("Failed to sandbox AppImage:", err)
		cleanExit(1)
	}

	cleanExit(0)
}

func cleanExit(exitCode int) {
	err = aisap.UnmountAppImage(ai)
	os.Exit(exitCode)
}

// Convert xdg and full directories into their shortened counterparts
func makePretty(str string) string {
    var xdgDirs = map[string]string{
        "xdg-home":        xdg.Home,
        "xdg-desktop":     xdg.UserDirs.Desktop,
        "xdg-download":    xdg.UserDirs.Download,
        "xdg-documents":   xdg.UserDirs.Documents,
        "xdg-music":       xdg.UserDirs.Music,
        "xdg-pictures":    xdg.UserDirs.Pictures,
        "xdg-videos":      xdg.UserDirs.Videos,
        "xdg-templates":   xdg.UserDirs.Templates,
        "xdg-publicshare": xdg.UserDirs.PublicShare,
        "xdg-config":      xdg.ConfigHome,
        "xdg-cache":       xdg.CacheHome,
        "xdg-data":        xdg.DataHome,
    }

	// Convert xdg-style to real paths
	for key, val := range(xdgDirs) {
		str = strings.Replace(str, key, val, 1)
	}

	// Clean file path, so that stuff like `xdg-desktop/..` doesn't get past
	slice := strings.Split(str, ":")
	s1 := strings.Join(slice[:len(slice)-1], ":")
	s2 := ":"+strings.Split(str, ":")[len(slice)-1]
	str = filepath.Clean(s1)+s2

	// Pretty it up by replacing `/home/$USERNAME` with `~`
	str = strings.Replace(str, xdg.Home, "~", 1)

	return str
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
