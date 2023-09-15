// Drop-in replacemnt for go-appimage for sandboxing and use with shappimages
// NOT FINISHED AND STILL LACKING BASIC FEATURES
// THIS SHOULD BE USED FOR TESTING PURPOSES *ONLY* UNTIL IN A STABLE STATE

package aisap

import (
	"bufio"
	"crypto/md5"
	"debug/elf"
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
	squashfs    "github.com/CalebQ42/squashfs"
	xdg         "github.com/adrg/xdg"
)

type AppImage struct {
	Desktop       *ini.File                  // INI of internal desktop entry
	Perms         *permissions.AppImagePerms // Permissions
	Path           string // Location of AppImage
	dataDir        string // The AppImage's `HOME` directory
	rootDir        string // Can be used to give the AppImage fake system files
	tempDir        string // The AppImage's `/tmp` directory
	mountDir       string // The location the AppImage is mounted at
	md5            string // MD5 of AppImage's URI
	Name           string // AppImage name from the desktop entry
	Version        string
	UpdateInfo     string
	Offset         int    // Offset of SquashFS image
	imageType      int    // Type of AppImage (1=ISO 9660 ELF, 2=squashfs ELF, -2=shImg shell)
	architecture []string // List of CPU architectures supported by the bundle
	reader        *squashfs.Reader
	file          *os.File

	// These will both be removed when the Zig-implemented C bindings
	// become usable
	CurrentArg     int    // Should only ever be used for the C bindings
	WrapArgsList []string // Should only ever be used for the C bindings
}

// Current version of aisap
const (
	Version = "0.9.7-alpha"
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

	ai.imageType, err = helpers.GetAppImageType(ai.Path)
	if err != nil { return nil, err }

	ai.rootDir = "/"
	ai.dataDir = ai.Path + ".home"
	ai.Offset, err = helpers.GetOffset(src)
	if err != nil { return nil, err }

	if ai.imageType == -2 || ai.imageType == 2 {
		ai.file, err = os.Open(ai.Path)
		if err != nil { return nil, err }

		info, _ := ai.file.Stat()
		off64 := int64(ai.Offset)
		r := io.NewSectionReader(ai.file, off64, info.Size()-off64)

		ai.reader, err = squashfs.NewReader(r)
		if err != nil { return nil, err }
	}

	// Prefer local entry if it exists (located at $XDG_DATA_HOME/aisap/[ai.Name])
	desktopReader, err := ai.getEntry()
	if err != nil { return ai, err }

	ai.Desktop, err = ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, desktopReader)
	if err != nil { return ai, err }

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
	if _, present := os.LookupEnv("PREFER_AISAP_PROFILE"); present {
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

	// Fall back to permissions inside AppImage if all else fails
	if err != nil {
		ai.Perms, _ = permissions.FromIni(ai.Desktop)
	}

	return ai, nil
}

// Returns `true` if the AppImage in question is both executable and has
// its profile copied to the aisap config dir. This is to ensure the
// permissions can't change under the user's feet through an update to the
// AppImage
func (ai *AppImage) Trusted() bool {
	aisapConfig := filepath.Join(xdg.DataHome, "aisap", "profiles")
	filePath := filepath.Join(aisapConfig, ai.Name)

	// If the AppImage permissions exist in aisap's config directory and the
	// AppImage is executable, we consider it trusted
	if helpers.FileExists(filePath) {
		info, err := os.Stat(ai.Path)
		if err != nil {
			return false
		}

		return info.Mode() & 0100 != 0
	}

	return false
}

func (ai *AppImage) SetTrusted(trusted bool) error {
	aisapConfig := filepath.Join(xdg.DataHome, "aisap", "profiles")
	filePath := filepath.Join(aisapConfig, ai.Name)

	if trusted {
		if !helpers.DirExists(aisapConfig) {
			os.MkdirAll(aisapConfig, 0744)
		}

		info, err := os.Stat(ai.Path)
		if err != nil {
			return err
		}

		os.Chmod(ai.Path, info.Mode() | 0100)

		if helpers.FileExists(filePath) {
			return errors.New("entry already exists in aisap config dir")
		}

		desktopReader, _ := ai.getEntry()
		permFile, _ := os.Create(filePath)
		io.Copy(permFile, desktopReader)
	} else {
		os.Remove(filePath)
	}

	return nil
}

// Return a reader for the `.DirIcon` file of the AppImage
func (ai *AppImage) Thumbnail() (io.Reader, error) {
	// Try to extract from zip, continue to SquashFS if it fails
	if ai.imageType == -2 {
		r, err := helpers.ExtractResourceReader(ai.Path, "icon/256.png")
		if err == nil { return r, nil }
	}

	return ai.ExtractFileReader(".DirIcon")
}

func (ai *AppImage) RootDir() string {
	return ai.rootDir
}

func (ai *AppImage) DataDir() string {
	return ai.dataDir
}

func (ai *AppImage) TempDir() string {
	return ai.tempDir
}

func (ai *AppImage) Md5() string {
	return ai.md5
}

func (ai *AppImage) MountDir() string {
	return ai.mountDir
}

// Set the directory the sandbox pulls system files from
func (ai *AppImage) SetRootDir(d string) {
	ai.rootDir = d
}

// Set the directory for the sandboxed AppImage's `HOME`
func (ai *AppImage) SetDataDir(d string) {
	ai.dataDir = d
}

// Set the directory for the sandboxed AppImage's `TMPDIR`
func (ai *AppImage) SetTempDir(d string) {
	ai.tempDir = d
}

// Return type of AppImage
func (ai *AppImage) Type() int {
	t, _ := helpers.GetAppImageType(ai.Path)

	return t
}

func (ai *AppImage) Architectures() []string {
	s, _ := ai.getArchitectures()

	return s
}

// Extract a file from the AppImage's interal filesystem image
func (ai *AppImage) ExtractFile(path string, dest string, resolveSymlinks bool) error {
	// Remove file if it already exists
	os.Remove(filepath.Join(dest))
	info, err := os.Lstat(path)

	// True if file is symlink and `resolveSymlinks` is false
	if info != nil && !resolveSymlinks &&
	info.Mode()&os.ModeSymlink == os.ModeSymlink {
		target, _ := os.Readlink(path)
		err = os.Symlink(target, dest)
	} else {
		inF, err := ai.ExtractFileReader(path)
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
func (ai *AppImage) ExtractFileReader(path string) (io.ReadCloser, error) {
	f, err := ai.reader.Open(path)
	if err != nil {
		return f, err
	}

	r := f.(*squashfs.File)

	if r.IsSymlink() {
		r = r.GetSymlinkFile()
	}

	return r, err
}

// Returns the icon reader of the AppImage, valid formats are SVG and PNG
func (ai *AppImage) Icon() (io.ReadCloser, string, error) {
	if ai.imageType == -2 {
		r, err := helpers.ExtractResourceReader(ai.Path, "icon/default.svg")
		if err == nil { return r, "icon/default.svg", nil }

		r, err  = helpers.ExtractResourceReader(ai.Path, "icon/default.png")
		if err == nil { return r, "icon/default.png", nil }
	}

	if ai.Desktop == nil {
		return nil, "", InvalidDesktopFile
	}

	// Return error if desktop file has no icon
	iconf := ai.Desktop.Section("Desktop Entry").Key("Icon").Value()
	if iconf == "" {
		return nil, "", NoIcon
	}

	// If the desktop entry specifies an extension, use it
	if strings.HasSuffix(iconf, ".png") || strings.HasSuffix(iconf, ".svg") {
		r, err := ai.ExtractFileReader(iconf)
		return r, iconf, err
	}

	// If not, iterate through all AppImage specified image formats
	extensions := []string{
		".png",
		".svg",
	}

	for _, ext := range(extensions) {
		r, err := ai.ExtractFileReader(iconf + ext)

		if err == nil {
			return r, path.Base(iconf + ext), err
		}
	}

	return nil, "", InvalidIconExtension
}

// Extract the desktop file from the AppImage
func (ai *AppImage) getEntry() (io.Reader, error) {
	var r   io.Reader
	var err error

	if ai.imageType == -2 {
		r, err = helpers.ExtractResourceReader(ai.Path, "desktop_entry")
	}

	// Extract from SquashFS if type 2 or zip fails
	if ai.imageType == 2 || err != nil {
		// Return all `.desktop` files. A vadid AppImage should only have one
		var fp []string

		fp, err = ai.reader.Glob("*.desktop")
		if len(fp) != 1 {
			return nil, NoDesktopFile
		}

		entry, err := ai.reader.Open(fp[0])

		r := entry.(*squashfs.File)

		if r.IsSymlink() {
			r = r .GetSymlinkFile()
		}

		return r, err
	}


	return r, err
}

// Determine what architectures a bundle supports
func (ai *AppImage) getArchitectures() ([]string, error) {
	a := ai.Desktop.Section("Desktop Entry").Key("X-AppImage-Architecture").Value()
	s := helpers.SplitKey(a)

	if len(s) > 0 {
		return s, nil
	}

	// If undefined in the desktop entry, assume arch via ELF AppImage runtime
	if ai.Type() >= 0 {
		e, err := elf.NewFile(ai.file)
		if err != nil {return s, err}

		switch e.Machine {
			case elf.EM_386:
				return []string{"i386"},    nil
			case elf.EM_X86_64:
				return []string{"x86_64"},  nil
			case elf.EM_ARM:
				return []string{"armhf"},   nil
			case elf.EM_AARCH64:
				return []string{"aarch64"}, nil
		}
	}

	// Assume arch via shImg runtime
	if ai.Type() < -1 {
		scanner  := bufio.NewScanner(ai.file)
		arches := []string{}

		counter := 0
		for scanner.Scan() {
			counter++
			if strings.HasPrefix(scanner.Text(), "arch='") {
				str   := scanner.Text()
				str    = strings.ReplaceAll(str, "arch='", "")
				str    = strings.ReplaceAll(str, "'",      "")
				arches = helpers.SplitKey(str)
				return arches, nil
			}

			// All shImg info should be at the top of the file, 50 is more than
			// enough
			if counter >= 50 {
				break
			}
		}
	}

	return s, errors.New("failed to determine arch")
}
