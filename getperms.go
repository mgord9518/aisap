package aisap

import (
	"bytes"
	"io/ioutil"
	"strings"
	"strconv"
	"errors"

	helpers  "github.com/mgord9518/aisap/helpers"
	permissions "github.com/mgord9518/aisap/permissions"
	profiles "github.com/mgord9518/aisap/profiles"
	ini	     "gopkg.in/ini.v1"
)

// GetPerms attemps to read permissions from a provided desktop entry, if
// fail, it will return an error and empty *permissions.AppImagePerms
func getPermsFromEntry(entryFile string) (*permissions.AppImagePerms, error) {
	aiPerms := permissions.AppImagePerms{}

	if !helpers.FileExists(entryFile) {
		return nil, errors.New("failed to find requested desktop entry! ("+entryFile+")")
	}

	e, err := ioutil.ReadFile(entryFile)
	if err != nil { return nil, err }

	// Replace ';' with fullwidth semicolon so that the ini package doesn't consider it a break
	e = bytes.ReplaceAll(e, []byte(";"), []byte("；"))

	entry, err := ini.Load(e)
	if err != nil { return nil, err }

	aiPerms, err = loadPerms(aiPerms, entry)

	if err != nil {
		aiPerms.Level = -1
	}

	return &aiPerms, err
}

// Attempt to fetch permissions from the AppImage itself, fall back on aisap
// internal permissinos library
func getPermsFromAppImage(ai *AppImage) (*permissions.AppImagePerms, error) {
	var err error
	var present bool

	aiPerms := permissions.AppImagePerms{}

	// Use the aisap internal profile if it exists
	// If not, set its level as invalid
	if aiPerms, present = profiles.Profiles[strings.ToLower(ai.Name)]; present {
		return &aiPerms, nil
	} else {
		aiPerms.Level = -1
	}

	aiPerms, err = loadPerms(aiPerms, ai.Desktop)
	if err != nil {return &aiPerms, err}

	return &aiPerms, err
}

// Load permissions from INI
func loadPerms(p permissions.AppImagePerms, f *ini.File) (permissions.AppImagePerms, error) {
	var err error

	// Get permissions from keys
	level       := f.Section("X-AppImage-Required-Permissions").Key("Level").Value()
	filePerms   := f.Section("X-AppImage-Required-Permissions").Key("Files").Value()
	devicePerms := f.Section("X-AppImage-Required-Permissions").Key("Devices").Value()
	socketPerms := f.Section("X-AppImage-Required-Permissions").Key("Sockets").Value()

	if level != "" {
		l, err := strconv.Atoi(level)

		if err != nil || l < 0 || l > 3 {
			p.Level = -1
			return p, errors.New("invalid permissions level (must be 0-3)")
		} else {
			p.Level = l
		}
	} else {
		p.Level = -1
		return p, errors.New("profile does not have required flag `Level` under section [X-AppImage-Required-Permissions]")
	}

	p.Files = helpers.SplitKey(filePerms)
	p.Devices = helpers.SplitKey(devicePerms)
	p.Sockets = helpers.SplitKey(socketPerms)

	// Assume readonly if unspecified
	for i := range(p.Files) {
		ex := p.Files[i][len(p.Files[i])-3:]

		if len(strings.Split(p.Files[i], ":")) < 2 ||
		ex != ":ro" && ex != ":rw" {
			p.Files[i] = p.Files[i]+":ro"
		}
	}

	// Convert devices to shorthand if not already
	for i, val := range(p.Devices) {
		if len(val) > 5 && val[0:5] == "/dev/" {
			p.Devices[i] = strings.Replace(val, "/dev/", "", 1)
		}
	}

	return p, err
}
