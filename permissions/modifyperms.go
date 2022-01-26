package permissions

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"strings"

	ini     "gopkg.in/ini.v1"
	helpers "github.com/mgord9518/aisap/helpers"
)

func (p *AppImagePerms) AddFile(str string) {
	// Clear out previous file if it already exists
	p.RemoveFile(str)

	p.Files = append(p.Files, helpers.CleanFile(str))
}

func (p *AppImagePerms) AddFiles(s []string) {
	// Remove previous files of the same name if they exist
	p.RemoveFiles(s)

	p.Files = append(p.Files, helpers.CleanFiles(s)...)
}

func (p *AppImagePerms) AddDevice(str string) {
	p.RemoveDevice(str)

	p.Devices = append(p.Devices, helpers.CleanDevice(str))
}

func (p *AppImagePerms) AddDevices(s []string) {
	p.RemoveDevices(s)

	p.Devices = append(p.Devices, helpers.CleanDevices(s)...)
}

func (p *AppImagePerms) AddSocket(str string) {
	p.RemoveSocket(str)

	p.Sockets = append(p.Sockets, str)
}

func (p *AppImagePerms) AddSockets(s []string) {
	p.RemoveSockets(s)

	p.Sockets = append(p.Sockets, s...)
}

func (p *AppImagePerms) RemoveFile(str string) {
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

func (p *AppImagePerms) RemoveFiles(s []string) {
	for i := range(s) {
		p.RemoveFile(s[i])
	}
}

func (p *AppImagePerms) RemoveDevice(str string) {
	if i, present := helpers.Contains(p.Devices, str); present {
		p.Devices = append(p.Devices[:i], p.Devices[i+1:]...)
	}
}

func (p *AppImagePerms) RemoveDevices(s []string) {
	for i := range(s) {
		p.RemoveDevice(s[i])
	}
}

func (p *AppImagePerms) RemoveSocket(str string) {
	if i, present := helpers.Contains(p.Sockets, str); present {
		p.Sockets = append(p.Sockets[:i], p.Sockets[i+1:]...)
	}
}

func (p *AppImagePerms) RemoveSockets(s []string) {
	for i := range(s) {
		p.RemoveSocket(s[i])
	}
}

// DEPRICATED, use permissions.FromReader()
func (p *AppImagePerms) SetPerms(entryFile string) error {
	r, err := os.Open(entryFile)
	if err != nil { return err }

	e, err := ioutil.ReadAll(r)
	if err != nil { return err }

	e = bytes.ReplaceAll(e, []byte(";"), []byte("；"))

	entry, err := ini.Load(e)
	if err != nil { return err }

	p2, err := FromIni(entry)
	*p = *p2

	return err
}

// Set sandbox base permission level
func (p *AppImagePerms) SetLevel(l int) error {
	if l < 0 || l > 3 {
		return errors.New("permissions level must be int from 0-3")
	}

	p.Level = l

	return nil
}
