package aisap

import (
	"bufio"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"errors"
	"path"
	"path/filepath"
)

// mount mounts the requested AppImage (src) to the destination
// directory (dest)
// Quick, hacky implementation, ideally this should be redone using the
// squashfuse library
func mount(src string, dest string, offset int) error {
	var squashfuse string

	e, err := os.Executable()
	if err != nil {
		return errors.New("failed to find executable! this shouldn't happen")
	}

	// Make sure squashfuse exists
	if squashfuse, err = exec.LookPath("squashfuse"); err != nil {
		squashfuse, err = exec.LookPath(filepath.Join(path.Dir(e), "squashfuse"))
		if err != nil {
			return errors.New("failed to find squashfuse binary! cannot mount AppImage")
		}
	}

	// Convert the offset to a string and mount using squashfuse
	o := strconv.Itoa(offset)
	mnt := exec.Command(squashfuse, "-o", "offset=" + o, src, dest)

	return mnt.Run()
}

// Unmounts an AppImage
func Unmount(ai *AppImage) error {
	if (ai == nil) {
		return errors.New("AppImage is nil")
	} else if ai.Path == "" {
		return errors.New("AppImage contains no path")
	}

	err := unmountDir(ai.MountDir())
	if err != nil { return err }

	// Clean up
	err = os.RemoveAll(ai.TempDir())

	return err
}

// Unmounts a directory (lazily in case the process is finishing up)
func unmountDir(mntPt string) error {
	var umount *exec.Cmd

	if _, err := exec.LookPath("fusermount"); err == nil {
		umount = exec.Command("fusermount", "-uz", mntPt)
	} else {
		umount = exec.Command("umount", "-l", mntPt)
	}

	// Run unmount command, returning the stdout if failed
	out, err := umount.CombinedOutput()
	if err != nil {
		err = errors.New(string(out))
	}

	return err
}

// Returns true if directory is detected as already being mounted
func isMountPoint(dir string) bool {
	f, _ := os.Open("/proc/self/mountinfo")

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		str := strings.Split(scanner.Text(), " ")[4]
		if str == dir {
			return true
		}
	}

	return false
}
