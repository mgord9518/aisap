// Drop-in replacemnt for go-appimage for sandboxing and use with shappimages
// NOT FINISHED AND STILL LACKING BASIC FEATURES
// THIS SHOULD BE USED FOR TESTING PURPOSES *ONLY* UNTIL IN A STABLE STATE

package aisap

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"errors"
	"io"
	"path"
	"path/filepath"
	"os"
	"strings"

	ini         "gopkg.in/ini.v1"
	helpers     "github.com/mgord9518/aisap/helpers"
	profiles    "github.com/mgord9518/aisap/profiles"
	permissions "github.com/mgord9518/aisap/permissions"
	imgconv     "github.com/mgord9518/imgconv"
	xdg         "github.com/adrg/xdg"
)

type AppImage struct {
	Desktop     *ini.File                  // INI of internal desktop entry
	Perms       *permissions.AppImagePerms // Permissions
	Path         string // Location of AppImage
	dataDir      string // The AppImage's `~` directory
	rootDir      string // Can be used to give the AppImage fake system files
	tempDir      string // The AppImage's `/tmp` directory
	mountDir     string // The location the AppImage is mounted at
	md5          string // MD5 of AppImage's URI
	runId        string // Random string associated with this specific run instance
	Name         string // AppImage name from the desktop entry 
	Version      string // Version of the AppImage
	UpdateInfo   string // Update information
	Offset       int    // Offset of SquashFS image
	imageType    int    // Type of AppImage (either 1 or 2)
}

// Current version of aisap
const (
	Version = "0.4.6-alpha"
)

// Create a new AppImage object from a path
func NewAppImage(src string) (*AppImage, error) {
	var err error
	ai := &AppImage{Path: src}

	if !helpers.FileExists(ai.Path) {
		return nil, errors.New("file not found!")
	}

	b := md5.Sum([]byte("file://" + ai.Path))
	ai.md5 = fmt.Sprintf("%x", b)

	// Set the runId, tempDir and rootDir of the AppImage
	pfx := path.Base(ai.Path)
	if len(pfx) > 6 {
		pfx = pfx[0:6]
	}

	// Now use a chunk of the MD5sum as seed instead of date
	seed := int(b[0] + b[1] + b[2] + b[3] + b[4] + b[5] + b[6] + b[7] + b[8])
	ai.runId = pfx + helpers.RandString(seed, 6)

	ai.imageType, err = helpers.GetAppImageType(ai.Path)
	if err != nil { return nil, err }

	ai.tempDir, err = helpers.MakeTemp(filepath.Join(xdg.RuntimeDir, "aisap"), ai.runId)
	if err != nil { return nil, err }
	ai.rootDir = "/"

	ai.mountDir, err = helpers.MakeTemp(ai.tempDir, ".mount_" + ai.runId)

	ai.Offset, err = helpers.GetOffset(src)
	if err != nil { return nil, err }

	// Only mount if no previous instances (launced the same day) are already
	// mounted there. This is to reuse their libraries, save on RAM and to spam
	// the mount list as little as possible
	if !isMountPoint(ai.mountDir) {
		err = mount(src, ai.mountDir, ai.Offset)
		if err != nil { return nil, err }
	}

	// Prefer local entry if it exists (located at
	// ~/.local/share/aisap/[ai.Name])
	ai.Desktop, err = getEntry(ai)
	ai.Name    = ai.Desktop.Section("Desktop Entry").Key("Name").Value()
	ai.Version = ai.Desktop.Section("Desktop Entry").Key("X-AppImage-Version").Value()

	ai.UpdateInfo, _ = helpers.ReadUpdateInfo(ai.Path)

	if ai.Version == "" {
		ai.Version = "1.0"
	}

	// If PREFER_AISAP_PROFILE is set, attempt to use it over the AppImage's
	// suggested permissions. If no profile exists in aisap, fall back on saved
	// permissions in aisap, and then finally the AppImage's internal desktop
	// entry
	// Typically this should be unset unless testing a custom profile against
	// aisap's
	_, present := os.LookupEnv("PREFER_AISAP_PROFILE")
	if present {
		ai.Perms, err = profiles.FromName(ai.Name)
		if err != nil {
			ai.Perms, err = permissions.FromSystem(ai.Name)
		}
	} else {
		ai.Perms, err = permissions.FromSystem(ai.Name)
		if err != nil {
			ai.Perms, err = profiles.FromName(ai.Name)
		}
	}

	if err != nil {
		ai.Perms, _ = permissions.FromIni(ai.Desktop)
	}

	return ai, nil
}

// Return a reader for the `.DirIcon` file of the AppImage, converting it to
// PNG if it's in SVG or XPM format
func (ai AppImage) Thumbnail() (io.Reader, error) {
	// Try to extract from zip, continue to SquashFS if it fails
	if ai.imageType == -2 {
		r, err := helpers.ExtractResourceReader(ai.Path, "icon/256.png")
		if err == nil { return r, nil }
	}

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

func (ai AppImage) Md5() string {
	return ai.md5
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
	if ai.imageType == -2 {
		r, err := helpers.ExtractResourceReader(ai.Path, "icon/default.svg")
		// Didn't really know what to put in the string here as the name inside
		// the zip is always `default`, so just decided to use the extension
		if err == nil { return r, ".svg", nil }

		r, err  = helpers.ExtractResourceReader(ai.Path, "icon/default.png")
		if err == nil { return r, ".png", nil }
	}

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

// Extract the desktop file from the AppImage
func getEntry(ai *AppImage) (*ini.File, error) {
	var err error
	var f   io.ReadCloser
	var e   string

	if ai.imageType == -2 {
		f, err = helpers.ExtractResourceReader(ai.Path, "desktop_entry")
	}

	// Extract from SquashFS if type 2 or zip fails
	if ai.imageType == 2 || err != nil {
		// Return all `.desktop` files. A vadid AppImage should only have one
		var fp []string
		fp, err = filepath.Glob(ai.mountDir + "/*.desktop")
		if err != nil { return nil, err }
		f, err = os.Open(fp[0])
		defer f.Close()
	}

	// Replace normal semicolons with fullwidth semicolons so that it doen't
	// interfere with the INI parsing
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		e = e + strings.ReplaceAll(scanner.Text(), ";", "；") + "\n"
	}

	return ini.Load([]byte(e))
}
