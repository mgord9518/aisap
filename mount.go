package aisap

import (
	"bufio"
	"bytes"
	"path"
	"path/filepath"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"errors"

	helpers "github.com/mgord9518/aisap/helpers"
	xdg     "github.com/adrg/xdg"
)

// mount mounts the requested AppImage `src` to `dest`
// Quick, hacky implementation, ideally this should be redone using the
// squashfuse library
// Also planning on not requiring AppImages to be mounted unless they're being
// sandboxed
func mount(src string, dest string, offset int) error {
	squashfuse, present := helpers.CommandExists("squashfuse")
	if !present {
		return errors.New("failed to find squashfuse binary! cannot mount AppImage")
	}

	// Store the error message in a string
	errBuf := &bytes.Buffer{}

	// Convert the offset to a string and mount using squashfuse
	o := strconv.Itoa(offset)
	mnt := exec.Command(squashfuse, "-o", "offset=" + o, src, dest)
	mnt.Stderr = errBuf

	if mnt.Run() != nil {
		return errors.New(errBuf.String())
	}

	return nil
}

// Experimental mounting through Go squashfuse implementaion
//func mountNoBin(src string, dest string, offset int) error (
	
//)

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

	pfx := path.Base(ai.Path)
	if len(pfx) > 6 {
		pfx = pfx[0:6]
	}

	// Generate a seed based on the AppImage URI MD5sum. This shouldn't cause
	// any issues as AppImages will have a different path given a different
	// version
	seed, err := strconv.ParseInt(ai.md5[0:15], 16, 64)
	ai.runId = pfx + helpers.RandString(int(seed), 6)

	ai.tempDir, err = helpers.MakeTemp(filepath.Join(xdg.RuntimeDir, "aisap"), ai.runId)
	if err != nil { return err }

	ai.mountDir, err = helpers.MakeTemp(ai.tempDir, ".mount_" + ai.runId)
	if err != nil { return err }

	// Only mount if no previous instances (launched of the same version) are
	// already mounted there. This is to reuse their libraries, save on RAM and
	// to spam the mount list as little as possible
	//if _, present := os.LookupEnv("AISAP_EXPERIMENTAL"); present {
	//	if !isMountPoint(ai.mountDir) {
	//		err = mountNoBin(ai.Path, ai.mountDir, ai.Offset)
	//	}
	//} else {
		if !isMountPoint(ai.mountDir) {
			err = mount(ai.Path, ai.mountDir, ai.Offset)
		}
	//}

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
	if err != nil { return err }

	ai.mountDir = ""

	ai.file.Close()

	// Clean up
	err = os.RemoveAll(ai.TempDir())

	return err
}

func (ai *AppImage) IsMounted() bool {
	if ai.mountDir == "" {
		return false
	}

	return true
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
