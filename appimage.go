// Drop-in replacemnt for go-appimage for sandboxing and use with shappimages
// NOT FINISHED AND STILL LACKING BASIC FEATURES

package aisap

import (
	"path/filepath"
	"time"
	"io"
	"os"
	"io/ioutil"

	ini      "gopkg.in/ini.v1"
	helpers  "github.com/mgord9518/aisap/helpers"
	profiles "github.com/mgord9518/aisap/profiles"
)

type AppImage struct {
	Desktop  *ini.File      // INI of internal desktop entry
	Perms    *profiles.AppImagePerms // Permissions
	Path      string        // Location of AppImage
	tempDir   string        // The AppImage's `/tmp` directory
	mountDir  string        // The location the AppImage is mounted at
	runId     string        // Random string associated with this specific run instance
	Name      string        // AppImage name from the desktop entry 
	Version   string        // Version of the AppImage
	Offset    int           // Offset of SquashFS image
	imageType int
}

func NewAppImage(src string) (*AppImage, error) {
	var err error

    ai := &AppImage{}
    ai.Path = src

	ai.runId = helpers.RandString(int(time.Now().UTC().UnixNano()), 8)
	ai.tempDir, err = helpers.MakeTemp("/tmp", ".aisapTemp_"+ai.RunId())
	if err != nil { return nil, err }

	ai.mountDir, err = helpers.MakeTemp(ai.TempDir(), ".mount_"+ai.RunId())

	err = MountAppImage(src, ai.mountDir)
	if err != nil { return nil, err }

	// Return all `.desktop` files. A vadid AppImage should only have one
	fp, err := filepath.Glob(ai.mountDir+"/*.desktop")
	if err != nil { return nil, err }

	// Load the first (should be only) desktop file
	e, err := ioutil.ReadFile(fp[0])
	entry, _ := ini.Load(e)

	ai.Desktop = entry
	ai.Name    = entry.Section("Desktop Entry").Key("Name").Value()

    return ai, err
}

func (ai AppImage) Thumbnail() (io.ReadCloser, error) {
	f, err := os.Open(ai.mountDir+"/.DirIcon")

	return f, err
}

func (ai AppImage) TempDir() string {
	return ai.tempDir
}

func (ai AppImage) MountDir() string {
	return ai.mountDir
}

func (ai AppImage) RunId() string {
	return ai.runId
}
