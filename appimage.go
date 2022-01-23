// Drop-in replacemnt for go-appimage for sandboxing and use with shappimages
// NOT FINISHED AND STILL LACKING BASIC FEATURES
// THIS SHOULD BE USED FOR TESTING PURPOSES *ONLY* UNTIL IN A STABLE STATE

package aisap

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	ini         "gopkg.in/ini.v1"
	helpers     "github.com/mgord9518/aisap/helpers"
	profiles    "github.com/mgord9518/aisap/profiles"
	permissions "github.com/mgord9518/aisap/permissions"
	imgconv     "github.com/mgord9518/imgconv"
)

var (
	usern string
	homed string

	sysTemp   string
	mnt      *exec.Cmd

	err     error
	present bool
)

type AppImage struct {
	Desktop     *ini.File                  // INI of internal desktop entry
	Perms       *permissions.AppImagePerms // Permissions
	Path         string // Location of AppImage
	dataDir      string // The AppImage's `~` directory
	rootDir      string // Can be used to give the AppImage fake system files
	tempDir      string // The AppImage's `/tmp` directory
	mountDir     string // The location the AppImage is mounted at
	runId        string // Random string associated with this specific run instance
	Name         string // AppImage name from the desktop entry 
	Version      string // Version of the AppImage
	UpdateInfo   string // Update information
	Offset       int    // Offset of SquashFS image
	imageType    int    // Type of AppImage (either 1 or 2)
}

func init() {
	usr, _ := user.Current()
	usern   = usr.Username
	homed   = filepath.Join("/home", usern)
	uid := strconv.Itoa(os.Getuid())

	sysTemp, present = os.LookupEnv("XDG_RUNTIME_DIR")
	if !present {
		sysTemp = filepath.Join("/run", "user", uid)
	}
}

// Create a new AppImage object from a path
func NewAppImage(src string) (*AppImage, error) {
	if !helpers.FileExists(src) {
		return nil, errors.New("file not found!")
	}

	ai := &AppImage{}
	ai.Path = src

	// Set the runId, tempDir and rootDir of the AppImage
	pfx := path.Base(ai.Path)[0:6]
	ai.runId = pfx + helpers.RandString(int(time.Now().UTC().UnixNano()), 6)

	ai.tempDir, err = helpers.MakeTemp(filepath.Join(sysTemp, "aisap"), ai.runId)
	if err != nil { return nil, err }
	ai.rootDir = "/"

	ai.mountDir, err = helpers.MakeTemp(ai.tempDir, ".mount_" + ai.runId)

	ai.Offset, err = helpers.GetOffset(src)
	if err != nil { return nil, err }

	err = mount(src, ai.mountDir, ai.Offset)
	if err != nil { return nil, err }

	// Return all `.desktop` files. A vadid AppImage should only have one
	fp, err := filepath.Glob(ai.mountDir + "/*.desktop")
	if err != nil { return nil, err }
	f, err := os.Open(fp[0])
	defer f.Close()

	// Replace normal semicolons with fullwidth semicolons so that it doen't
	// interfere with the INI parsing
	var e string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		e = e + strings.ReplaceAll(scanner.Text(), ";", "；") + "\n"
	}
	entry, _ := ini.Load([]byte(e))

	ai.Desktop = entry
	ai.Name    = entry.Section("Desktop Entry").Key("Name").Value()
	ai.Version = entry.Section("Desktop Entry").Key("X-AppImage-Version").Value()

	ai.UpdateInfo, _ = helpers.ReadUpdateInfo(ai.Path)

	if ai.Version == "" {
		ai.Version = "1.0"
	}

	// Prefer level from aisap internal library, if none exist, attempt to
	// retrieve profile from the desktop entry itself. If neither exists,
	// `ai.Perms.Level` will be set to -1 and should be regarded as invalid
	ai.Perms = profiles.FromName(ai.Name)
	if ai.Perms.Level == -1 {
		*ai.Perms = *permissions.FromIni(ai.Desktop)
	}

	return ai, err
}

// Return a reader for the `.DirIcon` file of the AppImage, converting it to
// PNG if it's in SVG or XPM format
func (ai AppImage) Thumbnail() (io.Reader, error) {
	f, err := os.Open(filepath.Join(ai.mountDir, ".DirIcon"))
	if err != nil { return nil, err }

	// Convert `.DirIcon` to PNG format if it isn't already
	// Note: the only other officially supported formats for AppImage are XPM
	// and SVG
	if !helpers.HasMagic(f, "\x89PNG", 0) {
		f.Seek(0, io.SeekStart)
		return imgconv.ConvertWithAspect(f, 256, "png")
	}

	f.Seek(0, io.SeekStart)

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

// TODO: Move these functions to `ai.Perms` because it makes more sense
func (ai AppImage) AddFile(str string) {
	// Clear out previous file if it already exists
	ai.RemoveFile(str)

	// TODO: Add clean functions for strings instead of slices to helpers, use
	// them here instead
	s := []string{str}
	ai.Perms.Files = append(ai.Perms.Files, helpers.CleanFiles(s)...)
}

func (ai AppImage) AddFiles(s []string) {
	// Remove previous files of the same name if they exist
	ai.RemoveFiles(s)

	ai.Perms.Files = append(ai.Perms.Files, helpers.CleanFiles(s)...)
}

// TODO: Create AddDevice and AddSocke tfor indiviual strings
func (ai AppImage) AddDevices(s []string) {
	ai.Perms.Devices = append(ai.Perms.Devices, helpers.CleanDevices(s)...)
}

func (ai AppImage) AddSockets(s []string) {
	ai.Perms.Sockets = append(ai.Perms.Sockets, s...)
}

func (ai AppImage) RemoveFile(str string) {
	// Done this way to ensure there is an `extension` eg: `:ro` on the string,
	// it will then be used to detect if that file already exists
	str = helpers.CleanFiles([]string{str})[0]
	s  := strings.Split(str, ":")
	str = strings.Join(s[:len(s)-1], ":")

	if i, present := helpers.ContainsAny(ai.Perms.Files,
	[]string{ str + ":ro", str + ":rw" }); present {
		ai.Perms.Files = append(ai.Perms.Files[:i], ai.Perms.Files[i+1:]...)
	}
}

func (ai AppImage) RemoveFiles(s []string) {
	for i := range(s) {
		ai.RemoveFile(s[i])
	}
}

func (ai AppImage) RemoveDevice(str string) {
	if i, present := helpers.Contains(ai.Perms.Devices, str); present {
		ai.Perms.Devices = append(ai.Perms.Devices[:i], ai.Perms.Devices[i+1:]...)
	}
}

func (ai AppImage) RemoveDevices(s []string) {
	for i := range(s) {
		ai.RemoveDevice(s[i])
	}
}

func (ai AppImage) RemoveSocket(str string) {
	if i, present := helpers.Contains(ai.Perms.Sockets, str); present {
		ai.Perms.Sockets = append(ai.Perms.Sockets[:i], ai.Perms.Sockets[i+1:]...)
	}
}

func (ai AppImage) RemoveSockets(s []string) {
	for i := range(s) {
		ai.RemoveSocket(s[i])
	}
}

func (ai AppImage) SetPerms(entryFile string) error {
	r, err := os.Open(entryFile)
	if err != nil { return err }

	e, err := ioutil.ReadAll(r)
	if err != nil { return err }

	e = bytes.ReplaceAll(e, []byte(";"), []byte("；"))

	entry, err := ini.Load(e)
	if err != nil { return err }

	*ai.Perms = *permissions.FromIni(entry)

	return err
}

// Set the directory the sandbox pulls system files from
func (ai AppImage) SetRootDir(d string) {
	ai.rootDir = d
}

// Set the directory for the sandboxed AppImage's `HOME`
func (ai AppImage) SetDataDir(d string) {
	ai.dataDir = d
}

// Set the directory for the sandboxed AppImage's `TMPDIR`
func (ai AppImage) SetTempDir(d string) {
	ai.tempDir = d
}

// Set sandbox base permission level
func (ai AppImage) SetLevel(l int) error {
	if l < 0 || l > 3 {
		return errors.New("permissions level must be int from 0-3")
	}

	ai.Perms.Level = l

	return nil
}

// Return type of AppImage
func (ai AppImage) Type() int {
	t, _ := helpers.GetAppImageType(ai.Path)

	return t
}

// Extract a file from the AppImage's interal SquashFS image
func (ai AppImage) ExtractFile(path string, dest string, resolveSymlinks bool) error {
	path = filepath.Join(ai.mountDir, path)

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

		info, err := os.Stat(path)
		perms := info.Mode().Perm()

		outF, err := os.Create(dest)
		defer outF.Close()
		if err != nil { return err }

		err = os.Chmod(dest, perms)
		if err != nil { return err }

		_, err = io.Copy(outF, inF)
		if err != nil { return err }
	}

	return err
}

// Like `ExtractFile()` but gives access to the reader instead of extracting
func (ai AppImage) ExtractFileReader(path string) (io.ReadCloser, error) {
	path = filepath.Join(ai.mountDir, path)

	return os.Open(path)
}

// Returns the icon reader of the AppImage, valid formats are SVG and PNG
func (ai AppImage) Icon() (io.ReadCloser, string, error) {
	if ai.Desktop == nil {
		return nil, "", errors.New("desktop file wasn't parsed")
	}

	// Return error if desktop file has no icon
	iconf := ai.Desktop.Section("Desktop Entry").Key("Icon").Value()
	if iconf == "" {
		return nil, "", errors.New("desktop file doesn't specify an icon")
	}

	// If the desktop entry specifies an extension, use it
	if strings.HasSuffix(iconf, ".png") || strings.HasSuffix(iconf, ".svg") {
		r, err := os.Open(filepath.Join(ai.mountDir, iconf))
		return r, iconf, err
	}

	// If not, iterate through all AppImage specified formats
	fp, err := filepath.Glob(filepath.Join(ai.mountDir, iconf) + "*")
	if err != nil { return nil, "", err }

	for _, v := range(fp) {
		if strings.HasSuffix(v, ".png") || strings.HasSuffix(v, ".svg") {
			r, err := os.Open(v)

			return r, path.Base(v), err
		}
	}

	return nil, "", errors.New("unable to find icon with valid extension (.png, .svg) inside AppImage")
}
