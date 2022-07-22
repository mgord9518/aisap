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
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	aisap       "github.com/mgord9518/aisap"
	permissions "github.com/mgord9518/aisap/permissions"
	cli         "github.com/mgord9518/cli"
	check       "github.com/mgord9518/aisap/spooky"
	flag        "github.com/spf13/pflag"
	helpers     "github.com/mgord9518/aisap/helpers"
	ini         "gopkg.in/ini.v1"
	clr         "github.com/gookit/color"
)

var (
	ai   *aisap.AppImage
	argv0 string

	invalidBundle          = errors.New("failed to open bundle:")
	invalidIcon            = errors.New("failed to extract icon:")
	invalidThumbnail       = errors.New("failed to extract thumbnail preview:")
	invalidPerms           = errors.New("failed to get permissions from profile:")
	invalidPermLevel       = errors.New("failed to set permissions level (this shouldn't happen!):")
	invalidFallbackProfile = errors.New("failed to set fallback profile:")
)

// Process flags
func main() {
	if len(flag.Args()) < 1 {
		flag.Usage()
	}

	ai, err := aisap.NewAppImage(flag.Args()[0])
	defer ai.Destroy()

	if err != nil {
		cli.Fatal(invalidBundle, err)
		return
	}

	if *extractIcon != "" {
		if *verbose {
			cli.Notify("extracting icon to", *extractIcon)
		}

		icon, _, err := ai.Icon()
		defer icon.Close()
		if err != nil {
			cli.Fatal(invalidIcon, err)
			return
		}

		f, err := os.Create(*extractIcon)
		if err != nil {
			cli.Fatal(invalidIcon, err)
			return
		}

		_, err = io.Copy(f, icon)
		if err != nil {
			cli.Fatal(invalidIcon, err)
		}

		return
	}

	if *extractThumbnail != "" {
		if *verbose {
			cli.Notify("extracting thumbnail preview to", *extractThumbnail)
		}

		thumbnail, err := ai.Thumbnail()
		if err != nil {
			cli.Fatal(invalidThumbnail, err)
			return
		}

		f, err := os.Create(*extractThumbnail)
		if err != nil {
			cli.Fatal(invalidThumbnail, err)
			return
		}

		_, err = io.Copy(f, thumbnail)
		if err != nil {
			cli.Fatal(invalidThumbnail, err)
		}

		return
	}

	if *profile != "" {
		f, err := os.Open(*profile)
		if err != nil {
			cli.Fatal(invalidPerms, err)
			return
		}

		ai.Perms, err = permissions.FromReader(f)
		if err != nil {
			cli.Fatal(invalidPerms, err)
			return
		}

		if err != nil {
			cli.Fatal(invalidPerms, err)
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
			cli.Fatal(invalidPermLevel, err)
			return
		}
	}

	noProfile := false

	// Fallback on `--fallback-profile` if set, otherwise just set base level to 3
	if ai.Perms.Level < 0 || ai.Perms.Level > 3 {
		if *fallbackProfile != "" {
			f, err := ini.LoadSources(ini.LoadOptions{
				IgnoreInlineComment: true,
			}, *fallbackProfile)

			if err != nil {
				cli.Fatal(invalidFallbackProfile, err)
				return
			}

			ai.Perms, err = permissions.FromIni(f)
		} else {
			ai.Perms.Level = 3
		}

		noProfile = true
	}

	// Give basic info and the permissions the AppImage requests
	if *listPerms {
		if *verbose {
			a := ai.Desktop.Section("Desktop Entry").Key("X-AppImage-Architecture").Value()
			s := helpers.SplitKey(a)

			ts := ""
			switch ai.Type() {
			case -2:
				ts = "shImg (SquashFS)"
			case 1:
				ts = "AppImage (ISO 9660)"
			case 2:
				ts = "AppImage (SquashFS)"
			}
			clr.Println("<yellow>bundle info</>:")
			cli.List("name", ai.Name, 11)
			cli.List("type", ts, 11)
			cli.List("version", ai.Version, 11)
			cli.List("update info", ai.UpdateInfo, 11)
			cli.List("arch", s, 11)
			fmt.Println()
		}

		cli.ListPerms(ai.Perms)

		fmt.Println()

		if ai.Perms.Level == 0 {
			cli.Warning("this app requests to be unsandboxed!")
			cli.Warning("use the CLI flag <cyan>--level</> <gray>[</><green>1</><gray>..</><green>3</><gray>]</> to sandbox it anyway")
			return
		}

		if noProfile {
			cli.Notify("this app has no profile! falling back to default")
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

	if *rootDir != "" {
		ai.SetRootDir(*rootDir)
	}

	if *dataDir != "" {
		ai.SetDataDir(*dataDir)
	}

	if *noDataDir {
		ai.Perms.NoDataDir = true
	}

	if *verbose {
		wrapArg, _ := ai.WrapArgs([]string{})
		cli.Notify("running with sandbox base level", ai.Perms.Level)
		cli.Notify("bwrap flags:", wrapArg)
	}

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
