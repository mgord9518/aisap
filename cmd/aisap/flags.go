package main

import (
	"fmt"
	"os"

	clr "github.com/gookit/color"
	aisap "github.com/mgord9518/aisap"
	flag "github.com/spf13/pflag"
)

type arrayFlags []string

var (
	// Normal flags
	help      = flag.BoolP("help", "h", false, "display this help menu")
	listPerms = flag.BoolP("list-perms", "l", false, "print all permissions to be granted to the app")
	verbose   = flag.BoolP("verbose", "v", false, "make output more verbose")

	// Long-only flags
	color            = flag.Bool("color", true, "whether to show color (default true)")
	example          = flag.Bool("example", false, "print out examples")
	level            = flag.Int("level", -1, "change the permissions level")
	rootDir          = flag.String("root-dir", "", "use a different filesystem root for system files")
	dataDir          = flag.String("data-dir", "", "change the AppImage's sandbox home location")
	noDataDir        = flag.Bool("no-data-dir", false, "force AppImage's HOME to be a tmpfs (default false)")
	extractIcon      = flag.String("extract-icon", "", "extract the AppImage's icon")
	extractThumbnail = flag.String("extract-thumbnail", "", "extract the AppImage's thumbnail preview")
	profile          = flag.String("profile", "", "use a profile from a desktop entry")
	fallbackProfile  = flag.String("fallback-profile", "", "set profile to fallback on if one isn't found")
	version          = flag.Bool("version", false, "show the version and quit")
	trustOnce        = flag.Bool("trust-once", false, "trust the AppImage for one run")
	trust            = flag.Bool("trust", false, "set whether the AppImage is trusted or not")

	addFile   arrayFlags
	addDevice arrayFlags
	addSocket arrayFlags
	rmFile    arrayFlags
	rmDevice  arrayFlags
	rmSocket  arrayFlags
	// Flags that can be called multiple times
//	extract   arrayFlags // Extract an arbitrary file from the AppImage
)

// Initialization of global variables and help menu
func init() {
	var present bool
	handleCtrlC()

	flag.Var(&addFile, "add-file", "give the sandbox access to a filesystem object")
	flag.Var(&addDevice, "add-device", "add a device to the sandbox (eg dri)")
	flag.Var(&addSocket, "add-socket", "allow the sandbox to access another socket (eg x11)")
	flag.Var(&rmFile, "rm-file", "revoke a file from the sandbox")
	flag.Var(&rmDevice, "rm-device", "remove access to a device")
	flag.Var(&rmSocket, "rm-socket", "disable a socket")

	// Prefer AppImage-provided variable `ARGV0` if present
	if argv0, present = os.LookupEnv("ARGV0"); !present {
		argv0 = os.Args[0]
	}

	flag.Usage = func() {
		clr.Printf("<yellow>usage</>: <blue>%s</> [<green>OPTIONS</>] [<green>APPIMAGE</>]\n", argv0)
		clr.Printf("<yellow>description</>: easily sandbox AppImages in bwrap\n")
		clr.Printf("\n<yellow>normal options</>:\n")
		printUsage("help")
		printUsage("list-perms")
		printUsage("verbose")
		clr.Printf("\n<yellow>long-only options</>:\n")
		printUsage("example")
		printUsage("level")
		printUsage("trust")
		printUsage("trust-once")
		printUsage("root-dir")
		printUsage("data-dir")
		printUsage("no-data-dir")
		printUsage("add-file")
		printUsage("add-device")
		printUsage("add-socket")
		printUsage("rm-file")
		printUsage("rm-device")
		printUsage("rm-socket")
		printUsage("extract-icon")
		printUsage("extract-thumbnail")
		printUsage("profile")
		printUsage("fallback-profile")
		printUsage("version")
		clr.Printf("\n<yellow>enviornment variables</>:\n")
		clr.Printf("  <cyan>NO_COLOR</>:                disable color\n")
		clr.Printf("  <cyan>PREFER_AISAP_PROFILE</>:    prefer aisap internal profile over the user defined one\n")
		clr.Printf("\n<yellow>config</>:\n")
		clr.Printf("  <cyan>~/.local/share/aisap</>: parent config directory\n")
		clr.Printf("  <cyan>profiles</>:             location to overwrite AppImage permissions\n")
		clr.Printf("\n<yellow>tips</>:\n")
		clr.Printf("  <cyan>1</>: permissions highlighted in <lightYellow>yellow</> are potential escape vectors")
		clr.Printf("\n\n<yellow>homepage</>: <https://github.com/mgord9518/aisap> (see for full documentation)\n\n")
		clr.Printf("<red>warning</>: THIS SOFTWARE IS PROVIDED WITH NO WARRANTY! USE AT YOUR OWN RISK!\n")
		os.Exit(0)
	}

	flag.Parse()

	if *version {
		fmt.Println(aisap.Version)
		os.Exit(0)
	}

	if *example {
		clr.Printf("<yellow>examples</>:\n")
		clr.Printf("  <blue>%s</> <cyan>--profile</>=./f.desktop ./f.app\n", argv0)
		clr.Printf("    sandbox `f.app` using permissions from `f.desktop`\n\n")
		clr.Printf("  <blue>%s</> <cyan>--level 2</> ./f.app\n", argv0)
		fmt.Printf("    change `f.app` sandbox base to level 2\n\n")
		clr.Printf("  <blue>%s</> <cyan>--add-file</> other.bin ./f.app <green>--add-file</>=./f.txt\n", argv0)
		fmt.Printf("    allow sandbox to access files `f.txt` and `other.bin`\n\n")
		clr.Printf("  <blue>%s</> <cyan>--rm-file</> secret.txt./f.app\n", argv0)
		fmt.Printf("    revoke access to `secret.txt` in the sandbox\n")
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

		for i := len(fg.Name); i < 19; i++ {
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

func flagUsed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
