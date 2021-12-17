package aisap

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	helpers "github.com/mgord9518/aisap/helpers"
	xdg     "github.com/adrg/xdg"
)

// Run the AppImage with zero sandboxing
func Run(ai *AppImage, args []string) error {
	err = setupRun(ai)
	if err != nil { return err }

	cmd := exec.Command(filepath.Join(ai.mountDir, "AppRun"), args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin  = os.Stdin

	return cmd.Run()
}

// Wrap allows execution of AppImages through bwrap
func Sandbox(ai *AppImage, args []string) error {
	bwrapArgs := GetWrapArgs(ai)

	if _, err := exec.LookPath("bwrap"); err != nil {
		return errors.New("bubblewrap not found! It's required to use sandboing")
	}

	err = setupRun(ai)
	if err != nil { return err }

	// Bind the fake `~` and `/tmp` dirs
	bwrapArgs = append([]string{
		"--bind",   ai.dataDir, homed,
		"--bind",   ai.tempDir, "/tmp",
		"--setenv", "APPDIR",   "/tmp/.mount_"+ai.runId,
	}, bwrapArgs...)

	// Run the AppImage's AppRun through bwrap
	bwrapArgs = append(bwrapArgs, "--",
		"/tmp/.mount_"+ai.runId+"/AppRun",
	)

	// Append console arguments provided by the user
	bwrapArgs = append(bwrapArgs, args...)

	bwrap := exec.Command("bwrap", bwrapArgs...)
	bwrap.Stdout = os.Stdout
	bwrap.Stderr = os.Stderr
	bwrap.Stdin  = os.Stdin

	return bwrap.Run()
}

func setupRun(ai *AppImage) error {
	if ai.dataDir == "" {
		ai.dataDir = ai.Path + ".home"
	}

	if !helpers.DirExists(ai.dataDir) {
		err := os.MkdirAll(ai.dataDir, 0744)
		if err != nil { return err }
	}

	if !helpers.DirExists(filepath.Join(ai.dataDir,  ".local/share/appimagekit")) {
		err := os.MkdirAll(filepath.Join(ai.dataDir, ".local/share/appimagekit"), 0744)
		if err != nil { return err }
	}

	// Tell AppImages not to ask for integration
	noIntegrate, err := os.Create(filepath.Join(ai.dataDir, ".local/share/appimagekit/no_desktopintegration"))
	noIntegrate.Close()

	// Set required vars to correctly mount our target AppImage
	// If sandboxed, these values will be overwritten
	os.Setenv("TMPDIR", ai.tempDir)
	os.Setenv("HOME",   ai.dataDir)
	os.Setenv("APPDIR", ai.mountDir)

	return err
}

func GetWrapArgs(ai *AppImage) []string {
	// Real UID, for level 1 RUID and UID are the same value
	ruid := strconv.Itoa(os.Getuid())
	// Basic arguments to be used at all sandboxing levels
	cmdArgs := []string{
			"--setenv", "TMPDIR",              "/tmp",
			"--setenv", "HOME",                homed,
			"--setenv", "XDG_DESKTOP_DIR",     filepath.Join(homed, "Desktop"),
			"--setenv", "XDG_DOWNLOAD_DIR",    filepath.Join(homed, "Downloads"),
			"--setenv", "XDG_DOCUMENTS_DIR",   filepath.Join(homed, "Documents"),
			"--setenv", "XDG_MUSIC_DIR",       filepath.Join(homed, "Music"),
			"--setenv", "XDG_PICTURES_DIR",    filepath.Join(homed, "Pictures"),
			"--setenv", "XDG_VIDEOS_DIR",      filepath.Join(homed, "Videos"),
			"--setenv", "XDG_TEMPLATES_DIR",   filepath.Join(homed, "Templates"),
			"--setenv", "XDG_PUBLICSHARE_DIR", filepath.Join(homed, "Share"),
			"--setenv", "XDG_DATA_HOME",       filepath.Join(homed, ".local/share"),
			"--setenv", "XDG_CONFIG_HOME",     filepath.Join(homed, ".config"),
			"--setenv", "XDG_CACHE_HOME",      filepath.Join(homed, ".cache"),
			"--setenv", "XDG_STATE_HOME",      filepath.Join(homed, ".local/state"),
			"--setenv", "LOGNAME",             usern,
			"--setenv", "USER",                usern,
			"--uid",    uid,
			"--die-with-parent",
			"--new-session",
			"--dir",         filepath.Join("/run/user", uid),
			"--dev",         "/dev",
			"--proc",        "/proc",
			"--tmpfs",       filepath.Join(homed, ".cache"),
			"--ro-bind",     aiRoot(ai, "opt"),       "/opt",
			"--ro-bind",     aiRoot(ai, "bin"),       "/bin",
			"--ro-bind",     aiRoot(ai, "sbin"),      "/sbin",
			"--ro-bind",     aiRoot(ai, "lib"),       "/lib",
			"--ro-bind-try", aiRoot(ai, "lib32"),     "/lib32",
			"--ro-bind-try", aiRoot(ai, "lib64"),     "/lib64",
			"--ro-bind",     aiRoot(ai, "usr/bin"),   "/usr/bin",
			"--ro-bind",     aiRoot(ai, "usr/sbin"),  "/usr/sbin",
			"--ro-bind",     aiRoot(ai, "usr/lib"),   "/usr/lib",
			"--ro-bind-try", aiRoot(ai, "usr/lib32"), "/usr/lib32",
			"--ro-bind-try", aiRoot(ai, "usr/lib64"), "/usr/lib64",
	}

	// Convert device perms to bwrap format
	for _, v := range(ai.Perms.Devices) {
		if len(v) < 5 || v[0:5] != "/dev/" {
			v = filepath.Join("/dev", v)
		}

		cmdArgs = append(cmdArgs, "--dev-bind-try", v, v)
	}

	// Convert requested files/ dirs to brap flags
	for _, val := range(ai.Perms.Files) {
		s   := strings.Split(val, ":")
		ex  := s[len(s)-1]
		if ex == "rw" {
			cmdArgs = append(cmdArgs, "--bind-try", ExpandDir(val), ExpandGenericDir(val))
		} else if ex == "ro" {
			cmdArgs = append(cmdArgs, "--ro-bind-try", ExpandDir(val), ExpandGenericDir(val))
		}
	}

	// Level 1 is minimal sandboxing, grants access to most system files, all devices and only really attempts to isolate home files
	if ai.Perms.Level == 1 {
		cmdArgs = append(cmdArgs, []string{
			"--dev-bind",    "/dev", "/dev",
			"--ro-bind",     "/sys", "/sys",
			"--ro-bind",     aiRoot(ai, "usr"),    "/usr",
			"--ro-bind-try", aiRoot(ai, "etc"),    "/etc",
			"--ro-bind-try", aiRoot(ai, ".fonts"),     filepath.Join(homed, ".fonts"),
			"--ro-bind-try", aiRoot(ai, "fontconfig"), filepath.Join(homed, ".config/fontconfig"),
			"--ro-bind-try", aiRoot(ai, "gtk-3.0"),    filepath.Join(homed, ".config/gtk-3.0"),
		}...)
	// Level 2 grants access to fewer system files, and all themes
	// Likely to add more files here for compatability.
	// This should be the standard level for GUI profiles
	} else if ai.Perms.Level == 2 {
		cmdArgs = append(cmdArgs, []string{
			// Testing removal of `/sys` from level 2, it ideally shouldn't be
			// here
//			"--ro-bind",     "/sys", "/sys",
			"--ro-bind-try", aiRoot(ai, "etc/fonts"),              "/etc/fonts",
			"--ro-bind-try", aiRoot(ai, "usr/share/fontconfig"),   "/usr/share/fontconfig",
			"--ro-bind-try", aiRoot(ai, "usr/share/fonts"),        "/usr/share/fonts",
			"--ro-bind-try", aiRoot(ai, "usr/share/icons"),        "/usr/share/icons",
			"--ro-bind-try", aiRoot(ai, "usr/share/themes"),       "/usr/share/themes",
			"--ro-bind-try", aiRoot(ai, "usr/share/applications"), "/usr/share/applications",
			"--ro-bind-try", aiRoot(ai, "usr/share/mime"),         "/usr/share/mime",
			"--ro-bind-try", aiRoot(ai, "usr/share/libdrm"),       "/usr/share/librdm",
			"--ro-bind-try", aiRoot(ai, "usr/share/glvnd"),        "/usr/share/glvnd",
			"--ro-bind-try", aiRoot(ai, "usr/share/glib-2.0"),     "/usr/share/glib-2.0",
			"--ro-bind-try", filepath.Join(xdg.Home,       ".fonts"),     filepath.Join(homed, ".fonts"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "fontconfig"), filepath.Join(homed, ".config/fontconfig"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "gtk-3.0"),    filepath.Join(homed, ".config/gtk-3.0"),
		}...)
	}

	// These vars will only be used if x11 socket is granted access
	xAuthority := os.Getenv("XAUTHORITY")
	xDisplay := strings.ReplaceAll(os.Getenv("DISPLAY"), ":", "")

	// Args if socket is enabled
	var sockets = map[string][]string {
		"cgroup": {},
		"ipc":    {},
		"network": {
				"--share-net",
				"--ro-bind-try", "/etc/ca-certificates", "/etc/ca-certificates",
				"--ro-bind",     "/etc/resolv.conf",     "/etc/resolv.conf",
				"--ro-bind-try", "/etc/ssl",             "/etc/ssl",
		},
		"pid": {},
		"pulseaudio": {
			"--ro-bind-try", "/run/user/"+ruid+"/pulse", "/run/user/"+ruid+"/pulse",
			"--ro-bind-try", "/usr/share/alsa",          "/usr/share/alsa",
		},
		"user": {},
		"uts":  {},
		"wayland": {
			"--ro-bind-try", "/run/user/"+ruid+"/wayland-0", "/run/user/"+ruid+"/wayland-0",
			"--ro-bind-try", "/usr/share/X11",               "/usr/share/X11",
		},
		// For some reason sometimes it doesn't work when binding X0 to another
		// socket ...but sometimes it does. X11 should be avoided if looking
		// for security anyway, as it easilly allows control of the keyboard
		// and mouse
		"x11": {
			"--ro-bind-try", xAuthority,                      homed+"/.Xauthority",
			"--ro-bind-try", sysTemp+"/.X11-unix/X"+xDisplay, "/tmp/.X11-unix/X"+xDisplay,
			"--ro-bind-try", "/usr/share/X11",                "/usr/share/X11",
			"--setenv",      "XAUTHORITY",                    homed+"/.Xauthority",
			"--setenv",      "DISPLAY",                       ":"+xDisplay,
		},
	}

	// Args to disable sockets if not requested
	var unsocks = map[string][]string {
		"cgroup":     { "--unshare-cgroup-try" },
		"ipc":        { "--unshare-ipc" },
		"network":    { "--unshare-net" },
		"pid":        { "--unshare-pid" },
		"pulseaudio": {},
		"user":       { "--unshare-user-try" },
		"uts":        { "--unshare-uts" },
		"wayland":    {},
		"x11":        {},
	}

	for s, _ := range sockets {
		_, present := helpers.Contains(ai.Perms.Sockets, s)
		if present {
			cmdArgs = append(cmdArgs, sockets[s]...)
		} else {
			cmdArgs = append(cmdArgs, unsocks[s]...)
		}
	}

	// Give access to all files needed to run device
	var devices = map[string][]string {
		"dri": {
			"--ro-bind",      "/sys/dev/char",           "/sys/dev/char",
			"--ro-bind",      "/sys/devices/pci0000:00", "/sys/devices/pci0000:00",
			"--dev-bind-try", "/dev/nvidiactl",          "/dev/nvidiactl",
			"--dev-bind-try", "/dev/nvidia0",            "/dev/nvidia0",
		},
		"input": {
			"--ro-bind",      "/sys/class/input",           "/sys/class/input",
		},
	}

	for device, _ := range(devices) {
		_, present := helpers.Contains(ai.Perms.Devices, device)
		if present {
			cmdArgs = append(cmdArgs, devices[device]...)
		}
	}

	return cmdArgs
}

// Expands XDG formatted directories into full paths depending on the input map
func expandEither(str string, xdgDirs map[string]string) string {
	for key, val := range xdgDirs {
		// If length of key bigger than requested directory or not equal to it
		// continue because there is no reason to look at it further
		if len(key) > len(str) || key != str[:len(key)] {
			continue
		}

		// The final byte of the key (used for splitting)
		c := str[len(key)]
		if c == byte('/') || c == byte(':') {
			str = strings.Replace(str, key, val, 1)
			break
		}
	}

	s   := strings.Split(str, ":")
	dir := strings.Join(s[:len(s)-1], ":")

	// Resolve `../` and clean up extra slashes if they exist
	str = filepath.Clean(dir)

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

// Returns the location of the requested directory on the host filesystem with
// symlinks resolved. This should solve systems like GoboLinux, where
// traditionally named directories are symlinks to something unconventional.
func aiRoot(ai *AppImage, src string) string {
	s, _ := filepath.EvalSymlinks(filepath.Join(ai.rootDir, src))

	return s
}
