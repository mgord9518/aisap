package aisap

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"strconv"
	"errors"

	helpers     "github.com/mgord9518/aisap/helpers"
	permissions "github.com/mgord9518/aisap/permissions"
	profiles     "github.com/mgord9518/aisap/profiles"
	ini         "gopkg.in/ini.v1"
)

// GetPerms attemps to read permissions from a provided desktop entry, if
// fail, it will return an error and empty *permissions.AppImagePerms
func getPermsFromEntry(r io.Reader) (*permissions.AppImagePerms, error) {
	aiPerms := &permissions.AppImagePerms{}

	e, err := ioutil.ReadAll(r)
	if err != nil { return nil, err }

	// Replace ';' with fullwidth semicolon so that the ini package doesn't consider it a break
	e = bytes.ReplaceAll(e, []byte(";"), []byte("；"))

	entry, err := ini.Load(e)
	if err != nil { return nil, err }

	return loadPerms(aiPerms, entry)
}

// Attempt to fetch permissions from the AppImage itself, fall back on aisap
// internal permissinos library
func getPermsFromAppImage(ai *AppImage) (*permissions.AppImagePerms, error) {
	// Use the aisap internal profile if it exists
	// If not, set its level as invalid
	if aiPerms, present := profiles.Profiles[strings.ToLower(ai.Name)]; present {
		return &aiPerms, nil
	} else {
		return loadPerms(&aiPerms, ai.Desktop)
	}

	return nil, errors.New("how the fuck did this happen?")
}

// Load permissions from INI
func loadPerms(p *permissions.AppImagePerms, f *ini.File) (*permissions.AppImagePerms, error) {
	// Get permissions from keys
	level       := f.Section("X-AppImage-Required-Permissions").Key("Level").Value()
	filePerms    := f.Section("X-AppImage-Required-Permissions").Key("Files").Value()
	devicePerms := f.Section("X-AppImage-Required-Permissions").Key("Devices").Value()
	socketPerms := f.Section("X-AppImage-Required-Permissions").Key("Sockets").Value()

	if level == "" {
		p.Level = -1
		return p, errors.New("profile does not have required flag `Level` under section [X-AppImage-Required-Permissions]")
	}

	l, err := strconv.Atoi(level)
	if err != nil || l < 0 || l > 3 {
		p.Level = -1
		return p, errors.New("invalid permissions level (must be 0-3)")
	} else {
		p.Level = l
	}

	// Split string into slice and clean up the names
	p.Files   = cleanFiles(helpers.SplitKey(filePerms))
	p.Devices = cleanDevices(helpers.SplitKey(devicePerms))
	p.Sockets = helpers.SplitKey(socketPerms)

	return p, nil
}
