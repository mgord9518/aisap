package aisap

import (
	"bufio"
	"io"
	"os"
	"debug/elf"
	"os/exec"
	"strings"
	"strconv"
	"errors"
	"encoding/binary"
)

// MountAppImage mounts the requested AppImage (src) to the destination
// directory (dest)
// Quick, hacky implementation, ideally this should be redone using the
// squashfuse library
func MountAppImage(src string, dest string) error {
	var squashfuse string
	var err error

	if squashfuse, err = exec.LookPath("squashfuse"); err != nil {
		squashfuse, err = exec.LookPath("./squashfuse")
		if err != nil {
			return errors.New("Failed to find squashfuse binary! Cannot mount AppImage")
		}
	}

	aiOffset, err := GetAppImageOffset(src)
	if err != nil { return err }
	n := strconv.Itoa(aiOffset)

	mnt = exec.Command(squashfuse, "-o", "offset="+n, src, dest)
	err = mnt.Run()

	return err
}

func UnmountAppImage(ai *AppImage) error {
	return UnmountAppImageFile(ai.MountDir())
}

func UnmountAppImageFile(mntPt string) error {
	var mntCmd string
	var err error

	if mntCmd, err = exec.LookPath("fusermount"); err != nil {
		mntCmd, err = exec.LookPath("mount")
	}

	umount := exec.Command(mntCmd, "-u", mntPt)
	err := umount.Run()

	return err
}

// GetAppImageOffset takes an AppImage as argument, returning its offset as int
// and an error if unsuccessful
func GetAppImageOffset(src string) (int, error) {
	var offset int

	format, err := GetAppImageType(src)
	if format == "shappimage" {
		offset, err = GetShappImageSize(src)
	} else if format == "2" {
		offset, err = GetElfSize(src)
	} else {
		return -1, errors.New("cannot find AppImage offset")
	}

	return offset, err
}

// Takes a src file as argument, returning the size of the shappimage header
// and an error if unsuccessful
// I HAVEN'T PUBLISHED THIS PROJECT YET, that's why these functions are
// undocumented. Probably will be a minute until I do too because I'm pretty
// caught up in aisap
func GetShappImageSize(src string) (int, error) {
	var binLength	int
	var scriptLength int
	var shappLength  int

	format, _ := GetAppImageType(src)
	if format != "shappimage" {
		return -1, errors.New("invalid shappimage; file does not contain shappimage header")
	}

	f, _ := os.Open(src)
	defer f.Close()
	s := bufio.NewScanner(f)
	foundBinLength := false

	for s.Scan() {
		str := s.Text()

		scriptLength += len(str)+1
		b := strings.Split(str, "=")
		if b[0] == "binLength" && !foundBinLength {
			binLength, _ = strconv.Atoi(b[1])
			foundBinLength = true
		}

		// This line signifies the end of the script
		if str == "#___SHAPPIMAGE_BIN_START___#" {
			shappLength = binLength + scriptLength
			return shappLength, nil
		}
	}

	err := errors.New("invalid shappimage; file does not contain shappimage footer")

	return -1, err
}

// Function from <github.com/probonopd/go-appimage/internal/helpers/elfsize.go>
// credit goes to respective author; modified from original
// GetElfSize takes a src file as argument, returning its size as an int
// and an error if unsuccessful
func GetElfSize(src string) (int, error) {
	format, _ := GetAppImageType(src)
	if format != "2" && format != "elf"{
		return -1, errors.New("Invalid ELF file, cannot calculate size")
	}

	f, _ := os.Open(src)
	defer f.Close()
	e, err := elf.NewFile(f)
	if err != nil { return -1, err }

	// Find offsets based on arch
	sr := io.NewSectionReader(f, 0, 1<<63-1)
	var shoff, shentsize, shnum int

	switch e.Class.String() {
	case "ELFCLASS64":
		hdr := new(elf.Header64)
		_, err = sr.Seek(0, 0)
		if err != nil { return -1, err }
		err = binary.Read(sr, e.ByteOrder, hdr)
		if err != nil { return -1, err }

		shoff	  = int(hdr.Shoff)
		shnum	  = int(hdr.Shnum)
		shentsize = int(hdr.Shentsize)
	case "ELFCLASS32":
		hdr := new(elf.Header32)
		_, err = sr.Seek(0, 0)
		if err != nil { return -1, err }
		err := binary.Read(sr, e.ByteOrder, hdr)
		if err != nil { return -1, err }

		shoff	 = int(hdr.Shoff)
		shnum	 = int(hdr.Shnum)
		shentsize = int(hdr.Shentsize)
	default:
		return 0, nil
	}

	elfsize := shoff + (shentsize * shnum)

	return elfsize, nil
}

// Find the type of AppImage
// Returns strings either `1` for ISO disk image AppImage, `2` for type 2
// SquashFS AppImage, or `shappimage` for shell script SquashFS AppImage
func GetAppImageType(src string) (string, error) {
	f, err := os.Open(src)
	defer f.Close()
	if err != nil { return "", err }

	_, err = f.Stat()
	if err != nil { return "", err }

	// Read header of file
	var magic [16]byte
	_, err = io.ReadFull(f, magic[:])
	if err != nil { return "", err }

	if magic[0] == '\x7f' &&
	magic[1] == 'E' && magic[2] == 'L' && magic[3] == 'F' &&
	magic[8] == 'A' && magic[9] == 'I' && magic[10] == '\x01' {
		// AppImage type is type 1 (standard)
		return "1", nil
	} else if magic[0] == '\x7f' &&
	magic[1] == 'E' && magic[2] == 'L' && magic[3] == 'F' &&
	magic[8] == 'A' && magic[9] == 'I' && magic[10] == '\x02' {
		// AppImage type is type 2 (standard)
		return "2", nil
	} else if magic[0] == '\x7f' &&
	magic[1] == 'E' && magic[2] == 'L' && magic[3] == 'F' {
		// Unknown AppImage, but valid ELF
		return "elf", nil
	} else if magic[10] == '#' &&
	magic[11] == 's' && magic[12] == 'h' &&
	magic[13] == 'A' && magic[14] == 'I' {
		// AppImage is shappimage (shell SquashFS implementation)
		return "shappimage", nil
	}

	err = errors.New("Unable to get AppImage type")
	return "", err
}
