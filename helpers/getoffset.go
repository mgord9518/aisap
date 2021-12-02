package helpers

import (
	"io"
	"os"
	"debug/elf"
	"errors"
	"encoding/binary"
)

// GetOffset takes an AppImage (either ELF or shappimage), returning the offset
// of its SquashFS archive
func GetOffset(src string) (int, error) {
	var offset int

	format, err := GetAppImageType(src)
	if format == -2 {
		offset, err = getShappImageSize(src)
	} else if format == 2 || format == 0 {
		offset, err = getElfSize(src)
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
func getShappImageSize(src string) (int, error) {
	var magic [17]byte

	f, err := os.Open(src)
	defer f.Close()
	if err != nil { return -1, err }

	_, err = f.Stat()
	if err != nil { return -1, err }

	_, err = io.ReadFull(f, magic[:])
	if err != nil { return -1, err }

	offset := binary.BigEndian.Uint16([]byte{ magic[15], magic[16] })

	return int(offset), nil
}

// Function from <github.com/probonopd/go-appimage/internal/helpers/elfsize.go>
// credit goes to respective author; modified from original
// getElfSize takes a src file as argument, returning its size as an int
// and an error if unsuccessful
func getElfSize(src string) (int, error) {
	format, _ := GetAppImageType(src)
	if format != 2 && format != 0 {
		return -1, errors.New("invalid ELF file, cannot calculate size")
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
// SquashFS AppImage, `0` for unknown valid ELF or `-2` for shell script
// SquashFS AppImage (shappimage)
func GetAppImageType(src string) (int, error) {
	f, err := os.Open(src)
	defer f.Close()
	if err != nil { return -1, err }

	_, err = f.Stat()
	if err != nil { return -1, err }

	// Read header of file
	var magic [16]byte
	_, err = io.ReadFull(f, magic[:])
	if err != nil { return -1, err }

	if magic[0] == '\x7f' &&
	magic[1] == 'E' && magic[2] == 'L' && magic[3] == 'F' &&
	magic[8] == 'A' && magic[9] == 'I' && magic[10] == '\x01' {
		// AppImage type is type 1 (standard)
		return 1, nil
	} else if magic[0] == '\x7f' &&
	magic[1] == 'E' && magic[2] == 'L' && magic[3] == 'F' &&
	magic[8] == 'A' && magic[9] == 'I' && magic[10] == '\x02' {
		// AppImage type is type 2 (standard)
		return 2, nil
	} else if magic[0] == '\x7f' &&
	magic[1] == 'E' && magic[2] == 'L' && magic[3] == 'F' {
		// Unknown AppImage, but valid ELF
		return 0, nil
	} else if magic[10] == '#' &&
	magic[11] == 's' && magic[12] == 'h' &&
	magic[13] == 'A' && magic[14] == 'I' {
		// AppImage is shappimage (shell script SquashFS implementation)
		return -2, nil
	}

	err = errors.New("unable to get AppImage type")
	return -1, err
}
