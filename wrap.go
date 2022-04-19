package aisap

import (
	"errors"
	"bytes"
	"bufio"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	helpers "github.com/mgord9518/aisap/helpers"
	xdg     "github.com/adrg/xdg"
)

// Run the AppImage with appropriate sandboxing. If `ai.Perms.Level` == 0, use
// no sandbox. If > 0, sandbox
func (ai *AppImage) Run(args []string) error {
	if ai.Perms.Level > 0 {
		return ai.Sandbox(args)
	} else if ai.Perms.Level < 0 {
		return errors.New("invalid permissions level!")
	}

	err := ai.setupRun()
	if err != nil { return err }

	cmd := exec.Command(filepath.Join(ai.mountDir, "AppRun"), args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin  = os.Stdin

	return cmd.Run()
}

// Executes AppImage through bwrap, fails if `ai.Perms.Level` < 1
// Also automatically creates a portable home
func (ai *AppImage) Sandbox(args []string) error {
	if ai.dataDir == "" {
		ai.dataDir = ai.Path + ".home"
	}

	cmdArgs, err := ai.WrapArgs(args)
	if err != nil { return err }

	bwrapStr, present := helpers.CommandExists("bwrap")
	if !present {
		return errors.New("failed to find bwrap! unable to sandbox application")
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

	bwrap := exec.Command(bwrapStr, cmdArgs...)
	bwrap.Stdout = os.Stdout
	bwrap.Stderr = os.Stderr
	bwrap.Stdin  = os.Stdin

	return bwrap.Run()
}

func (ai *AppImage) setupRun() error {
	if !ai.IsMounted() {
		return errors.New("AppImage must be mounted before running! call *AppImage.Mount() first")
	}

	if !helpers.DirExists(filepath.Join(xdg.CacheHome, "appimage", ai.md5)) {
		err := os.MkdirAll(filepath.Join(xdg.CacheHome, "appimage", ai.md5), 0744)
		if err != nil { return err }
	}

	// Set required vars to correctly mount our target AppImage
	// If sandboxed, these values will be overwritten
	os.Setenv("TMPDIR",         ai.tempDir)
	os.Setenv("APPDIR",         ai.mountDir)
	os.Setenv("APPIMAGE",       ai.Path)
	os.Setenv("ARGV0",          path.Base(ai.Path))
	os.Setenv("XDG_CACHE_HOME", filepath.Join(xdg.CacheHome, "appimage", ai.md5))

	return nil
}

// Returns the bwrap arguments to sandbox the AppImage
func (ai AppImage) WrapArgs(args []string) ([]string, error) {
	if !ai.IsMounted() {
		return []string{}, errors.New("AppImage must be mounted before getting its wrap arguments! call *AppImage.Mount() first")
	}

	home, present := unsetHome()
	defer restoreHome(home, present)

	if ai.Perms.Level == 0 { return args, nil }

	cmdArgs := ai.mainWrapArgs()

	err := ai.setupRun()
	if err != nil { return []string{}, err }

	cmdArgs = append([]string{
		"--bind",   ai.dataDir, xdg.Home,
		"--setenv", "APPDIR",   "/tmp/.mount_"+ai.runId,
	}, cmdArgs...)

	cmdArgs = append(cmdArgs, "--",
		"/tmp/.mount_"+ai.runId+"/AppRun",
	)

	// Append console arguments provided by the user
	return append(cmdArgs, args...), nil
}

func (ai *AppImage) mainWrapArgs() []string {
	uid := strconv.Itoa(os.Getuid())
	home, present := unsetHome()
	defer restoreHome(home, present)

	// Basic arguments to be used at all sandboxing levels
	cmdArgs := []string{
		"--setenv", "TMPDIR",              "/tmp",
		"--setenv", "HOME",                xdg.Home,
		"--setenv", "APPIMAGE",            filepath.Join("/app", path.Base(ai.Path)),
		"--setenv", "ARGV0",               filepath.Join(path.Base(ai.Path)),
		"--setenv", "XDG_DESKTOP_DIR",     filepath.Join(xdg.Home, "Desktop"),
		"--setenv", "XDG_DOWNLOAD_DIR",    filepath.Join(xdg.Home, "Downloads"),
		"--setenv", "XDG_DOCUMENTS_DIR",   filepath.Join(xdg.Home, "Documents"),
		"--setenv", "XDG_MUSIC_DIR",       filepath.Join(xdg.Home, "Music"),
		"--setenv", "XDG_PICTURES_DIR",    filepath.Join(xdg.Home, "Pictures"),
		"--setenv", "XDG_VIDEOS_DIR",      filepath.Join(xdg.Home, "Videos"),
		"--setenv", "XDG_TEMPLATES_DIR",   filepath.Join(xdg.Home, "Templates"),
		"--setenv", "XDG_PUBLICSHARE_DIR", filepath.Join(xdg.Home, "Share"),
		"--setenv", "XDG_DATA_HOME",       filepath.Join(xdg.Home, ".local/share"),
		"--setenv", "XDG_CONFIG_HOME",     filepath.Join(xdg.Home, ".config"),
		"--setenv", "XDG_CACHE_HOME",      filepath.Join(xdg.Home, ".cache"),
		"--setenv", "XDG_STATE_HOME",      filepath.Join(xdg.Home, ".local/state"),
		"--setenv", "XDG_RUNTIME_DIR",     filepath.Join("/run/user", uid),
		"--die-with-parent",
		"--perms",       "0700",
		"--dir",         filepath.Join("/run/user", uid),
		"--dev",         "/dev",
		"--proc",        "/proc",
		"--bind",        filepath.Join(xdg.CacheHome, "appimage", ai.md5), filepath.Join(xdg.Home, ".cache"),
		"--ro-bind",     ai.resolve("opt"),       "/opt",
		"--ro-bind",     ai.resolve("bin"),       "/bin",
		"--ro-bind",     ai.resolve("sbin"),      "/sbin",
		"--ro-bind",     ai.resolve("lib"),       "/lib",
		"--ro-bind-try", ai.resolve("lib32"),     "/lib32",
		"--ro-bind-try", ai.resolve("lib64"),     "/lib64",
		"--ro-bind",     ai.resolve("usr/bin"),   "/usr/bin",
		"--ro-bind",     ai.resolve("usr/sbin"),  "/usr/sbin",
		"--ro-bind",     ai.resolve("usr/lib"),   "/usr/lib",
		"--ro-bind-try", ai.resolve("usr/lib32"), "/usr/lib32",
		"--ro-bind-try", ai.resolve("usr/lib64"), "/usr/lib64",
		"--dir",         "/app",
		"--bind",        ai.Path, filepath.Join("/app", path.Base(ai.Path)),
	}

	// Level 1 is minimal sandboxing, grants access to most system files, all devices and only really attempts to isolate home files
	if ai.Perms.Level == 1 {
		cmdArgs = append(cmdArgs, []string{
			"--dev-bind",    "/dev", "/dev",
			"--ro-bind",     "/sys", "/sys",
			"--ro-bind",     ai.resolve("usr"), "/usr",
			"--ro-bind-try", ai.resolve("etc"), "/etc",
			"--ro-bind-try", filepath.Join(xdg.Home,       ".fonts"),     filepath.Join(xdg.Home, ".fonts"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "fontconfig"), filepath.Join(xdg.Home, ".config/fontconfig"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "gtk-3.0"),    filepath.Join(xdg.Home, ".config/gtk-3.0"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "kdeglobals"), filepath.Join(xdg.Home, ".config/kdeglobals"),
		}...)
	// Level 2 grants access to fewer system files, and all themes
	// Likely to add more files here for compatability.
	// This should be the standard level for GUI profiles
	} else if ai.Perms.Level == 2 {
		cmdArgs = append(cmdArgs, []string{
			"--ro-bind-try", ai.resolve("etc/fonts"),              "/etc/fonts",
			"--ro-bind-try", ai.resolve("etc/ld.so.cache"),        "/etc/ld.so.cache",
			"--ro-bind-try", ai.resolve("etc/mime.types"),         "/etc/mime.types",
			"--ro-bind-try", ai.resolve("etc/xdg"),                "/etc/xdg",
			"--ro-bind-try", ai.resolve("usr/share/fontconfig"),   "/usr/share/fontconfig",
			"--ro-bind-try", ai.resolve("usr/share/fonts"),        "/usr/share/fonts",
			"--ro-bind-try", ai.resolve("usr/share/icons"),        "/usr/share/icons",
			"--ro-bind-try", ai.resolve("usr/share/themes"),       "/usr/share/themes",
			"--ro-bind-try", ai.resolve("usr/share/applications"), "/usr/share/applications",
			"--ro-bind-try", ai.resolve("usr/share/mime"),         "/usr/share/mime",
			"--ro-bind-try", ai.resolve("usr/share/libdrm"),       "/usr/share/librdm",
			"--ro-bind-try", ai.resolve("usr/share/glvnd"),        "/usr/share/glvnd",
			"--ro-bind-try", ai.resolve("usr/share/glib-2.0"),     "/usr/share/glib-2.0",
			"--ro-bind-try", ai.resolve("usr/share/terminfo"),     "/usr/share/terminfo",
			"--ro-bind-try", filepath.Join(xdg.Home,       ".fonts"),     filepath.Join(xdg.Home, ".fonts"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "fontconfig"), filepath.Join(xdg.Home, ".config/fontconfig"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "gtk-3.0"),    filepath.Join(xdg.Home, ".config/gtk-3.0"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "kdeglobals"), filepath.Join(xdg.Home, ".config/kdeglobals"),
		}...)
	} else if ai.Perms.Level > 3 || ai.Perms.Level < 1 {
		return []string{}
	}

	cmdArgs = append(cmdArgs, parseFiles(ai)...)
	cmdArgs = append(cmdArgs, parseSockets(ai)...)
	cmdArgs = append(cmdArgs, parseDevices(ai)...)

	// Only supply libraries that aren't present on the host system to the
	// sandbox. This needs more work (eg: move whole directories over if they
	// contain no libs), but for some AppImages it may reduce RAM usage, reduce
	// launch time and significantly speed up execution in emulating a
	// different architecture (although this isn't a common thing)
	_, present = os.LookupEnv("PREFER_SYSTEM_LIBRARIES")
	if present {
		// Get a list of the system libraries into a string by running ldconfig
		r := &bytes.Buffer{}
		cmd := exec.Command("ldconfig", "-p")
		cmd.Stdout = r
		cmd.Run()
		scanner := bufio.NewScanner(r)
		sysLibs := []string{}
		for scanner.Scan() {
			s := strings.Split(scanner.Text(), " ")
			sysLibs = append(sysLibs, s[len(s)-1])
		}

		filepath.Walk(ai.mountDir, func(dir string, info os.FileInfo, err error) error {
			if err != nil { return err }
			if info.IsDir() {
				return nil
			}

			var nDir string
			foundLib := false
			for _, val := range(sysLibs) {
				if path.Base(val) == path.Base(dir) {
					nDir = val
					foundLib = true
				}
			}

			if !foundLib {
				nDir = strings.Replace(dir, ai.mountDir, "", 1)
			}

			cmdArgs = append([]string{
				"--ro-bind", dir, "/tmp/.mount_" + ai.runId+nDir,
			}, cmdArgs...)

			return nil
		})
		cmdArgs = append([]string{
			"--dir", "/tmp/.mount_" + ai.runId,
		}, cmdArgs...)
	} else {
		cmdArgs = append([]string{
			"--bind",   ai.tempDir, "/tmp",
		}, cmdArgs...)
	}

	return cmdArgs
}

// Returns the location of the requested directory on the host filesystem with
// symlinks resolved. This should solve systems like GoboLinux, where
// traditionally named directories are symlinks to something unconventional.
func (ai *AppImage) resolve(src string) string {
	s, _ := filepath.EvalSymlinks(filepath.Join(ai.rootDir, src))

	if s == "" {
		s = "/" + src
	}

	return s
}

func parseFiles(ai *AppImage) []string {
	var s []string

	// Convert requested files/ dirs to brap flags
	for _, val := range(ai.Perms.Files) {
		sl  := strings.Split(val, ":")
		ex  := sl[len(sl)-1]
		dir := strings.Join(sl[:len(sl)-1], ":")

		if ex == "rw" {
			s = append(s, "--bind-try", helpers.ExpandDir(dir), helpers.ExpandGenericDir(dir))
		} else if ex == "ro" {
			s = append(s, "--ro-bind-try", helpers.ExpandDir(dir), helpers.ExpandGenericDir(dir))
		}
	}

	return s
}

// Give all requried flags to add the devices
func parseDevices(ai *AppImage) []string {
	var d []string

	// Convert device perms to bwrap format
	for _, v := range(ai.Perms.Devices) {
		if len(v) < 5 || v[0:5] != "/dev/" {
			v = filepath.Join("/dev", v)
		}

		d = append(d, "--dev-bind-try", v, v)
	}

	// Required files to go along with them
	var devices = map[string][]string {
		"dri": {
			"--ro-bind",      "/sys/dev/char",           "/sys/dev/char",
			"--ro-bind",      "/sys/devices/pci0000:00", "/sys/devices/pci0000:00",
			"--dev-bind-try", "/dev/nvidiactl",          "/dev/nvidiactl",
			"--dev-bind-try", "/dev/nvidia0",            "/dev/nvidia0",
			"--dev-bind-try", "/dev/nvidia-modeset",     "/dev/nvidia-modeset",
			"--ro-bind-try",  ai.resolve("usr/share/glvnd"), "/usr/share/glvnd",
		},
		"input": {
			"--ro-bind", "/sys/class/input", "/sys/class/input",
		},
	}

	for device, _ := range(devices) {
		if _, present := helpers.Contains(ai.Perms.Devices, device); present {
			d = append(d, devices[device]...)
		}
	}

	return d
}

func parseSockets(ai *AppImage) []string {
	var s []string
	uid := strconv.Itoa(os.Getuid())

	// These vars will only be used if x11 socket is granted access
	xAuthority := os.Getenv("XAUTHORITY")
	xDisplay := strings.ReplaceAll(os.Getenv("DISPLAY"), ":", "")
	tempDir, present := os.LookupEnv("TMPDIR")
	if !present {
		tempDir = "/tmp"
	}

	// Set if Wayland is running on the host machine
	// Using different Wayland display sessions currently not tested
	wDisplay, waylandEnabled := os.LookupEnv("WAYLAND_DISPLAY")

	// Args if socket is enabled
	var sockets = map[string][]string {
		// Encompasses ALSA, Pulse and pipewire. Easiest for convience, but for
		// more security, specify the specific audio system
		"alsa": {
			"--ro-bind-try", "/usr/share/alsa", "/usr/share/alsa",
			"--ro-bind-try", "/etc/alsa",       "/etc/alsa",
			"--ro-bind-try", "/etc/group",      "/etc/group",
			"--dev-bind",    "/dev/snd",        "/dev/snd",
		},
		"audio": {
			"--ro-bind-try", filepath.Join(xdg.RuntimeDir, "pulse"), "/run/user/"+uid+"/pulse",
			"--ro-bind-try", "/usr/share/alsa",         "/usr/share/alsa",
			"--ro-bind-try", "/usr/share/pulseaudio",   "/usr/share/pulseaudio",
			"--ro-bind-try", "/etc/alsa",               "/etc/alsa",
			"--ro-bind-try", "/etc/group",              "/etc/group",
			"--ro-bind-try", "/etc/pulse",              "/etc/pulse",
			"--dev-bind",    "/dev/snd",                "/dev/snd",
		},
		"cgroup": {},
		"dbus": {
			"--ro-bind-try", filepath.Join(xdg.RuntimeDir, "bus"), "/run/user/"+uid+"/bus",
		},
		"ipc":    {},
		"network": {
				"--share-net",
				"--ro-bind-try", "/etc/ca-certificates",       "/etc/ca-certificates",
				"--ro-bind",     "/etc/resolv.conf",           "/etc/resolv.conf",
				"--ro-bind-try", "/etc/ssl",                   "/etc/ssl",
				"--ro-bind-try", "/usr/share/ca-certificates", "/usr/share/ca-certificates",
		},
		"pid": {},
		"pipewire": {
			"--ro-bind-try", filepath.Join(xdg.RuntimeDir, "pipewire-0"), "/run/user/"+uid+"/pipewire-0",
		},
		"pulseaudio": {
			"--ro-bind-try", filepath.Join(xdg.RuntimeDir, "pulse"), "/run/user/"+uid+"/pulse",
			"--ro-bind-try", "/etc/pulse",              "/etc/pulse",
		},
		"session": {},
		"user":    {},
		"uts":     {},
		"wayland": {
			"--ro-bind-try", filepath.Join(xdg.RuntimeDir, wDisplay), "/run/user/"+uid+"/wayland-0",
			"--ro-bind-try", "/usr/share/X11",               "/usr/share/X11",
			// TODO: Add more enviornment variables for app compatability
			// maybe theres a better way to do this?
			"--setenv", "WAYLAND_DISPLAY",             "wayland-0",
			"--setenv", "_JAVA_AWT_WM_NONREPARENTING", "1",
			"--setenv", "MOZ_ENABLE_WAYLAND",          "1",
			"--setenv", "XDG_SESSION_TYPE",            "wayland",
		},
		// For some reason sometimes it doesn't work when binding X0 to another
		// socket ...but sometimes it does. X11 should be avoided if looking
		// for security anyway, as it easilly allows control of the keyboard
		// and mouse
		"x11": {
			"--ro-bind-try", xAuthority,                      xdg.Home+"/.Xauthority",
			"--ro-bind-try", tempDir+"/.X11-unix/X"+xDisplay, "/tmp/.X11-unix/X"+xDisplay,
			"--ro-bind-try", "/usr/share/X11",                "/usr/share/X11",
			"--setenv",      "DISPLAY",         ":"+xDisplay,
			"--setenv",      "QT_QPA_PLATFORM", "xcb",
			"--setenv",      "XAUTHORITY",      xdg.Home+"/.Xauthority",
		},
	}

	// Args to disable sockets if not given
	var unsocks = map[string][]string {
		"alsa":       {},
		"audio":      {},
		"cgroup":     { "--unshare-cgroup-try" },
		"ipc":        { "--unshare-ipc" },
		"network":    { "--unshare-net" },
		"pid":        { "--unshare-pid" },
		"pipewire":   {},
		"pulseaudio": {},
		"session":    { "--new-session" },
		"user":       { "--unshare-user-try" },
		"uts":        { "--unshare-uts" },
		"wayland":    {},
		"x11":        {},
	}

	for soc, _ := range(sockets) {
		if _, present := helpers.Contains(ai.Perms.Sockets, soc); present {
			// Don't give access to X11 if wayland is running on the machine
			// and the app supports it
			if _, waylandApp := helpers.Contains(ai.Perms.Sockets, "wayland");
			waylandEnabled && waylandApp && soc == "x11" {
				continue
			}
			s = append(s, sockets[soc]...)
		} else {
			s = append(s, unsocks[soc]...)
		}
	}

	return s
}

// Unset HOME in case the program using aisap is an AppImage using a portable
// home. This is done because aisap needs access to the acual XDG directories
// to share them. Otherwise, an AppImage requesting `xdg-download` would be
// given the "Download" directory inside of aisap's portable home
func unsetHome() (string, bool) {
	home, present := os.LookupEnv("HOME")

	newHome, _ := helpers.RealHome()

	os.Setenv("HOME", newHome)
	xdg.Reload()

	return home, present 
}

// Return the HOME variable to normal
func restoreHome(home string, present bool) {
	if present {
		os.Setenv("HOME", home)
	}

	xdg.Reload()
}
