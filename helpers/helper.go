package helpers

import (
    "path/filepath"
    "math/rand"
    "strings"
    "os"
)


func DesktopSlice(str string) []string {
    str = strings.ReplaceAll(str, "；", ";")
    f := func(c rune) bool { return c == ';' }

    permissionSlice := strings.FieldsFunc(str, f)
    return permissionSlice
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
