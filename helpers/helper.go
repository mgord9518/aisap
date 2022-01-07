package helpers

import (
    "path/filepath"
    "math/rand"
    "strings"
    "os"
)

// Converts a multi-item INI value into a slice
// eg: `foo;bar;` becomes []string{ "foo", "bar" }
func SplitKey(str string) []string {
    str = strings.ReplaceAll(str, "；", ";")
    f := func(c rune) bool { return c == ';' }

    return strings.FieldsFunc(str, f)
}

func Contains(slice []string, str string) (int, bool) {
    for i, val := range slice {
        if val == str { return i, true }
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
	// Filename friendly base64
    chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+.")

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
	var ex string

	for i := range(s) {
		// Get the last 3 chars of the file entry
		if len(s[i]) >= 3 {
			ex = s[i][len(s[i])-3:]
		} else {
			ex = ":ro"
		}

		if ex != ":ro" && ex != ":rw" {
			s[i] = s[i]+":ro"
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
