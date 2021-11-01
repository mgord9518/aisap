package aisap

import (
	"errors"
	"os/exec"
	"strings"
	"os"
	"strconv"

	helpers  "github.com/mgord9518/aisap/helpers"
	profiles "github.com/mgord9518/aisap/profiles"
	xdg	     "github.com/adrg/xdg"
)

// Run the AppImage with zero sandboxing
func Run(ai *AppImage, args []string) error {
	if dataDir == "" {
		dataDir = ai.Path+".home"
	}

	if !helpers.DirExists(dataDir) {
		err := os.MkdirAll(dataDir, 0744)
		if err != nil { return err }
	}

	if !helpers.DirExists(dataDir+"/.local/share/appimagekit/") {
		err := os.MkdirAll(dataDir+"/.local/share/appimagekit/", 0744)
		if err != nil { return err }
	}

	// Tell AppImages not to ask for integration
	noIntegrate, err := os.Create(dataDir+"/.local/share/appimagekit/no_desktopintegration")
	noIntegrate.Close()
	if err != nil { return err }

	// Set required vars to correctly mount our target AppImage
	os.Setenv("TMPDIR", ai.TempDir())
	os.Setenv("HOME",   dataDir)
//	mntDir := ai.MountDir()
	if err != nil { return err }

	cmd := exec.Command(ai.MountDir()+"/AppRun", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin  = os.Stdin
	cmd.Start()
	err = cmd.Wait()
	//mnt.Process.Signal(syscall.SIGINT)
	if err != nil { return err }

	// Clean up after the app is closed
	// Sleep is needed to wait until the AppImage is unmounted before deleting the temporary dir
	err = UnmountAppImage(ai)
	if err != nil {return err}
	err = os.RemoveAll(ai.TempDir())
	return err
}

// Wrap is a re-implementation of the aibwrap shell script, allowing execution of AppImages through bwrap
func Wrap(ai *AppImage, perms *profiles.AppImagePerms, args []string) error {
	//runId := helpers.RandString(int(time.Now().UTC().UnixNano()), 8)

	bwrapArgs := GetWrapArgs(perms)

	if _, err := exec.LookPath("bwrap"); err != nil {
		err := errors.New("Bubblewrap not found! It is required to use aisap.Wrap()")
		return err
	}

	if dataDir == "" {
		dataDir = ai.Path+".home"
	}

	if !helpers.DirExists(dataDir) {
		err := os.MkdirAll(dataDir, 0744)
		if err != nil { return err }
	}

	if !helpers.DirExists(dataDir+"/.local/share/appimagekit/") {
		err := os.MkdirAll(dataDir+"/.local/share/appimagekit/", 0744)
		if err != nil { return err }
	}

	// Tell AppImages not to ask for integration
	noIntegrate, err := os.Create(dataDir+"/.local/share/appimagekit/no_desktopintegration")
	noIntegrate.Close()
	if err != nil { return err }

	// Set required vars to correctly mount our target AppImage
	os.Setenv("TMPDIR", ai.TempDir())
	//mntDir, _ := helpers.MakeTemp(runDir, ".mount_"+runId)
	//mntDir := ai.TempDir()+"/.mount_"+ai.RunId()
	//err = MountAppImage(ai.Path, mntDir)
	if err != nil { return err }

	// Bind the fake /home and /tmp dirs
	bwrapArgs = append([]string{
		"--bind", dataDir, "/home/"+usern,
		"--bind", ai.TempDir(), "/tmp",
	}, bwrapArgs...)

	bwrapArgs = append(bwrapArgs, "--",
		"/tmp/.mount_"+ai.RunId()+"/AppRun",
	)

	bwrapArgs = append(bwrapArgs, args...)

	bwrap := exec.Command("bwrap", bwrapArgs...)

	bwrap.Stdout = os.Stdout
	bwrap.Stderr = os.Stderr
	bwrap.Stdin  = os.Stdin
	bwrap.Start()
	err = bwrap.Wait()
	//mnt.Process.Signal(syscall.SIGINT)
	if err != nil { return err }

	// Clean up after the app is closed
	// Sleep is needed to wait until the AppImage is unmounted before deleting the temporary dir
	err = UnmountAppImage(ai)
	if err != nil {return err}
	//time.Sleep(50 * time.Millisecond)
	err = os.RemoveAll(ai.TempDir())
	return err
}

func GetWrapArgs(perms *profiles.AppImagePerms) []string {
	var devicePermArgs []string
	var filePermArgs []string
	var socketPermArgs []string
	var sharePermArgs []string
	var stdArgs     []string

	ruid := strconv.Itoa(os.Getuid()) // Real UID, for level 1 RUID and UID are the same value

	// Convert device perms to bwrap format
	for _, v := range(perms.DevicePerms) {
		device := v
		devicePermArgs = append(devicePermArgs, "--dev-bind-try", "/dev/"+device, "/dev/"+device)
	}

	// Convert directory perms to bwrap format
	fp := getXdg(perms.FilePerms, perms.Level)
	for i, _ := range fp {
		dir  := strings.Split(i, ":")[0] // The directory at hand
		auth := strings.Split(i, ":")[1] // Whether it has ro or rw permissions
		genericDir := fp[i]

		// Convert "rw"/"ro" into bwrap command line syntax so we can call it
		if auth == "rw" {
			filePermArgs = append(filePermArgs, "--bind-try", dir, genericDir)
		} else if auth == "ro" {
			filePermArgs = append(filePermArgs, "--ro-bind-try", dir, genericDir)
		}
	}

	// Level 1 is minimal sandboxing, grants access to most system files, all devices and only really attempts to isolate home files
	if perms.Level == 1 {
		stdArgs = []string{
			"--dev-bind",    "/dev", "/dev",
			"--ro-bind",	 "/sys", "/sys",
			"--ro-bind",	 "/usr", "/usr",
			"--ro-bind-try", "/etc", "/etc",
			"--ro-bind-try", xdg.Home+"/.fonts",                     homed+"/.fonts",
			"--ro-bind-try", xdg.ConfigHome+"/fontconfig",           homed+"/.config/fontconfig",
			"--ro-bind-try", xdg.ConfigHome+"/gtk-3.0/gtk.css",      homed+"/.config/gtk-3.0/gtk.css",
			"--ro-bind-try", xdg.ConfigHome+"/gtk-3.0/settings.ini", homed+"/.config/gtk-3.0/settings.ini",
		}
	// Level 2 grants access to fewer system files, and all themes
	} else if perms.Level == 2 {
		stdArgs = []string{
			"--ro-bind-try", "/etc/fonts",              "/etc/fonts",
			"--ro-bind-try", "/usr/share/fontconfig",   "/usr/share/fontconfig",
			"--ro-bind-try", "/usr/share/applications", "/usr/share/applications",
			"--ro-bind-try", "/usr/share/mime",         "/usr/share/mime",
			"--ro-bind-try", "/usr/share/libdrm",       "/usr/share/librdm",
			"--ro-bind-try", "/usr/share/glvnd",        "/usr/share/glvnd",
			"--ro-bind-try", "/usr/share/glib-2.0",     "/usr/share/glib-2.0",
			"--ro-bind-try", xdg.Home+"/.fonts",           homed+"/.fonts",
			"--ro-bind-try", xdg.ConfigHome+"/fontconfig", homed+"/.config/fontconfig",
			"--ro-bind-try", xdg.ConfigHome+"/gtk-3.0",    homed+"/.config/gtk-3.0",
		}
	// Level 3 grants access to only minimal system files (only binaries and libraries and system themes)
	} else if perms.Level == 3 {
		stdArgs = []string{}
	}

	// Basic arguments to be used at any sandboxing level
	stdEnv := []string{
			"--setenv",	  "TMPDIR",              "/tmp",
			"--setenv",	  "HOME",                homed,
			"--setenv",	  "XDG_DESKTOP_DIR",     homed+"/Desktop",
			"--setenv",	  "XDG_DOWNLOAD_DIR",    homed+"/Downloads",
			"--setenv",	  "XDG_DOCUMENTS_DIR",   homed+"/Documents",
			"--setenv",	  "XDG_MUSIC_DIR",       homed+"/Music",
			"--setenv",	  "XDG_PICTURES_DIR",    homed+"/Pictures",
			"--setenv",	  "XDG_VIDEOS_DIR",      homed+"/Videos",
			"--setenv",	  "XDG_TEMPLATES_DIR",   homed+"/Templates",
			"--setenv",	  "XDG_PUBLICSHARE_DIR", homed+"/Templates",
			"--setenv",	  "XDG_DATA_HOME",       homed+"/.local/share",
			"--setenv",	  "XDG_CONFIG_HOME",     homed+"/.config",
			"--setenv",	  "XDG_CACHE_HOME",      homed+"/.cache",
			"--setenv",	  "LOGNAME",             usern,
			"--setenv",	  "USER",                usern,
			"--uid",       uid,
			"--unshare-pid",
			"--unshare-user-try",
			"--die-with-parent",
			"--new-session",
			"--dev",		 "/dev",
			"--proc",		"/proc",
			"--ro-bind",	 "/bin",              "/bin",
			"--ro-bind",	 "/lib",              "/lib",
			"--ro-bind-try", "/lib32",            "/lib32",
			"--ro-bind-try", "/lib64",            "/lib64",
			"--ro-bind",	 "/usr/bin",          "/usr/bin",
			"--ro-bind",	 "/usr/lib",          "/usr/lib",
			"--ro-bind-try", "/usr/lib32",        "/usr/lib32",
			"--ro-bind-try", "/usr/lib64",        "/usr/lib64",
			"--ro-bind-try", "/usr/share/fonts",  "/usr/share/fonts",
			"--ro-bind-try", "/usr/share/icons",  "/usr/share/icons",
			"--ro-bind-try", "/usr/share/themes", "/usr/share/themes",
	}

	// These vars will only be used if x11 socket is granted access
	xAuthority := os.Getenv("XAUTHORITY")
	xDisplay := strings.Replace(os.Getenv("DISPLAY"), ":", "", 1)

	stdArgs = append(stdEnv, stdArgs...)

	// Used if this socket is enabled
	var sockets = map[string][]string {
		// For some reason sometimes it doesn't work when binding X0 to another socket
		// ...but sometimes it does
		"x11": {
			"--ro-bind",	 xAuthority,                      homed+"/.Xauthority",
			"--ro-bind",	 tempDir+"/.X11-unix/X"+xDisplay, "/tmp/.X11-unix/X"+xDisplay,
			"--ro-bind-try", "/usr/share/X11",                "/usr/share/X11",
			"--setenv",      "XAUTHORITY",                    homed+"/.Xauthority",
			"--setenv",      "DISPLAY",                       ":"+xDisplay,
		},
		"pulseaudio": {
			"--ro-bind-try", "/run/user/"+ruid+"/pulse", "/run/user/"+uid+"/pulse",
		},
	}

	for socket, _ := range(sockets) {
		_, present := helpers.Contains(perms.SocketPerms, socket)
		if present {
			socketPermArgs = append(socketPermArgs, sockets[socket]...)
		}
	}

	var unshares = map[string]string {
		"user":    "--unshare-user-try",
		"ipc":     "--unshare-ipc",
		"pid":     "--unshare-pid",
		"net":     "--unshare-net",
		"network": "--unshare-net",
		"uts":     "--unshare-uts",
		"cgroup":  "--unshare-cgroup-try",
	}

	for s, _ := range unshares {
		_, present := helpers.Contains(perms.SharePerms, s)
		if present {
			// Single exception, network share requires `/etc/resolv.conf`
			if s == "net" || s == "network" {
				sharePermArgs = append(sharePermArgs, "--share-net", "--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf")
			}
		} else {
			sharePermArgs = append(sharePermArgs, unshares[s])
		}
	}

	cmdArgs := append(stdArgs, devicePermArgs...)
	cmdArgs  = append(cmdArgs, filePermArgs...)
	cmdArgs  = append(cmdArgs, socketPermArgs...)
	cmdArgs  = append(cmdArgs, sharePermArgs...)

	return cmdArgs
}

// Parses requested file and directories into bwrap flags
func getXdg(s []string, level int) map[string]string {
    var genericDir string
	loadName(level)
    dirPerms := make(map[string]string)

    // Map out the XDG directories
    var xdgDirs = map[string]string {
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
    }

    // Anonymize directories by giving them generic names in case the user has
    // changed the location of their XDG-dirs
    var xdgGeneric = map[string]string {
        "xdg-home":        homed,
        "xdg-desktop":     homed+"/Desktop",
        "xdg-download":    homed+"/Downloads",
        "xdg-documents":   homed+"/Documents",
        "xdg-music":       homed+"/Music",
        "xdg-pictures":    homed+"/Pictures",
        "xdg-videos":      homed+"/Videos",
        "xdg-templates":   homed+"/Templaates",
        "xdg-publicshare": homed+"/Share",
        "xdg-config":      homed+"/.config",
        "xdg-cache":       homed+"/.cache",
        "xdg-data":        homed+"/.local/share",
    }

    //s := helpers.DesktopSlice(dirs)
    for i, _ := range s {
        str := s[i]

        // If neither "ro" or "rw" provided, assume "ro"
        l := len(str) //
        if l <= 3 || str[l-3:] != ":ro" && str[l-3:] != ":rw" {
           str = str+":ro"
        }

        // Replace the xdg-* strings with the corresponding directories on the user's machine
        for key, val := range xdgDirs {

            // If length of key bigger than requested directory or not equal to it continue because there is no reason to look at it further
            if len(key) > len(str) || key != str[:len(key)] {
                continue
            }

            // If the last byte of the requested path shortened to key length is a '/' or ':' we know it's the parent dir, so resolve it using the xdgDirs map
            c := str[len(key)]          // The final byte of the key (used for splitting)
            r := str[len(key):] // Every string after that byte
            if c == byte('/') || c == byte(':') {
                genericDir = xdgGeneric[key] + strings.Split(r, ":")[0]
                s[i] = strings.Replace(str, key, val, 1)
                break
            } else {
                genericDir = strings.Split(str, ":")[0]
            }
        }

        dirPerms[s[i]] = genericDir
    }

    return dirPerms
}
