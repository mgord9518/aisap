// Drop-in replacemnt for go-appimage for sandboxing and use with shappimages
// NOT FINISHED AND STILL LACKING BASIC FEATURES
// THIS SHOULD BE USED FOR TESTING PURPOSES *ONLY* UNTIL IN A STABLE STATE

package aisap

import (
	"bytes"
	"errors"
	"path/filepath"
	"time"
	"io"
	"os"
	"strings"
	"io/ioutil"

	ini      "gopkg.in/ini.v1"
	helpers  "github.com/mgord9518/aisap/helpers"
	profiles "github.com/mgord9518/aisap/profiles"
	imgconv  "github.com/mgord9518/imgconv"
)

type AppImage struct {
	Desktop  *ini.File               // INI of internal desktop entry
	Perms    *profiles.AppImagePerms // Permissions
	Path      string                 // Location of AppImage
	tempDir   string                 // The AppImage's `/tmp` directory
	mountDir  string                 // The location the AppImage is mounted at
	runId     string                 // Random string associated with this specific run instance
	Name      string                 // AppImage name from the desktop entry 
	Version   string                 // Version of the AppImage
	Offset    int                    // Offset of SquashFS image
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

	e, err := ioutil.ReadFile(fp[0])
	entry, _ := ini.Load(e)

	ai.Desktop  = entry
	ai.Name     = entry.Section("Desktop Entry").Key("Name").Value()
	ai.Version  = entry.Section("Desktop Entry").Key("X-AppImage-Version").Value()

	if ai.Version == "" {
		ai.Version = "1.0"
	}

	ai.Perms, _ = getPermsFromAppImage(ai)

    return ai, err
}

// Return a reader for the `.DirIcon` file of the AppImage, converting it to
// PNG if it's in SVG or XPM format
func (ai AppImage) Thumbnail() (io.Reader, error) {
	var f io.Reader

	f, err = os.Open(filepath.Join(ai.mountDir, ".DirIcon"))
	if err != nil { return nil, err }

	// Get the file's magic number
	id := make([]byte, 4)
	io.ReadAtLeast(f, id, 4)

	// Convert `.DirIcon` to PNG format if it isn't already
	// Note: the only other officially supported formats for AppImage are XPM
	// and SVG
	if id[0] != 0x89 || id[1] != 'P' ||
	   id[2] != 'N'  || id[3] != 'G' {
		// Recombine the file's magic number with the rest of the reader
		f = io.MultiReader(bytes.NewReader(id), f)
		f, err = imgconv.ConvertWithAspect(f, 256, "png")
	}

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

func (ai AppImage) AddFiles(s []string) {
	// Add `:ro` if the file doesn't specify
	for i := range(s) {
		// Get the last 3 chars of the file entry
		ex := s[i][len(s[i])-3:]

		if len(strings.Split(s[i], ":")) < 2 ||
		ex != ":ro" && ex != ":rw" {
			s[i] = s[i]+":ro"
		}
	}

	ai.Perms.Files = append(ai.Perms.Files, s...)
}

func (ai AppImage) AddDevices(s []string) {
	ai.Perms.Devices = append(ai.Perms.Devices, s...)
}

func (ai AppImage) AddSockets(s []string) {
	ai.Perms.Sockets = append(ai.Perms.Sockets, s...)
}

func (ai AppImage) AddShare(s []string) {
	ai.Perms.Share = append(ai.Perms.Share, s...)
}

func (ai AppImage) SetPerms(entryFile string) error {
	nPerms, err := getPermsFromEntry(entryFile)
	*ai.Perms = *nPerms

	return err
}

func (ai AppImage) SetRootDir(d string) {
	rootDir = d
}

func (ai AppImage) SetDataDir(d string) {
	dataDir = d
}

func (ai AppImage) SetTempDir(d string) {
	tempDir = d
}

// Currently only works with type 2 AppImages, so return 2 as placeholder
func (ai AppImage) Type() int {
	return 2
}

// TODO: preserve file permissions
func (ai AppImage) ExtractFile(path string, dest string, resolveSymlinks bool) error {
	path = filepath.Join(ai.MountDir(), path)

	// Remove file if it already exists
	os.Remove(filepath.Join(dest))
	info, err := os.Lstat(path)

	// True if file is symlink and `resolveSymlinks` is false
	if info != nil && !resolveSymlinks &&
	info.Mode()&os.ModeSymlink == os.ModeSymlink {
		target, _ := os.Readlink(path)
		err = os.Symlink(target, dest)
	} else {
		inF, err := os.Open(path)
		defer inF.Close()
		if err != nil { return err }

		outF, err := os.Create(dest)
		defer outF.Close()
		if err != nil { return err }

		_, err = io.Copy(outF, inF)
		if err != nil { return err }
	}

	return err
}

func (ai AppImage) Icon() (io.ReadCloser, string, error) {
	if ai.Desktop == nil {
		return nil, "", errors.New("desktop file wasn't parsed")
	}

	iconf := ai.Desktop.Section("Desktop Entry").Key("Icon").Value()

	if iconf == "" {
		return nil, "", errors.New("desktop file doesn't specify an icon")
	}

	// If the desktop entry specifies an extension, use it
	if strings.HasSuffix(iconf, ".png") || strings.HasSuffix(iconf, ".svg") {
		r, err := os.Open(ai.mountDir+"/"+iconf)
		return r, iconf, err
	}

	// If not, iterate through all AppImage specified formats
	fp, err := filepath.Glob(ai.mountDir+"/"+iconf+"*")
	if err != nil { return nil, "", err }

	for _, v := range(fp) {
		if strings.HasSuffix(v, ".png") || strings.HasSuffix(v, ".svg") {
			r, err := os.Open(v)

			return r, v, err
		}
	}

	return nil, "", errors.New("unable to find icon with valid extension (.png, .svg) inside AppImage")
}
