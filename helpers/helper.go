package helpers

import (
	"archive/zip"
	"bufio"
	"errors"
	"io"
	"path/filepath"
   	"math/rand"
	"strconv"
   	"strings"
   	"os"

	xdg "github.com/adrg/xdg"
)

// Converts a multi-item INI value into a slice
// eg: `foo;bar;` becomes []string{ "foo", "bar" }
func SplitKey(str string) []string {
	str = strings.ReplaceAll(str, "；", ";")
	f := func(c rune) bool { return c == ';' }

	return strings.FieldsFunc(str, f)
}

func Contains(s []string, str string) (int, bool) {
	for i, val := range(s) {
		if val == str { return i, true }
	}

	return -1, false
}

// Checks if an array contains any of the elements from another array
func ContainsAny(s []string, s2 []string) (int, bool) {
	for i := range(s2) {
		n, present := Contains(s, s2[i])
		if present { return n, true }
	}

	return -1, false
}

// Takes a full path and prefix, creates a temporary directory and returns its path
func MakeTemp(path string, name string) (string, error) {
	dir := filepath.Clean(filepath.Join(path, name))
	err := os.MkdirAll(dir, 0744)
	return dir, err
}

func RandString(seed int, length int) string {
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	rand.Seed(int64(seed))

	s := make([]rune, length)

	for i := range s {
		s[i] = chars[rand.Intn(len(chars))]
	}

	return string(s)
}

func DirExists(path string) bool {
	info, err := os.Stat(path)

	if os.IsNotExist(err) { return false }

	return info.IsDir()
}

func FileExists(path string) bool {
	_, err := os.Stat(path)

	if os.IsNotExist(err) { return false }

	return true
}

func CleanFile(str string) string {
	// Get the last 3 chars of the file entry
	var ex string
	if len(str) >= 3 {
		ex = str[len(str)-3:]
	} else {
		ex = ":ro"
	}

	str = ExpandDir(str)

	if ex != ":ro" && ex != ":rw" {
		str = str + ":ro"
	}

	return str
}

func CleanFiles(s []string) []string {
	for i := range(s) {
		s[i] = CleanFile(s[i])
	}

	return s
}

func CleanDevice(str string) string {
	if len(str) > 5 && str[0:5] == "/dev/" {
		str = strings.Replace(str, "/dev/", "", 1)
	}

	return str
}

func CleanDevices(s []string) []string {
	for i := range(s) {
		s[i] = CleanDevice(s[i])
	}

	return s
}

// Expands XDG formatted directories into full paths depending on the input map
func expandEither(str string, xdgDirs map[string]string) string {
	for key, val := range xdgDirs {
		// If length of key bigger than requested directory or not equal to it
		// continue because there is no reason to look at it further
		if len(key) > len(str) || key != str[:len(key)] {
			continue
		}

		if key == str[:len(key)] {
			str = strings.Replace(str, key, val, 1)
		} else if c := str[len(key)]; c == byte('/') {
			str = strings.Replace(str, key, val, 1)
		}
	}

	// Resolve `../` and clean up extra slashes if they exist
	str = filepath.Clean(str)

	// Expand tilde with the true home directory if not generic, otherwise use
	// a generic representation
	if str[0] == '~' {
		str = strings.Replace(str, "~", xdgDirs["xdg-home"], 1)
	}

	// If generic, will fake the home dir. Otherwise does nothing
	str = strings.Replace(str, xdg.Home, xdgDirs["xdg-home"], 1)

	return str
}

// Expand xdg and shorthand directories into either real directories on the
// user's machine or some generic names to be used to protect the actual path
// names in case the user has changed them
func ExpandDir(str string) string {
	// Reset HOME and reload XDG because these directories NEED to be the
	// actual system dirs or aisap won't be able to give access to them
	// otherwise. All of aisap's config files will still be stored in its
	// portable directory if it exists
	home, present := os.LookupEnv("HOME")
	newHome, _ := RealHome()
	os.Setenv("HOME", newHome)
	xdg.Reload()

	xdgDirs := map[string]string{
		"xdg-home":        xdg.Home,
		"xdg-desktop":     xdg.UserDirs.Desktop,
		"xdg-download":    xdg.UserDirs.Download,
		"xdg-documents":   xdg.UserDirs.Documents,
		"xdg-music":       xdg.UserDirs.Music,
		"xdg-pictures":    xdg.UserDirs.Pictures,
		"xdg-videos":      xdg.UserDirs.Videos,
		"xdg-templates":   xdg.UserDirs.Templates,
		"xdg-publicshare": xdg.UserDirs.PublicShare,
		"xdg-config":      xdg.ConfigHome,
		"xdg-cache":       xdg.CacheHome,
		"xdg-data":        xdg.DataHome,
		"xdg-state":       xdg.StateHome,
	}

	if present {
		os.Setenv("HOME", home)
	}
	xdg.Reload()

	return expandEither(str, xdgDirs)
}

func ExpandGenericDir(str string) string {
	home, present := os.LookupEnv("HOME")
	newHome, _ := RealHome()
	os.Setenv("HOME", newHome)
	xdg.Reload()

	xdgDirs := map[string]string{
		"xdg-home":        xdg.Home,
		"xdg-desktop":     filepath.Join(xdg.Home, "Desktop"),
		"xdg-download":    filepath.Join(xdg.Home, "Downloads"),
		"xdg-documents":   filepath.Join(xdg.Home, "Documents"),
		"xdg-music":       filepath.Join(xdg.Home, "Music"),
		"xdg-pictures":    filepath.Join(xdg.Home, "Pictures"),
		"xdg-videos":      filepath.Join(xdg.Home, "Videos"),
		"xdg-templates":   filepath.Join(xdg.Home, "Templates"),
		"xdg-publicshare": filepath.Join(xdg.Home, "Share"),
		"xdg-config":      filepath.Join(xdg.Home, ".config"),
		"xdg-cache":       filepath.Join(xdg.Home, ".cache"),
		"xdg-data":        filepath.Join(xdg.Home, ".local/share"),
		"xdg-state":       filepath.Join(xdg.Home, ".local/state"),
	}

	if present {
		os.Setenv("HOME", home)
	}
	xdg.Reload()

	return expandEither(str, xdgDirs)
}

func ExtractResource(aiPath string, src string, dest string) error {
	inF, err := ExtractResourceReader(aiPath, src)
	defer inF.Close()
	if err != nil { return err }

	outF, err := os.Create(dest)
	defer outF.Close()
	if err != nil { return err }

	 _, err = io.Copy(outF, inF)
	return err
}

func ExtractResourceReader(aiPath string, src string) (io.ReadCloser, error) {
	zr, err := zip.OpenReader(aiPath)
	if err != nil { return nil, err }

	for _, f := range(zr.File) {
		if f.Name == filepath.Join(".APPIMAGE_RESOURCES", src) {
			rc, err := f.Open()
			if err != nil { return nil, err }

			return rc, nil
		}
	}

	return nil, errors.New("failed to find `" + src + "` in AppImage resources")
}

// Get the home directory using `/etc/passwd`, discarding the $HOME variable.
// This is used in aisap so that its config files can be stored in 
func RealHome() (string, error) {
	uid := strconv.Itoa(os.Getuid())

	f, err := os.Open("/etc/passwd")
	defer f.Close()
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ":")
		if s[2] == uid {
			return s[5], nil
		}
	}

	return "", errors.New("failed to find home for uid `" + uid + "`!")
}
