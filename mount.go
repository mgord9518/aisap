package aisap

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	xdg "github.com/adrg/xdg"
	helpers "github.com/mgord9518/aisap/helpers"
)

// mount mounts the requested AppImage `src` to `dest`
// Quick, hacky implementation, ideally this should be redone using the
// squashfuse library
func mount(src string, dest string, offset int) error {
	squashfuse, present := helpers.CommandExists("squashfuse")
	if !present {
		return errors.New("failed to find squashfuse binary! cannot mount AppImage")
	}

	// Store the error message in a string
	errBuf := &bytes.Buffer{}

	// Convert the offset to a string and mount using squashfuse
	o := strconv.Itoa(offset)
	mnt := exec.Command(squashfuse, "-o", "offset="+o, src, dest)
	mnt.Stderr = errBuf

	if mnt.Run() != nil {
		return errors.New(errBuf.String())
	}

	return nil
}

// Takes an optional argument to mount at a specific location (failing if it
// doesn't exist or more than one arg given. If none given, automatically
// create a temporary directory and mount to it
func (ai *AppImage) Mount(dest ...string) error {
	// If arg given
	if len(dest) > 1 {
		panic("only one argument allowed with *AppImage.Mount()!")
	} else if len(dest) == 1 {
		if !helpers.DirExists(dest[0]) {
			return NoMountPoint
		}

		if !isMountPoint(ai.mountDir) {
			return mount(ai.Path, ai.mountDir, ai.Offset)
		}

		return nil
	}

	var err error

	ai.tempDir, err = helpers.MakeTemp(xdg.RuntimeDir+"/aisap/tmp", ai.md5)
	if err != nil {
		return err
	}

	ai.mountDir, err = helpers.MakeTemp(xdg.RuntimeDir+"/aisap/mount", ai.md5)
	if err != nil {
		return err
	}

	fmt.Println(ai.mountDir)
	fmt.Println(ai.tempDir)

	// Only mount if no previous instances (launched of the same version) are
	// already mounted there. This is to reuse their libraries, save on RAM and
	// to spam the mount list as little as possible
	if !isMountPoint(ai.mountDir) {
		err = mount(ai.Path, ai.mountDir, ai.Offset)
	}

	return err
}

// Deprecated: *AppImage.Destroy() should be used instead
func (ai *AppImage) Unmount() error {
	return ai.Destroy()
}

// Unmounts an AppImage
func (ai *AppImage) Destroy() error {
	if ai == nil {
		return NilAppImage
	} else if ai.Path == "" {
		return NoPath
	} else if !ai.IsMounted() {
		return NotMounted
	}

	err := unmountDir(ai.MountDir())
	if err != nil {
		return err
	}

	ai.mountDir = ""

	ai.file.Close()

	// Clean up
	err = os.RemoveAll(ai.TempDir())

	ai = nil

	return err
}

func (ai *AppImage) IsMounted() bool {
	return ai.mountDir != ""
}

// Unmounts a directory (lazily in case the process is finishing up)
func unmountDir(mntPt string) error {
	var umount *exec.Cmd

	if _, err := exec.LookPath("fusermount"); err == nil {
		umount = exec.Command("fusermount", "-uz", mntPt)
	} else {
		umount = exec.Command("umount", "-l", mntPt)
	}

	// Run unmount command, returning the stdout+stderr if fail
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
