// Copyright © 2021 Mathew Gordon <github.com/mgord9518>
// AppImage SAndboxing Program (aisap)
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

// This package can be used as a drop-in replacement for go-appimage, but it IS
// slower, the only good use cases for doing so would be adding support for
// shappimages (I'll publish this project soon) or adding sandboxing
// ^^^ NOT YET IT CAN'T! Still missing basic features, priority is stable
// sandboxing, and this probably won't be a viable alternative for a while

// This software is in ALPHA,  it isn't recommended to  run it for any  reasons
// besides testing!

package aisap

import (
	"os"
	"os/user"
	"os/exec"
	"strconv"
)

var (
	usern string
	homed string
	uid   string

	dataDir	   string
	rootDir	 = "/"
	tempDir	 = "/tmp"
	mnt		 *exec.Cmd

	preFilePerms   []string
	preSocketPerms []string
	preSharePerms  []string
	preDevicePerms []string
)

func loadName(permLevel int) {
	if permLevel == 1 {
		usr, _ := user.Current()
		uid   = strconv.Itoa(os.Getuid())
		usern = usr.Username
		homed = "/home/"+usern
	} else {
		uid   = "256"
		usern = "ai"
		homed = "/home/"+usern
	}
}
