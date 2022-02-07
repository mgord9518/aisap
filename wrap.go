package aisap

import (
	"errors"
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
func Run(ai *AppImage, args []string) error {
	err := setupRun(ai)
	if err != nil { return err }

	cmd := exec.Command(filepath.Join(ai.mountDir, "AppRun"), args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin  = os.Stdin

	return cmd.Run()
}

// Executes AppImage through bwrap, fails if `ai.Perms.Level` < 1
func Sandbox(ai *AppImage, args []string) error {
	bwrapArgs, err := GetWrapArgs(ai)
	if err != nil { return err }

	if _, err := exec.LookPath("bwrap"); err != nil {
		return errors.New("bubblewrap not found! It's required to use sandboxing")
	}

	err = setupRun(ai)
	if err != nil { return err }

	// Bind the fake `~` and `/tmp` dirs
	bwrapArgs = append([]string{
		"--bind",   ai.dataDir, xdg.Home,
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

	if !helpers.DirExists(filepath.Join(xdg.CacheHome, "appimagekit_" + ai.md5)) {
		err := os.MkdirAll(filepath.Join(xdg.CacheHome, "appimagekit_" + ai.md5), 0744)
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
	os.Setenv("TMPDIR",         ai.tempDir)
	os.Setenv("HOME",           ai.dataDir)
	os.Setenv("APPDIR",         ai.mountDir)
	os.Setenv("APPIMAGE",       ai.Path)
	os.Setenv("ARGV0",          ai.Path)
	os.Setenv("XDG_CACHE_HOME", ai.Path)

	return err
}

func GetWrapArgs(ai *AppImage) ([]string, error) {
	uid := strconv.Itoa(os.Getuid())
	// Basic arguments to be used at all sandboxing levels
	cmdArgs := []string{
			"--setenv", "TMPDIR",              "/tmp",
			"--setenv", "HOME",                xdg.Home,
			"--setenv", "APPIMAGE",            filepath.Join("/app", path.Base(ai.Path)),
			"--setenv", "ARGV0",               filepath.Join("/app", path.Base(ai.Path)),
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
			"--die-with-parent",
			"--new-session",
			"--dir",         filepath.Join("/run/user", uid),
			"--dev",         "/dev",
			"--proc",        "/proc",
			"--bind",        filepath.Join(xdg.Home, ".cache", "appimagekit_" + ai.md5), filepath.Join(xdg.Home, ".cache"),
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
			"--dir",         "/app",
			"--bind",        ai.Path, filepath.Join("/app", path.Base(ai.Path)),
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
		dir := strings.Join(s[:len(s)-1], ":")

		if ex == "rw" {
			cmdArgs = append(cmdArgs, "--bind-try", helpers.ExpandDir(dir), helpers.ExpandGenericDir(dir))
		} else if ex == "ro" {
			cmdArgs = append(cmdArgs, "--ro-bind-try", helpers.ExpandDir(dir), helpers.ExpandGenericDir(dir))
		}
	}

	// Level 1 is minimal sandboxing, grants access to most system files, all devices and only really attempts to isolate home files
	if ai.Perms.Level == 1 {
		cmdArgs = append(cmdArgs, []string{
			"--dev-bind",    "/dev", "/dev",
			"--ro-bind",     "/sys", "/sys",
			"--ro-bind",     aiRoot(ai, "usr"), "/usr",
			"--ro-bind-try", aiRoot(ai, "etc"), "/etc",
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
			"--ro-bind-try", filepath.Join(xdg.Home,       ".fonts"),     filepath.Join(xdg.Home, ".fonts"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "fontconfig"), filepath.Join(xdg.Home, ".config/fontconfig"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "gtk-3.0"),    filepath.Join(xdg.Home, ".config/gtk-3.0"),
			"--ro-bind-try", filepath.Join(xdg.ConfigHome, "kdeglobals"), filepath.Join(xdg.Home, ".config/kdeglobals"),
		}...)
	} else if ai.Perms.Level > 3 || ai.Perms.Level < 1 {
		return []string{}, errors.New("AppImage permissions level does not allow sandboxing")
	}

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
			"--ro-bind-try", "/run/user/"+uid+"/pulse", "/run/user/"+uid+"/pulse",
			"--ro-bind-try", "/usr/share/alsa",         "/usr/share/alsa",
			"--ro-bind-try", "/usr/share/pulseaudio",   "/usr/share/pulseaudio",
			"--ro-bind-try", "/etc/alsa",               "/etc/alsa",
			"--ro-bind-try", "/etc/group",              "/etc/group",
			"--ro-bind-try", "/etc/pulse",              "/etc/pulse",
			"--dev-bind",    "/dev/snd",                "/dev/snd",
		},
		"cgroup": {},
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
			"--ro-bind-try", "/run/user/"+uid+"/pipewire-0", "/run/user/"+uid+"/pipewire-0",
		},
		"pulseaudio": {
			"--ro-bind-try", "/run/user/"+uid+"/pulse", "/run/user/"+uid+"/pulse",
			"--ro-bind-try", "/etc/pulse",              "/etc/pulse",
		},
		"user": {},
		"uts":  {},
		"wayland": {
			"--ro-bind-try", "/run/user/"+uid+"/"+wDisplay, "/run/user/"+uid+"/wayland-0",
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

	// Args to disable sockets if not requested
	var unsocks = map[string][]string {
		"alsa":       {},
		"audio":      {},
		"cgroup":     { "--unshare-cgroup-try" },
		"ipc":        { "--unshare-ipc" },
		"network":    { "--unshare-net" },
		"pid":        { "--unshare-pid" },
		"pipewire":   {},
		"pulseaudio": {},
		"user":       { "--unshare-user-try" },
		"uts":        { "--unshare-uts" },
		"wayland":    {},
		"x11":        {},
	}

	for s, _ := range(sockets) {
		if _, present := helpers.Contains(ai.Perms.Sockets, s); present {
			// Don't give access to X11 if wayland is running on the machine
			// and the app supports it
			if _, waylandApp := helpers.Contains(ai.Perms.Sockets, "wayland");
			waylandEnabled && waylandApp && s == "x11" {
				continue
			}

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
			"--dev-bind-try", "/dev/nvidia-modeset",     "/dev/nvidia-modeset",
			"--ro-bind-try",  aiRoot(ai, "usr/share/glvnd"), "/usr/share/glvnd",
		},
		"input": {
			"--ro-bind", "/sys/class/input", "/sys/class/input",
		},
	}

	for device, _ := range(devices) {
		if _, present := helpers.Contains(ai.Perms.Devices, device); present {
			cmdArgs = append(cmdArgs, devices[device]...)
		}
	}

	return cmdArgs, nil
}

// Returns the location of the requested directory on the host filesystem with
// symlinks resolved. This should solve systems like GoboLinux, where
// traditionally named directories are symlinks to something unconventional.
func aiRoot(ai *AppImage, src string) string {
	s, _ := filepath.EvalSymlinks(filepath.Join(ai.rootDir, src))

	return s
}
