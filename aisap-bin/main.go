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
	"os"
	"strconv"

	aisap	"github.com/mgord9518/aisap"
	flag	 "github.com/spf13/pflag"
	helpers  "github.com/mgord9518/aisap/helpers"
)

var (
	ai   *aisap.AppImage
	err   error
	argv0 string
)

// Process flags
func main() {
	ai, err = aisap.NewAppImage(flag.Args()[0])
	helpers.ErrorCheck("Failed to load AppImage!", err, true)

	// Add extra permissions as passed from flags. eg: `--file`
	// Note: If *not* using XDG standard names (eg: `xdg-desktop`) you MUST
	// Provide the full filepath when using `AddFiles`
	ai.AddFiles(addFile)
	ai.AddDevices(addDev)
	ai.AddSockets(addSoc)
	ai.AddShare(addShare)

	// If the `--level` flag is used, set the AppImage to that level
	if *level > -1 && *level < 3 {
		ai.Perms.Level = *level
	}

	if *permFile != "" {
		err = ai.SetPerms(*permFile)
		if err != nil { fmt.Println(err) }
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
		fmt.Printf("%sSandbox base level: %s\n", y, strconv.Itoa(ai.Perms.Level))
		if ai.Perms.Level == 1 {
			fmt.Printf(" %s>%s All system files, including machine identifiable information\n", y, z)
		} else if ai.Perms.Level == 2 {
			fmt.Printf(" %s>%s Some system files such as themes\n", y, z)
		} else if ai.Perms.Level == 3 {
			fmt.Printf(" %s>%s Minimal system files\n", y, z)
		}

		if len(ai.Perms.Files) > 0 {
			fmt.Printf("%sFiles/directories:\n", y)
			for _, v := range(ai.Perms.Files) {
				fmt.Println(" \033[32m>\033[0m "+v)
			}
		}
		if len(ai.Perms.Devices) > 0 {
			fmt.Printf("%sDevice files:\n", y)
			for _, v := range(ai.Perms.Devices) {
				fmt.Println(" \033[32m>\033[0m "+v)
			}
		}
		if len(ai.Perms.Sockets) > 0 {
			fmt.Printf("%sSockets:\n", y)
			for _, v := range(ai.Perms.Sockets) {
				fmt.Println(" \033[32m>\033[0m "+v)
			}
		}
		if len(ai.Perms.Share) > 0 {
			fmt.Printf("%sShare:\n", y)
			for _, v := range(ai.Perms.Share) {
				fmt.Println(" \033[32m>\033[0m "+v)
			}
		}
	} else if *listPerms && ai.Perms.Level == 0 {
		fmt.Println("\033[33mApplication `"+ai.Name+"` requests to be used unsandboxed!\033[0m")
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

	helpers.ErrorCheck("Failed to wrap AppImage:", err, true)

	// Unmount the AppImage, otherwise the user's temporary directory will
	// get cluttered with mountpoints
	aisap.UnmountAppImage(ai)
}

func cleanExit(exitCode int) {
	err = aisap.UnmountAppImage(ai)
	os.Exit(exitCode)
}
