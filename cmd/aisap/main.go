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

// This software is in ALPHA,  it isn't recommended to  run it for any  reasons
// besides testing!

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	aisap   "github.com/mgord9518/aisap"
	check   "github.com/mgord9518/aisap/spooky"
	clr     "github.com/gookit/color"
	flag    "github.com/spf13/pflag"
	helpers "github.com/mgord9518/aisap/helpers"
)

var (
	ai   *aisap.AppImage
	argv0 string
)

// Process flags
func main() {
	if len(flag.Args()) < 1 {
		flag.Usage()
	}

	ai, err := aisap.NewAppImage(flag.Args()[0])
	defer ai.Unmount()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to open AppImage:", err)
		return
	}

	if *profile != "" {
		err := ai.Perms.SetPerms(*profile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to get permissions from file:", err)
			return
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
		err := ai.Perms.SetLevel(*level)
		if err != nil {
			clr.Fprintln(os.Stderr, "<red>error</> (this shouldn't happen!): failed to set permissions level:", err)
		}
	}

	if ai.Perms.Level < 0 || ai.Perms.Level > 3 {
		clr.Println("<yellow>info</>: this app has no profile! defaulting to level 3")
		clr.Println("use the command line flag <cyan>--level</> [<green>1</>-<green>3</>] to try to sandbox it anyway\n")
		ai.Perms.SetLevel(3)
	}

	if *listPerms && ai.Perms.Level == 0 {
		clr.Println("<yellow>permissions</>:")
		prettyList("level", 0, 11)
		prettyList("filesystem", "ALL", 11)
		prettyList("devices", "ALL", 11)
		prettyList("sockets", "ALL", 11)

		clr.Printf("\n<lightYellow>warning</>: this app requests to be unsandboxed\n")
		clr.Println("use the command line flag <cyan>--level</> [<green>1</>-<green>3</>] to try to sandbox it anyway\n")
		return
	}

	// Give basic info on the permissions the AppImage requests
	if *listPerms {
		clr.Println("<yellow>permissions</>:")

		prettyList("level", ai.Perms.Level, 11)
		prettyListFiles("filesystem", ai.Perms.Files, 11)
		prettyList("devices", ai.Perms.Devices, 11)
		prettyListSockets("sockets", ai.Perms.Sockets, 11)
		fmt.Println()

		// Warns if the AppImage contains potential escape vectors or suspicious files
		for _, v := range(ai.Perms.Files) {
			if check.IsSpooky(v) {
				clr.Fprintf(os.Stdout, "<lightYellow>warning</>: this app requests files/ directories that can be used to escape sandboxing\n")
				break
			}
		}

		spookySockets := []string{
			"session",
			"x11",
		}
		if _, present := helpers.ContainsAny(ai.Perms.Sockets, spookySockets); present {
			clr.Fprintf(os.Stdout, "<lightYellow>warning</>: sockets requested by this app can be used to escape the sandbox\n")
		}

		return
	}

	// Sandbox only if level is above 0
	err = ai.Run(flag.Args()[1:])

	if err != nil {
		fmt.Fprintln(os.Stdout, "exited non-zero status:", err)
		return
	}
}

func handleCtrlC() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		ai.Unmount()
	}()
}
