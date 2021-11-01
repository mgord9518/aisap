package aisap

import (
	"bytes"
	"io/ioutil"
	"strings"
	"errors"

	helpers  "github.com/mgord9518/aisap/helpers"
	profiles "github.com/mgord9518/aisap/profiles"
	ini	  "gopkg.in/ini.v1"
	xdg	  "github.com/adrg/xdg"
)

// GetPerms attemps to read permissions from a provided desktop entry, if
// fail, it will return an error and empty *profiles.AppImagePerms
func GetPermsFromEntry(entryFile string, permLevel int) (*profiles.AppImagePerms, error) {
	loadName(permLevel)

	aiPerms := new(profiles.AppImagePerms)

	if !helpers.FileExists(entryFile) {
		return nil, errors.New("failed to find requested desktop entry! ("+entryFile+")")
	}

	e, err := ioutil.ReadFile(entryFile)
	if err != nil { return nil, err }

	// Replace ';' with fullwidth semicolon so that the ini package doesn't consider it a break
	e = bytes.ReplaceAll(e, []byte(";"), []byte("；"))

	entry, err := ini.Load(e)
	if err != nil { return nil, err }

	filePerms   := entry.Section("Desktop Entry").Key("X-AppImage-File-Permissions").Value()
	devicePerms := entry.Section("Desktop Entry").Key("X-AppImage-Device-Permissions").Value()
	socketPerms := entry.Section("Desktop Entry").Key("X-AppImage-Socket-Permissions").Value()
	sharePerms  := entry.Section("Desktop Entry").Key("X-AppImage-Share-Permissions").Value()

	aiPerms.Level	    = permLevel
	aiPerms.DevicePerms = parseDevicePerms(devicePerms, permLevel)
	aiPerms.FilePerms   = parseFilePerms(filePerms,   permLevel)
	aiPerms.SocketPerms = parseSocketPerms(socketPerms, permLevel)
	aiPerms.SharePerms  = parseSharePerms(sharePerms, permLevel)

	if len(preFilePerms) > 0 {
		for i, val := range(preFilePerms) {
			nDir := val[:len(val)-3]
			nDir = strings.Replace(nDir, xdg.Home, "/home/"+usern, 1)
			aiPerms.FilePerms[i] = nDir
		}
	}

	aiPerms.DevicePerms = append(preDevicePerms, aiPerms.DevicePerms...)
	aiPerms.SocketPerms = append(preSocketPerms, aiPerms.SocketPerms...)

	return aiPerms, nil
}

// Attempt to fetch permissions from the AppImage itself, fall back on internal
// permissinos library
func GetPermsFromAppImage(ai *AppImage, permLevel int) (*profiles.AppImagePerms, error) {
	loadName(permLevel)

	aiPerms := new(profiles.AppImagePerms)

	filePerms   := ai.Desktop.Section("Desktop Entry").Key("X-AppImage-File-Permissions").Value()
	devicePerms := ai.Desktop.Section("Desktop Entry").Key("X-AppImage-Device-Permissions").Value()
	socketPerms := ai.Desktop.Section("Desktop Entry").Key("X-AppImage-Socket-Permissions").Value()
	sharePerms  := ai.Desktop.Section("Desktop Entry").Key("X-AppImage-Share-Permissions").Value()

	// If all keys are empty, fall back on internal library
	if filePerms   == "" && devicePerms == "" &&
		socketPerms == "" && sharePerms  == "" {

		*aiPerms = profiles.Profiles[ai.Name]
	} else {
		aiPerms.Level	    = permLevel
		aiPerms.DevicePerms = parseDevicePerms(devicePerms, permLevel)
		aiPerms.FilePerms   = parseFilePerms(filePerms,   permLevel)
		aiPerms.SocketPerms = parseSocketPerms(socketPerms, permLevel)
		aiPerms.SharePerms  = parseSharePerms(sharePerms, permLevel)
	}

	if len(preFilePerms) > 0 {
		for i, val := range(preFilePerms) {
			nDir := val[:len(preFilePerms[i])-3]
			nDir = strings.Replace(nDir, xdg.Home, "/home/"+usern, 1)
			aiPerms.FilePerms[i] = nDir
		}
	}

//	aiPerms.FilePerms = append(preFilePerms, aiPerms.FilePerms...)
	aiPerms.DevicePerms = append(preDevicePerms, aiPerms.DevicePerms...)
	aiPerms.SocketPerms = append(preSocketPerms, aiPerms.SocketPerms...)
	aiPerms.SharePerms  = append(preSharePerms, aiPerms.SharePerms...)


	return aiPerms, nil
}
