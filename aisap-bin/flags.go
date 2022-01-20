package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

type arrayFlags []string

var (
	g = "\033[0;32m" // Green
	y = "\033[0;33m" // Yellow
	r = "\033[0;31m" // Red
	c = "\033[0;36m" // Cyan
	z = "\033[0;0m"  // Reset

	ver = "0.2.5"
)

var (
	// Normal flags
	help      = flag.BoolP("help",       "h", false, "")
	verbose   = flag.BoolP("verbose",    "v", false, "")
	listPerms = flag.BoolP("list-perms", "l", false, "")

	// Long-only flags
	color   =   flag.Bool("color",   true,  "")
	example =   flag.Bool("example", false, "")
	version =   flag.Bool("version", false, "")
	profile =   flag.String("profile", "",  "")
	level   =   flag.Int("level",   -1,     "")

	// Flags that can be called multiple times
	file     arrayFlags
	device   arrayFlags
	socket   arrayFlags
	rmfile   arrayFlags
	rmdevice arrayFlags
	rmsocket arrayFlags
)

// Initialization of global variables and help menu
func init() {
	var present bool
	handleCtrlC()

	flag.Var(&file,   "file",   "")
	flag.Var(&device, "device", "")
	flag.Var(&socket, "socket", "")
	flag.Var(&rmfile,   "rmfile",   "")
	flag.Var(&rmdevice, "rmdevice", "")
	flag.Var(&rmsocket, "rmsocket", "")

	// Prefer AppImage-provided variable `ARGV0` if present
	if argv0, present = os.LookupEnv("ARGV0"); !present {
		argv0 = os.Args[0]
	}

	flag.Usage = func() {
		fmt.Printf("usage: %s%s%s [OPTIONS] [APPIMAGE]\n\n", g, argv0, z)
		fmt.Printf("wrapper binary around aisap to easily sandbox AppImages in BubbleWrap\n")
		fmt.Printf("with no PERMFILE, read permissions directly from AppImage or internal\n")
		fmt.Printf("permissions library\n")
		fmt.Printf("sandbox level of 0 only changes data directory, not actually sandboxed!\n\n")
		fmt.Printf("%snormal options:\n", y)
		fmt.Printf("%s  -h, --help        %sdisplay this help menu\n", g, z)
		fmt.Printf("%s  -l, --list-perms  %slist permissions to be granted to the app\n\n", g, z)
//		fmt.Printf("%s  -v, --verbose  %sbe more verbose (NEI)\n\n", g, z)
		fmt.Printf("%slong-only options:\n", y)
		fmt.Printf("%s  --color    %swhether color should be shown (default: true)\n", g, z)
		fmt.Printf("%s  --example  %sshow usage examples\n", g, z)
		fmt.Printf("%s  --file     %sadd file to sandbox\n", g, z)
		fmt.Printf("%s  --socket   %sallow access to additional sockets\n", g, z)
		fmt.Printf("%s  --device   %sallow access to additional /dev files\n", g ,z)
		fmt.Printf("%s  --rmfile   %sremove file from sandbox\n", g, z)
		fmt.Printf("%s  --rmsocket %sremove sandbox access to socket\n", g, z)
		fmt.Printf("%s  --rmdevice %sremove access to device file\n", g ,z)
		fmt.Printf("%s  --level    %schange the base security level of the sandbox (min: 0, max: 3)\n", g, z)
		fmt.Printf("%s  --profile  %slook for permissions in this entry instead of the AppImage\n", g, z)
		fmt.Printf("%s  --version  %sprint the version and exit\n\n", g, z)
		fmt.Printf("%sWARNING:%s no sandbox is impossible to escape! This is to *aid* security, not\n", r, z)
		fmt.Printf("guarentee safety when downloading sketchy stuff online. Don't be stupid!\n\n")
		fmt.Printf("for more information, visit <://github.com/mgord9518/aisap>\n\n")
		fmt.Printf("Plus, this is ALPHA software! Very little testing has been done;\n")
		fmt.Printf("%sUSE AT YOUR OWN RISK!%s\n", r, z)
		os.Exit(0)
	}

	flag.Parse()

	// Remove color if `color=false`
	if !*color {
		g = ""
		y = ""
		r = ""
	}


	if *version {
		fmt.Println(ver)
		os.Exit(0)
	}

	if *example {
		fmt.Printf("%sexamples:%s\n", y, z)
		fmt.Printf("  %s%s --profile%s=./f.desktop -- ./f.app\n", g, argv0, z)
		fmt.Printf("    sandbox `f.app` using permissions from `f.desktop`\n\n")
		fmt.Printf("  %s%s ./f.app --level%s=2\n", g, argv0, z)
		fmt.Printf("    tighten `f.app` sandbox to level 2 (default: 1)\n\n")
		fmt.Printf("  %s%s --file%s=./f.txt %s--file%s ./other.bin ./f.app\n", g, argv0, z, g, z)
		fmt.Printf("    allow sandbox to access files `f.txt` and `other.bin`\n")
		os.Exit(0)
	}

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
