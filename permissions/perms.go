package permissions

import (
	"errors"
	"strings"

	helpers "github.com/mgord9518/aisap/helpers"
)

type AppImagePerms struct {
	Level        int    `json:"level"`       // How much access to system files
	Files      []string `json:"filesystem"`  // Grant permission to access files
	Devices    []string `json:"devices"`     // Access device files (eg: dri, input)
	Sockets    []string `json:"sockets"`     // Use sockets (eg: x11, pulseaudio, network)
	NoDataDir    bool   `json:"no_data_dir"` // Whether or not a data dir should be created (only
	// use if the AppImage saves ZERO data eg: 100% online or a game without
	// save files)
}

var (
    InvalidSocket = errors.New("socket invalid")
)

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
	p.RemoveSockets(s...)

	for i := range(s) {
		if IsSocketValid(str) {
			p.Sockets = append(p.Sockets, str)
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
