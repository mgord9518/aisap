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
	"os/signal"
	"strconv"
	"syscall"

	aisap "github.com/mgord9518/aisap"
	flag  "github.com/spf13/pflag"
)

var (
	ai   *aisap.AppImage
	err   error
	argv0 string
)

// Process flags
func main() {
	if len(flag.Args()) >= 1 {
		ai, err = aisap.NewAppImage(flag.Args()[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to open AppImage:", err)
			cleanExit(1)
		}
	} else {
		flag.Usage()
	}

	if *profile != "" {
		err = ai.SetPerms(*profile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to get permissions from profile:", err)
			cleanExit(1)
		}
	}

	// Add extra permissions as passed from flags. eg: `--file`
	// Note: If *not* using XDG standard names (eg: `xdg-desktop`) you MUST
	// Provide the full filepath when using `AddFiles`
	ai.AddFiles(file)
	ai.AddDevices(device)
	ai.AddSockets(socket)

	// If the `--level` flag is used, set the AppImage to that level
	if *level > -1 && *level <= 3 {
		err = ai.SetLevel(*level)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to set permissions level:", err)
		}
	}

	if ai.Perms.Level < 0 || ai.Perms.Level > 3 {
		fmt.Println("Failed to retrieve AppImage permissions!")
		fmt.Println("Defaulting sandbox level to 3 with no further access")
		fmt.Println("In the case this sandbox does not work properly, use the command line")
		fmt.Println("flags to add the necessary minimum permissions or create a custom profile")
		ai.SetLevel(3)
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
			fmt.Printf("%sFiles and directories:\n", y)
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
				v = makeDevPretty(v)
				fmt.Printf(" %s>%s %s\n", g, z, v)
			}
		}
		if len(ai.Perms.Sockets) > 0 {
			fmt.Printf("%sSockets:\n", y)
			for _, v := range(ai.Perms.Sockets) {
				fmt.Printf(" %s>%s %s\n", g, z, v)
			}
		}
		if spookyBool {
			fmt.Fprintf(os.Stdout, "\n%sWARNING: This AppImage requests files/ directories that can potentially\n", y)
			fmt.Fprintln(os.Stdout, "be used to escape the sandbox (shown with red arrow under the file list)\n")
		}
	} else if *listPerms && ai.Perms.Level == 0 {
		fmt.Fprintf(os.Stdout, "%sApplication `"+ai.Name+"` requests to be used unsandboxed!%s\n", y, z)
		fmt.Fprintln(os.Stdout, "Use the command line flag `--level [1-3]` to try to sandbox it anyway")
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
		fmt.Fprintln(os.Stdout, "Failed to sandbox AppImage:", err)
		cleanExit(1)
	}

	cleanExit(0)
}

func cleanExit(exitCode int) {
	err = aisap.Unmount(ai)
	os.Exit(exitCode)
}

func handleCtrlC() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		cleanExit(0)
	}()
}
