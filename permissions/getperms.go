package permissions

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"strconv"

	helpers "github.com/mgord9518/aisap/helpers"
	ini     "gopkg.in/ini.v1"
	xdg     "github.com/adrg/xdg"
)

// FromIni attempts to read permissions from a provided *ini.File, if fail, it
// will return an *AppImagePerms with a `Level` value of -1 and and error
func FromIni(e *ini.File) (*AppImagePerms, error) {
	p := &AppImagePerms{}

	// Get permissions from keys
	level       := e.Section("X-App Permissions").Key("Level").Value()
	filePerms   := e.Section("X-App Permissions").Key("Files").Value()
	devicePerms := e.Section("X-App Permissions").Key("Devices").Value()
	socketPerms := e.Section("X-App Permissions").Key("Sockets").Value()

	// Enable saving to a data dir by default. If NoDataDir is true, the AppImage
	// HOME dir will be in RAM and non-persistent.
	if e.Section("X-App Permissions").Key("NoDataDir").Value() == "true" {
		p.NoDataDir = true
	}

	// Phasing out negative bools, I will eventually replace `NoDataDir`
	// in favor of `DataDir`
	if e.Section("X-App Permissions").Key("DataDir").Value() == "false" {
		p.NoDataDir = false
	}

	l, err := strconv.Atoi(level)
	if err != nil || l < 0 || l > 3 {
		p.Level = -1
		return p, err
	} else {
		p.Level = l
	}

	// Split string into slices and clean up the names
	p.AddFiles(helpers.SplitKey(filePerms)...)
	p.AddDevices(helpers.SplitKey(devicePerms)...)
	p.AddSockets(helpers.SplitKey(socketPerms)...)

	return p, nil
}

// FromSystem attempts to read permissions from a provided desktop entry at
// ~/.local/share/aisap/profiles/[ai.Name]
// This should be the preferred way to get permissions and gives maximum power
// to the user (provided they use a tool to easily edit these permissions, which
// I'm also planning on making)
func FromSystem(name string) (*AppImagePerms, error) {
	p := &AppImagePerms{}
	var e string

	fp := filepath.Join(xdg.DataHome, "aisap", "profiles", name)
	f, err := os.Open(fp)
	if err != nil {
		return p, err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		e = e + strings.ReplaceAll(scanner.Text(), ";", "；") + "\n"
	}

	entry, err := ini.Load([]byte(e))
	if err != nil {
		return p, err
	}

	p, err = FromIni(entry)

	return p, err
}

func FromReader(r io.Reader) (*AppImagePerms, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil { return nil, err }

	b = bytes.ReplaceAll(b, []byte(";"), []byte("；"))
	
	e, err := ini.Load(b)
	if err != nil { return nil, err }

	return FromIni(e)
}
