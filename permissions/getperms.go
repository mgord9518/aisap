package permissions

import (
	"strconv"

	helpers     "github.com/mgord9518/aisap/helpers"
	ini         "gopkg.in/ini.v1"
)

// FromIni attemps to read permissions from a provided *ini.File, if fail, it
// will return an *AppImagePerms with a `Level` value of -1
func FromIni(e *ini.File) *AppImagePerms {
	aiPerms := &AppImagePerms{}

	// Get permissions from keys
	level       := e.Section("X-AppImage-Required-Permissions").Key("Level").Value()
	filePerms   := e.Section("X-AppImage-Required-Permissions").Key("Files").Value()
	devicePerms := e.Section("X-AppImage-Required-Permissions").Key("Devices").Value()
	socketPerms := e.Section("X-AppImage-Required-Permissions").Key("Sockets").Value()

	l, err := strconv.Atoi(level)
	if err != nil || l < 0 || l > 3 {
		aiPerms.Level = -1
	} else {
		aiPerms.Level = l
	}

	// Split string into slice and clean up the names
	aiPerms.Files   = helpers.CleanFiles(helpers.SplitKey(filePerms))
	aiPerms.Devices = helpers.CleanDevices(helpers.SplitKey(devicePerms))
	aiPerms.Sockets = helpers.SplitKey(socketPerms)

	return aiPerms
}
