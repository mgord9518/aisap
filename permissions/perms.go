package permissions

import (
	"bufio"
	"bytes"
	"errors"
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

var (
    InvalidSocket = errors.New("socket invalid")
)

type File struct {
	Source   string
	Dest     string
	Writable bool
}

type Socket int

const (
    x11 = Socket(0)
	alsa
	audio
	pulseaudio
	wayland
	dbus
	cgroup
	network
	pid
	pipewire
	session
	user
	uts
)

type AppImagePerms struct {
	Level        int    `json:"level"`       // How much access to system files
	Files      []string `json:"filesystem"`  // Grant permission to access files
	Devices    []string `json:"devices"`     // Access device files (eg: dri, input)
	Sockets    []string `json:"sockets"`     // Use sockets (eg: x11, pulseaudio, network)

	// TODO: rename to PersistentHome or something
	DataDir    bool     `json:"data_dir"` // Whether or not a data dir should be created (only
	// use if the AppImage saves ZERO data eg: 100% online or a game without
	// save files)

	// Only intended for unmarshalling, should not be used for other purposes
	Names []string `json:"names"` 
}

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
		p.DataDir = false
	}

	// Phasing out negative bools, I will eventually replace `NoDataDir`
	// in favor of `DataDir`
	if e.Section("X-App Permissions").Key("DataDir").Value() == "false" {
		p.DataDir = true
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

func IsSocketValid(socket string) bool {
    _, present := helpers.Contains(ValidSockets(), socket)

    return present
}

func ValidSockets() []string {
    return []string{ "x11", "alsa", "audio", "pulseaudio", "wayland", "dbus", "cgroup", "network", "pid", "pipewire", "session", "user", "uts" }
}

func (p *AppImagePerms) AddFiles(s ...string) {
	// Remove previous files of the same name if they exist
	p.RemoveFiles(s...)

	p.Files = append(p.Files, helpers.CleanFiles(s)...)
}

func (p *AppImagePerms) AddDevices(s ...string) {
	p.RemoveDevices(s...)

	p.Devices = append(p.Devices, helpers.CleanDevices(s)...)
}

func (p *AppImagePerms) AddSockets(s ...string) error {
	if len(s) == 0 { return nil}

	p.RemoveSockets(s...)

	for i := range(s) {
		if IsSocketValid(s[i]) {
			p.Sockets = append(p.Sockets, s[i])
			return nil
		}
	}

	return InvalidSocket
}

func (p *AppImagePerms) removeFile(str string) {
	// Done this way to ensure there is an `extension` eg: `:ro` on the string,
	// it will then be used to detect if that file already exists
	str = helpers.CleanFiles([]string{str})[0]
	s  := strings.Split(str, ":")
	str = strings.Join(s[:len(s)-1], ":")

	if i, present := helpers.ContainsAny(p.Files,
	[]string{ str + ":ro", str + ":rw" }); present {
		p.Files = append(p.Files[:i], p.Files[i+1:]...)
	}
}

func (p *AppImagePerms) RemoveFiles(s ...string) {
	for i := range(s) {
		p.removeFile(s[i])
	}
}

func (p *AppImagePerms) removeDevice(str string) {
	if i, present := helpers.Contains(p.Devices, str); present {
		p.Devices = append(p.Devices[:i], p.Devices[i+1:]...)
	}
}

func (p *AppImagePerms) RemoveDevices(s ...string) {
	for i := range(s) {
		p.removeDevice(s[i])
	}
}

func (p *AppImagePerms) removeSocket(str string) {
	if i, present := helpers.Contains(p.Sockets, str); present {
		p.Sockets = append(p.Sockets[:i], p.Sockets[i+1:]...)
	}
}

func (p *AppImagePerms) RemoveSockets(s ...string) {
	for i := range(s) {
		p.removeSocket(s[i])
	}
}

// Set sandbox base permission level
func (p *AppImagePerms) SetLevel(l int) error {
	if l < 0 || l > 3 {
		return errors.New("permissions level must be int from 0-3")
	}

	p.Level = l

	return nil
}
