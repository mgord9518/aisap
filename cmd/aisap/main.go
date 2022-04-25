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
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	aisap       "github.com/mgord9518/aisap"
	permissions "github.com/mgord9518/aisap/permissions"
	cli         "github.com/mgord9518/cli"
	check       "github.com/mgord9518/aisap/spooky"
	flag        "github.com/spf13/pflag"
	helpers     "github.com/mgord9518/aisap/helpers"
	xdg         "github.com/adrg/xdg"
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
		cli.Fatal("failed to open AppImage:", err)
		return
	}

	// Currently only shImgs (type -2) don't need to be mounted to extract
	// desktop integration info but I'll soon fix this for type 2 AppImages
	// and eventually add support for type 1 (low priority as they're getting
	// less and less common) but as of now type 2s are automatically mounted
	// regardless of being sandboxed or not
	if *verbose && ai.Type() != -2 {
		cli.Notify("<blue>" + strings.Replace(ai.Path, xdg.Home, "~", 1) + " </>mounted at", ai.MountDir())
	}

	if *extractIcon != "" {
		if *verbose {
			cli.Notify("extracting icon to", *extractIcon)
		}

		icon, _, err := ai.Icon()
		defer icon.Close()
		if err != nil {
			cli.Fatal("failed to extract icon:", err)
			return
		}

		f, err := os.Create(*extractIcon)
		if err != nil {
			cli.Fatal("failed to extract icon:", err)
			return
		}

		_, err = io.Copy(f, icon)
		if err != nil {
			cli.Fatal("failed to extract icon:", err)
		}
		return
	}

	if *extractThumbnail != "" {
		if *verbose {
			cli.Notify("extracting thumbnail preview to", *extractThumbnail)
		}

		thumbnail, err := ai.Thumbnail()
		if err != nil {
			cli.Fatal("failed to extract thumbnail:", err)
			return
		}

		f, err := os.Create(*extractThumbnail)
		if err != nil {
			cli.Fatal("failed to extract thumbnail:", err)
			return
		}

		_, err = io.Copy(f, thumbnail)
		if err != nil {
			cli.Fatal("failed to extract thumbnail:", err)
		}
		return
	}

	if *profile != "" {
		f, err := os.Open(*profile)
		if err != nil {
			cli.Fatal("failed to get permissions from profile:", err)
			return
		}

		ai.Perms, err = permissions.FromReader(f)
		if err != nil {
			cli.Fatal("failed to get permissions from profile:", err)
			return
		}

		if err != nil {
			cli.Fatal("failed to get permissions from profile:", err)
			return
		}
	}

	// Add (and remove) permissions as passed from flags. eg: `--add-file`
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
			cli.Fatal("failed to set permissions level (this shouldn't happen!):", err)
			return
		}
	}

	noProfile := false

	if ai.Perms.Level < 0 || ai.Perms.Level > 3 {
		ai.Perms.SetLevel(3)
		noProfile = true
	}

	// Give basic info on the permissions the AppImage requests
	if *listPerms {
		cli.ListPerms(ai.Perms)

		fmt.Println()

		if ai.Perms.Level == 0 {
			cli.Warning("this app requests to be unsandboxed!")
			cli.Warning("use the CLI flag <cyan>--level</> <gray>[</><green>1</><gray>..</><green>3</><gray>]</> to try to sandbox it anyway")
			return
		}

		if noProfile {
			cli.Notify("this app has no profile! defaulting to level 3")
			cli.Notify("use the CLI flag <cyan>--level</> <gray>[</><green>1</><gray>..</><green>3</><gray>]</> to try to sandbox it anyway")
		}

		// Warns if the AppImage contains potential escape vectors or suspicious files
		for _, v := range(ai.Perms.Files) {
			if check.IsSpooky(v) {
				cli.Warning("this app requests files/ directories that could be used to escape sandboxing")
				break
			}
		}

		spookySockets := []string{
			"session",
			"x11",
		}
		if _, present := helpers.ContainsAny(ai.Perms.Sockets, spookySockets); present {
			cli.Warning("sockets used by this app could be used to escape the sandbox")
		}

		return
	}

	ai.Mount()

	if *setRoot != "" {
		ai.SetRootDir(*setRoot)
	}

	if *verbose && ai.Type() == -2 {
		cli.Notify("<blue>" + strings.Replace(ai.Path, xdg.Home, "~", 1) + " </>mounted at", ai.MountDir())
	}

	if *verbose {
		wrapArg, _ := ai.WrapArgs([]string{})
		cli.Notify("running with sandbox base level", ai.Perms.Level)
		cli.Notify("bwrap flags:", wrapArg)
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
		if *verbose {
			fmt.Println()
			cli.Notify("quitting because <gray>[</><green>ctrl</><gray>]+</><green>c</> was hit!")
		}
		ai.Unmount()
	}()
}
