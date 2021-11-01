package helpers

import (
    "path/filepath"
    "math/rand"
    "strings"
    "os"
    "fmt"
)

func PrintlnVerbose(str string) {
    var verbose *bool

    if *verbose {
        fmt.Println(str)
    }
}

func ErrorCheck(str string, err error, fatal bool) {
    if err != nil && !fatal {
        fmt.Fprintln(os.Stderr, "WARNING: " + str, err)
    } else if err != nil && fatal {
        fmt.Fprintln(os.Stderr, "FATAL: " + str, err)
        os.Exit(1)
    }
}

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
    // Valid characters for creating the directory name

    dir := filepath.Clean(path+"/"+name)
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
