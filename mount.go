package aisap

import (
	"os"
	"os/exec"
	"strconv"
	"errors"
	"path"
)

// Mount mounts the requested AppImage (src) to the destination
// directory (dest)
// Quick, hacky implementation, ideally this should be redone using the
// squashfuse library
func Mount(src string, dest string, offset int) error {
	var squashfuse string
	var err error

	e, _ := os.Executable()

	if squashfuse, err = exec.LookPath("squashfuse"); err != nil {
		squashfuse, err = exec.LookPath(path.Dir(e)+"/squashfuse")
		if err != nil {
			return errors.New("Failed to find squashfuse binary! Cannot mount AppImage")
		}
	}

	n := strconv.Itoa(offset)

	mnt = exec.Command(squashfuse, "-o", "offset="+n, src, dest)
	err = mnt.Run()

	return err
}

// Unmounts an AppImage
func Unmount(ai *AppImage) error {
	if (ai == nil) {
		return errors.New("AppImage is nil")
	} else if ai.Path == "" {
		return errors.New("AppImage contains no path")
	}

	err = unmountDir(ai.MountDir())
	if err != nil { return err }

	// Clean up
	if ai.rmMountDir {
		err = os.RemoveAll(ai.TempDir())
	}

	return err
}

// Unmounts a directory
func unmountDir(mntPt string) error {
	var mntCmd string
	var err error

	if mntCmd, err = exec.LookPath("fusermount"); err != nil {
		mntCmd, err = exec.LookPath("mount")
	}

	umount := exec.Command(mntCmd, "-u", mntPt)
	out, err := umount.CombinedOutput()

	if err != nil {
		err = errors.New(string(out))
	}

	return err
}
