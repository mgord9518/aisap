package main

import (
    "fmt"
    "os"

    flag "github.com/spf13/pflag"
)

type arrayFlags []string

var (
    // Normal flags
    help      = flag.BoolP("help",    "h", false, "")
    verbose   = flag.BoolP("verbose", "v", false, "")

    // Long-only flags
    permFile = flag.String("perm-file", "", "")
    level    = flag.Int("level",        1,  "")

	// Flags that can be called multiple times
	addFile arrayFlags
	addDev  arrayFlags
	addSoc  arrayFlags
)

// Initialization of global variables and help menu
func init() {
    var present bool

    flag.Var(&addFile, "add-file", "")
    flag.Var(&addDev,  "add-dev",  "")
    flag.Var(&addSoc,  "add-soc",  "")

    if argv0, present = os.LookupEnv("ARGV0"); !present {
        argv0 = os.Args[0]
    }

    flag.Usage = func() {
        fmt.Printf("Usage: %s [OPTIONS] [APPIMAGE]\n", argv0)
        fmt.Println("Easily sandbox AppImages in BubbleWrap\n")
        fmt.Println("With no PERMFILE, read permissions directly from AppImage")
        fmt.Println("Sandbox level of 0 only changes data directory, not actually sandboxed!\n")
        fmt.Println("Normal options:")
        fmt.Println("  -h, --help    Display this help menu")
        fmt.Println("  -v, --verbose Be more verbose (NEI)\n")
        fmt.Println("Long-only options:")
        fmt.Println("  --add-file   Allow access to additional files")
        fmt.Println("  --add-soc    Allow access to additional sockets")
        fmt.Println("  --add-dev    Allow access to additional /dev files")
        fmt.Println("  --level      Change the base security level of the sandbox (min: 0, max: 3)")
        fmt.Println("  --perm-file  Look for permissions in this entry instead of the AppImage\n")
        fmt.Println("Examples:")
        fmt.Printf("  %s --perm-file=./f.desktop -- ./f.app\n", argv0)
        fmt.Println("    Sandbox `f.app` using permissions from `f.desktop`\n")
        fmt.Printf("  %s ./f.app --level=2\n", argv0)
        fmt.Println("    Tighten `f.app` sandbox to level 2 (default: 1)\n")
        fmt.Printf("  %s --add-file=./f.txt --add-file ./other.bin ./f.app\n", argv0)
        fmt.Println("    Allow sandbox to access files `f.txt` and `other.bin`\n")
        fmt.Println("WARNING: No sandbox is impossible to escape! This is to *aid* security, not")
        fmt.Println("guarentee safety when downloading sketchy stuff online. Don't be stupid!\n")
        fmt.Println("Plus, this is ALPHA software! Very little testing has been done; USE AT YOUR")
        fmt.Println("OWN RISK!")
        os.Exit(0)
    }

    flag.Parse()

    if *help || len(os.Args) < 2 {
        flag.Usage()
    }
}

func (i *arrayFlags) Set(value string) error {
    *i = append(*i, value)
    return nil
}

func (i *arrayFlags) String() string {
    return ""
}

func (i *arrayFlags) Type() string {
    return ""
}
