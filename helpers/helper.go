package helpers

import (
	"path/filepath"
   	"math/rand"
   	"strings"
   	"os"
	"os/user"

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

func CleanFiles(s []string) []string {
	for i := range(s) {
		// Get the last 3 chars of the file entry
		var ex string
		if len(s[i]) >= 3 {
			ex = s[i][len(s[i])-3:]
		} else {
			ex = ":ro"
		}

		s[i] = ExpandDir(s[i])

		if ex != ":ro" && ex != ":rw" {
			s[i] = s[i] + ":ro"
		}
	}

	return s
}

func CleanDevices(s []string) []string {
	for i := range(s) {
		if len(s[i]) > 5 && s[i][0:5] == "/dev/" {
			s[i] = strings.Replace(s[i], "/dev/", "", 1)
		}
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

	return expandEither(str, xdgDirs)
}

func ExpandGenericDir(str string) string {
	usr, _ := user.Current()
	homed  := filepath.Join("/home", usr.Username)

	xdgDirs := map[string]string{
		"xdg-home":        homed,
		"xdg-desktop":     filepath.Join(homed, "Desktop"),
		"xdg-download":    filepath.Join(homed, "Downloads"),
		"xdg-documents":   filepath.Join(homed, "Documents"),
		"xdg-music":       filepath.Join(homed, "Music"),
		"xdg-pictures":    filepath.Join(homed, "Pictures"),
		"xdg-videos":      filepath.Join(homed, "Videos"),
		"xdg-templates":   filepath.Join(homed, "Templates"),
		"xdg-publicshare": filepath.Join(homed, "Share"),
		"xdg-config":      filepath.Join(homed, ".config"),
		"xdg-cache":       filepath.Join(homed, ".cache"),
		"xdg-data":        filepath.Join(homed, ".local/share"),
		"xdg-state":       filepath.Join(homed, ".local/state"),
	}

	return expandEither(str, xdgDirs)
}
