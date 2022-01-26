// Copyright © 2021-2022 Mathew Gordon <github.com/mgord9518>
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
	"syscall"

	aisap "github.com/mgord9518/aisap"
	check "github.com/mgord9518/aisap/spooky"
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
			fmt.Fprintln(os.Stderr, "failed to open AppImage:", err)
			cleanExit(1)
		}
	} else {
		flag.Usage()
	}

	if *profile != "" {
		err = ai.Perms.SetPerms(*profile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to get permissions from file:", err)
			cleanExit(1)
		}
	}

	// Add (and remove) permissions as passed from flags. eg: `--file`
	// Note: If *not* using XDG standard names (eg: `xdg-desktop`) you MUST
	// Provide the full filepath when using `AddFiles`
	ai.Perms.RemoveFiles(rmFile)
	ai.Perms.RemoveDevices(rmDevice)
	ai.Perms.RemoveSockets(rmSocket)
	ai.Perms.AddFiles(addFile)
	ai.Perms.AddDevices(addDevice)
	ai.Perms.AddSockets(addSocket)

	// If the `--level` flag is used, set the AppImage to that level
	if *level > -1 && *level <= 3 {
		err = ai.Perms.SetLevel(*level)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to set permissions level:", err)
		}
	}

	if ai.Perms.Level < 0 || ai.Perms.Level > 3 {
		fmt.Println("failed to retrieve AppImage permissions!")
		fmt.Println("defaulting sandbox level to 3 with no further access")
		fmt.Println("in the case this sandbox does not work properly, use the command line")
		fmt.Println("flags to add the necessary minimum permissions or create a custom profile")
		ai.Perms.SetLevel(3)
	}

	if *listPerms && ai.Perms.Level == 0 {
		fmt.Fprintf(os.Stdout, "%sapplication `%s` requests to be used unsandboxed!%s\n", y, ai.Name, z)
		fmt.Fprintln(os.Stdout, "Use the command line flag `--level [1-3]` to try to sandbox it anyway")
		cleanExit(0)
	}

	// Give basic info on the permissions the AppImage requests
	if *listPerms {
		fmt.Printf("%spermissions: \n", y)

		fmt.Printf("%s - %slevel:      %s%d\n", g, z, c, ai.Perms.Level)
		prettyListFiles("filesystem: ", ai.Perms.Files)
		prettyList("devices:    ", ai.Perms.Devices)
		prettyList("sockets:    ", ai.Perms.Sockets)

		// Warns if the AppImage contains potential escape vectors or suspicious files
		for _, v := range(ai.Perms.Files) {
			if check.IsSpooky(v) {
				fmt.Fprintf(os.Stdout, "\n%sWARNING: this AppImage requests files/ directories that can potentially\n", y)
				fmt.Fprintln(os.Stdout, "be used to escape the sandbox (shown highlighted orange in the filesystem list)")
				break
			}
		}

		cleanExit(0)
	}

	// Sandbox only if level is above 0
	if ai.Perms.Level > 0 {
		err = aisap.Sandbox(ai, flag.Args()[1:])
	} else {
		err = aisap.Run(ai, flag.Args()[1:])
	}

	if err != nil {
		fmt.Fprintln(os.Stdout, "exited non-zero status:", err)
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
