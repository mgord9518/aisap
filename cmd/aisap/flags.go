package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	clr  "github.com/gookit/color"
)

type arrayFlags []string

var (
	ver = "0.3.7-alpha"
)

var (
	// Normal flags
	help      = flag.BoolP("help",       "h", false, "display this help menu")
	verbose   = flag.BoolP("verbose",    "v", false, "make output more verbose")
	listPerms = flag.BoolP("list-perms", "l", false, "print all permissions to be granted to the app")

	// Long-only flags
	color   = flag.Bool("color",   true,  "whether to show color (default true)")
	example = flag.Bool("example", false, "print out examples")
	version = flag.Bool("version", false, "show the version and quit")
	profile = flag.String("profile", "",  "use a profile from a desktop entry")
	level   = flag.Int("level",   -1,     "change the permissions level")

	// Flags that can be called multiple times
	addFile   arrayFlags
	addDevice arrayFlags
	addSocket arrayFlags
	rmFile    arrayFlags
	rmDevice  arrayFlags
	rmSocket  arrayFlags
)

// Initialization of global variables and help menu
func init() {
	var present bool
	handleCtrlC()

	flag.Var(&addFile,   "add-file",   "give the sandbox access to a filesystem object")
	flag.Var(&addDevice, "add-device", "add a device to the sandbox (eg dri)")
	flag.Var(&addSocket, "add-socket", "allow the sandbox to access another socket (eg x11)")
	flag.Var(&rmFile,    "rm-file",    "revoke a file from the sandbox")
	flag.Var(&rmDevice,  "rm-device",  "remove access to a device")
	flag.Var(&rmSocket,  "rm-socket",  "disable a socket")

	// Prefer AppImage-provided variable `ARGV0` if present
	if argv0, present = os.LookupEnv("ARGV0"); !present {
		argv0 = os.Args[0]
	}

	flag.Usage = func() {
		clr.Printf("<yellow>usage</>: <blue>%s</> [OPTIONS] [APPIMAGE]\n", argv0)
		clr.Printf("<yellow>description</>: easily sandbox AppImages in BubbleWrap\n")
		clr.Printf("\n<yellow>normal options</>:\n")
		printUsage("help")
		printUsage("list-perms")
		clr.Printf("\n<yellow>long-only options</>:\n")
		printUsage("example")
		printUsage("level")
		printUsage("add-file")
		printUsage("add-device")
		printUsage("add-socket")
		printUsage("rm-file")
		printUsage("rm-device")
		printUsage("rm-socket")
		printUsage("profile")
		printUsage("version")
		clr.Printf("\n<yellow>enviornment variables</>:\n")
		clr.Printf("  <cyan>NO_COLOR</>: disable color\n")

//		clr.Printf("\n<red>WARNING:</> no sandbox is impossible to escape! This is to *aid* security, not\n")
//		fmt.Printf("guarentee safety when downloading sketchy stuff online. Don't be stupid!\n\n")
		clr.Printf("\n<yellow>homepage</>: <https://github.com/mgord9518/aisap>\n\n")
//		fmt.Printf("Plus, this is ALPHA software! Very little testing has been done;\n")
		clr.Printf("<red>USE AT YOUR OWN RISK!</>\n")
		os.Exit(0)
	}

	flag.Parse()


	if *version {
		fmt.Println(ver)
		os.Exit(0)
	}

	if *example {
//		clr.Printf("<yellow>examples</>:\n")
//		fmt.Printf("  %s%s --profile%s=./f.desktop -- ./f.app\n", g, argv0, z)
//		fmt.Printf("    sandbox `f.app` using permissions from `f.desktop`\n\n")
//		fmt.Printf("  %s%s ./f.app --level%s=2\n", g, argv0, z)
//		fmt.Printf("    tighten `f.app` sandbox to level 2 (default: 1)\n\n")
//		fmt.Printf("  %s%s --add-file%s=./f.txt %s--add-file%s ./other.bin ./f.app\n", g, argv0, z, g, z)
//		fmt.Printf("    allow sandbox to access files `f.txt` and `other.bin`\n\n")
//		fmt.Printf("  %s%s --rm-file%s=./secret.txt ./f.app\n", g, argv0, z, g, z)
//		fmt.Printf("    revoke access to `secret.txt` in the sandbox\n")
		os.Exit(0)
	}

	if *help || len(os.Args) < 2 {
		flag.Usage()
	}
}

func printUsage(name string) {
	fg := flag.Lookup(name)

	if len(fg.Shorthand) > 0 {
		clr.Printf("  <cyan>-%s</>, <cyan>--%s</>:", fg.Shorthand, fg.Name)

		// Pad with spaces
		for i := len(fg.Name); i < 12; i++ {
			fmt.Print(" ")		
		}
	} else {
		clr.Printf("  <cyan>--%s</>:", fg.Name)

		for i := len(fg.Name); i < 12; i++ {
			fmt.Print(" ")		
		}
	}

	fmt.Println(fg.Usage)

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
