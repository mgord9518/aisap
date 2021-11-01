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
    "path/filepath"

    aisap    "github.com/mgord9518/aisap"
    flag     "github.com/spf13/pflag"
    helpers  "github.com/mgord9518/aisap/helpers"
    profiles "github.com/mgord9518/aisap/profiles"
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

    // Iterate through added files and make them absolute paths
    for i, _ := range(addFile) {
        f, _ := filepath.Abs(addFile[i])
        addFile[i] = f
    }

    // Add extra permissions as passed from flags. eg: `--add-file`
    aisap.AddFilePerms(addFile)
    aisap.AddDevicePerms(addDev)
    aisap.AddSocketPerms(addSoc)
    aisap.AddSharePerms(addSoc)

    var perm *profiles.AppImagePerms

    if *permFile == "" {
        perm, err = aisap.GetPermsFromAppImage(ai, *level)
    } else {
        perm, err = aisap.GetPermsFromEntry(*permFile, *level)
    }

    helpers.ErrorCheck("Failed to get AppImage permissions:", err, true)

	// Sandbox if level is above 0, run normally if 0
	if *level > 0 {
		err = aisap.Wrap(ai, perm, flag.Args()[1:])
	} else if *level == 0 {
		err = aisap.Run(ai, flag.Args()[1:])
	}

    helpers.ErrorCheck("Failed to wrap AppImage:", err, true)

	// Unmount the AppImage, otherwise the user's temporary directory will
	// get cluttered with mountpoints
	aisap.UnmountAppImage(ai)
}
