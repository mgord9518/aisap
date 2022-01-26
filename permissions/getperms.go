package permissions

import (
	"bytes"
	"io"
	"io/ioutil"
	"strconv"

	helpers     "github.com/mgord9518/aisap/helpers"
	ini         "gopkg.in/ini.v1"
)

// FromIni attemps to read permissions from a provided *ini.File, if fail, it
// will return an *AppImagePerms with a `Level` value of -1
func FromIni(e *ini.File) (*AppImagePerms, error) {
	p := &AppImagePerms{}

	// Get permissions from keys
	level       := e.Section("X-AppImage-Required-Permissions").Key("Level").Value()
	filePerms   := e.Section("X-AppImage-Required-Permissions").Key("Files").Value()
	devicePerms := e.Section("X-AppImage-Required-Permissions").Key("Devices").Value()
	socketPerms := e.Section("X-AppImage-Required-Permissions").Key("Sockets").Value()

	l, err := strconv.Atoi(level)
	if err != nil || l < 0 || l > 3 {
		p.Level = -1
		return p, err
	} else {
		p.Level = l
	}

	// Split string into slices and clean up the names
	p.AddFiles(helpers.SplitKey(filePerms))
	p.AddDevices(helpers.SplitKey(devicePerms))
	p.AddSockets(helpers.SplitKey(socketPerms))
//	p.Files   = helpers.CleanFiles(helpers.SplitKey(filePerms))
//	p.Devices = helpers.CleanDevices(helpers.SplitKey(devicePerms))
//	p.Sockets = helpers.SplitKey(socketPerms)

	return p, nil
}

func FromReader(r io.Reader) (*AppImagePerms, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil { return nil, err }

	b = bytes.ReplaceAll(b, []byte(";"), []byte("；"))
	
	e, err := ini.Load(b)
	if err != nil { return nil, err }

	return FromIni(e)
}
