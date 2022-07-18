package helpers

import (
	"bufio"
	"strings"
	"bytes"
	"errors"
	"debug/elf"
)

func ReadUpdateInfo(src string) (string, error) {
	format, err := GetAppImageType(src)
	if err != nil { return "", err }

	if format == 2 || format == 1 {
		return readUpdateInfoFromElf(src)
	} else if format == -2 {
		return readUpdateInfoFromShappimage(src)
	}

	return "", errors.New("AppImage is of unknown type")
}

// Taken and modified from
// <https://github.com/AppImageCrafters/appimage-update/blob/945dfa16017496be7a3f21c827a7ffb11124e548/util/util.go>
func readUpdateInfoFromElf(src string) (string, error) {
	elfFile, err := elf.Open(src)
	if err != nil {
		return "", err
	}

	updInfoSect := elfFile.Section(".upd_info")
	if updInfoSect == nil {
		return "", errors.New("ELF missing .upd_info section")
	}

	sectionData, err := updInfoSect.Data()
	if err != nil {
		return "", errors.New("unable to read update information from section")
	}

	str_end := bytes.Index(sectionData, []byte("\000"))
	if str_end == -1 || str_end == 0 {
		return "", errors.New("no update information found")
	}

	return string(sectionData[:str_end]), nil
}

func readUpdateInfoFromShappimage(src string) (string, error) {
	f, err := ExtractResourceReader(src, "update_info")
	if err != nil { return "", err }

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// Quit on first non-matching line as the update info should only be
		// one line long
		if !strings.Contains(scanner.Text(), " APPIMAGE [update_info]") {
			return scanner.Text(), nil
		}
	}
	
	return "", errors.New("unable to find update information in shImg")
}
