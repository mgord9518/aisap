package profiles

import (
	"errors"
	"strings"

	helpers     "github.com/mgord9518/aisap/helpers"
	permissions "github.com/mgord9518/aisap/permissions"
)

// List of all profiles supported by aisap out of the box.
// Most of these have only been tested on my (Arch and Nix) systems, so
// they may not work correctly on yours. If that is the case, please report the
// issue and any error messages you encounter so that I can try to fix them
// NOTE: Some app permissions are `aliases` of others, so care must be taken
// that modifying the parent permission will also affect apps based on it
// 90 unique apps currently supported
var profiles = map[string]permissions.AppImagePerms{
	"0 a.d.": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "alsa", "network" },
	},
	"aaaaxy": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "alsa" },
	},
	"aranym jit": {
		Level: 3,
		Devices: []string{ "dri", "input" },
		Files:   []string{ "xdg-download:ro", "~/Games:ro", "~/Roms:ro" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	"armagetron advanced": {
		Level: 3,
		Devices: []string{ "dri", "input" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	"appimage pool": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "~/Applications:rw" },
		Sockets: []string{ "wayland", "x11", "network" },
	},
	"appimageupdate": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "~/Applications:rw" },
		Sockets: []string{ "x11", "network" },
	},
	// Untested with Android device, left level 1 assuming it needs access to all
	// of `/dev`
	"apk editor studio": {
		Level: 1,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-templates:rw", "xdg-download:rw" },
		Sockets: []string{ "x11" },
	},
	"badlion client": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	"balenaetcher": {
		Level: 0,
	},
	"blender": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-templates:rw", "xdg-documents:rw" },
		Sockets: []string{ "x11" },
	},
	"brave": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-download:rw" },
		Sockets: []string{ "x11", "pulseaudio", "network" },
	},
	// TODO: Find files responsible for reporting MESA info to increase sandbox
	"bugdom": {
		Level: 1,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	"calibre": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:ro" },
		Sockets: []string{ "x11" },
	},
	"chromium": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-download:rw" },
		Sockets: []string{ "x11", "pulseaudio", "network" },
	},
	// I think it's an interesting idea to have a confined terminal
	// enviornment although it could also greatly hinder its usefullness
	// so I'd like to hear feedback
	// TODO: add more files but keep it isolated from the host system
	// Untested with real equipment but launches
	"cool retro term": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "~/.config/nvim:ro", "~/.profile:ro",
		                   "~/.bashrc:ro",      "~/.zshrc:ro",
		                   "~/.viminfo:ro"},
		Sockets: []string{ "x11", "network" },
	},
	"conky": {
		Level: 1,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "pid" },
	},
	"deemix-gui": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-music:rw" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	"densify": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw" },
		Sockets: []string{ "x11" },
	},
	"desmume": {
		Level: 2,
		Devices: []string{ "dri", "input" },
		Files:   []string{ "xdg-download:rw", "~/Games:rw", "~/Roms:rw" },
		Sockets: []string{ "x11", "alsa" },
	},
	// Network for netplay
	"dolphin emulator": {
		Level: 2,
		Devices: []string{ "dri", "input" },
		Files:   []string{ "xdg-download:ro", "~/Games:ro", "~/Roms:ro" },
		Sockets: []string{ "x11", "alsa", "network" },
	},
	"dust3d": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-templates:rw", "xdg-documents:rw" },
		Sockets: []string{ "x11" },
	},
	"eagle mode": {
		Level: 1,
		Devices: []string{ "dri" },
		// Not really sure if the better way to go about it is just supplying
		// it with access to the home directory or giving XDG directories like
		// so
		Files:   []string{ "xdg-documents:rw", "xdg-publicshare:rw",
		                   "xdg-templates:rw",  "xdg-desktop:rw",
						   "xdg-documents:rw",  "xdg-download:rw",
					       "xdg-music:rw",      "xdg-videos:rw"},
		Sockets: []string{ "x11", "audio" },
	},
	"edex-ui": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "~/.config/nvim:ro", "~/.profile:ro",
		                   "~/.bashrc:ro",      "~/.zshrc:ro",
		                   "~/.viminfo:ro"},
		Sockets: []string{ "x11", "network" },
	},
	"element": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "network" },
	},
	"endless sky": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-templates:rw", "xdg-documents:rw" },
		Sockets: []string{ "x11", "alsa" },
	},
	// Currently doesn't work at higher sandboxing levels "requires graphical
	// enviornment to run" error
	"eternal lands (appimage)": {
		Level: 1,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	"firefox": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-download:rw" },
		Sockets: []string{ "x11", "pulseaudio", "network", "dbus" },
	},
	"fontforge": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw", "~/.fonts:rw" },
		Sockets: []string{ "x11" },
	},
	"fractale": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11" },
	},
	// Also requires `/etc/passwd`
	"freecad conda": {
		Level: 1,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw", "xdg-templates:rw" },
		Sockets: []string{ "x11" },
	},
	"gambatte_qt": {
		Level: 2,
		Devices: []string{ "dri", "input" },
		Files:   []string{ "xdg-download:rw", "~/Games:rw", "~/Roms:rw" },
		Sockets: []string{ "x11", "alsa" },
	},
	"geometrize": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-pictures:rw" },
		Sockets: []string{ "x11" },
	},
	"go": {
		Level: 3,
		Files:   []string{ "xdg-documents:rw" },
	},
	"google chrome": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-download:rw" },
		Sockets: []string{ "x11", "pulseaudio", "network" },
	},
	"gnu image manipulation program": {
		Level: 1,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-pictures:rw" },
		Sockets: []string{ "x11" },
	},
	"hearts": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "network", "alsa" },
	},
	"hyper": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "~/.config/nvim:ro", "~/.profile:ro",
		                   "~/.bashrc:ro",      "~/.zshrc:ro",
		                   "~/.viminfo:ro"},
		Sockets: []string{ "x11", "network" },
	},
	"imagemagick": {
		Level: 3,
		Files:   []string{ "xdg-documents:rw", "xdg-pictures:rw" },
		Devices: []string{ "dri" },
	},
	"inkscape": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw", "xdg-pictures:rw" },
		Sockets: []string{ "x11" },
	},
	"iqpuzzle": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11" },
	},
	// Network for plugins and syncing
	"joplin": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw" },
		Sockets: []string{ "x11", "network" },
	},
	"krita": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-pictures:rw" },
		Sockets: []string{ "x11" },
	},
	"leocad": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw", "xdg-templates:rw" },
		Sockets: []string{ "x11" },
	},
	"librewolf": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-download:rw" },
		Sockets: []string{ "x11", "pulseaudio", "network", "dbus" },
	},
	"linedancer": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11" },
		NoDataDir: true,
	},
	"liteide": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw", "~/go:rw" },
		Sockets: []string{ "x11" },
	},
	"microsoft edge": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-download:rw" },
		Sockets: []string{ "x11", "pulseaudio", "network" },
	},
	// Minecraft requires access to keyring in order to launch correctly,
	// until a fix is found Minecraft will have to be run without a sandbox
	"minecraft": {
		Level: 0,
	},
	// Minetest demonstrates that not all GUI apps need level 2 or lower
	// fully self-contained apps that don't use system fonts, etc. can be
	// run in level 3
	"minetest": {
		Level: 3,
		Devices: []string{ "dri", "input" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	"mypaint": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-pictures:rw", "xdg-templates:rw" },
		Sockets: []string{ "x11" },
	},
	// Write access given to because save files are stored in the same
	// directory as the rom
	"mgba": {
		Level: 2,
		Devices: []string{ "dri", "input" },
		Files:   []string{ "xdg-download:rw", "~/Games:rw", "~/Roms:rw" },
		Sockets: []string{ "x11", "alsa" },
	},
	"naev": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "wayland", "x11", "pulseaudio" },
	},
	"newton adventure": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "alsa" },
	},
	"nmeasimulator": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "network" },
	},
	"notepad next": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-download:rw", "xdg-documents:rw", "xdg-templates:rw" },
		Sockets: []string{ "x11", "network" },
	},
	"nx-software-center": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "~/Applications:rw" },
		Sockets: []string{ "x11", "network" },
	},
	// Network needed for cloud service, and can run in level 2 if given
	// `/etc/passwd`
	// TODO: Provide a fake `/etc/passwd` when running in level 2 or 3
	"onlyoffice desktop editors": {
		Level: 1,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	"openblok": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "alsa" },
	},
	"passy": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11" },
	},
	"photogimp": {
		Level: 1,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-pictures:rw" },
		Sockets: []string{ "x11" },
	},
	"pix": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-pictures:rw", "xdg-download:rw" },
		Sockets: []string{ "wayland", "x11" },
	},
	"pixsrt": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-pictures:rw" },
		Sockets: []string{ "x11" },
	},
	"play 2048": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11" },
	},
	"potato presenter": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw" },
		Sockets: []string{ "x11" },
	},
	"powder toy": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "network" },
	},
	"ppsspp": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-download:ro", "~/Games:ro", "~/Roms:ro" },
		Sockets: []string{ "x11", "audio" },
	},
	// Python given no extra permissions, but can easily be customized for
	// scripts that require more
	"python3.10.1": {
		Level: 3,
	},
	// Only partially tested (I don't have an RS acct) but title screen
	// works as intended
	"runelite": {
		Level: 1,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	// Audio for notification sounds
	"sengi": {
		Level: 1,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	// Link to device not tested
	"signal": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "network" },
	},
	"smallbasic": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw" },
		Sockets: []string{ "x11" },
	},
	"space cadet pinball": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "alsa", "dbus" },
	},
	"stackandconquer": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11" },
	},
	"stacer": {
		Level: 0,
	},
	"stallboard": {
		Level: 1,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "audio" },
	},
	"station": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "~/.config/nvim:ro", "~/.profile:ro",
		                   "~/.bashrc:ro",      "~/.zshrc:ro",
		                   "~/.viminfo:ro"},
		Sockets: []string{ "wayland", "x11", "network" },
	},
	"stellarium": {
		Level: 1,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11" },
	},
	// TODO: Properly test Subsurface
	"subsurface": {
		Level: 1,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:ro" },
		Sockets: []string{ "x11" },
	},
	"stunt car remake": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "alsa" },
	},
	"supertux 2": {
		Level: 3,
		Devices: []string{ "dri", "input" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	"supertuxkart": {
		Level: 3,
		Devices: []string{ "dri", "input" },
		Sockets: []string{ "x11", "audio", "network" },
	},
	"synthein": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "alsa" },
	},
	// Must be level 1 to get necessary files from /usr
	"texstudio": {
		Level: 1,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw", "xdg-templates:rw" },
		Sockets: []string{ "x11" },
	},
	"thorium browser": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-download:rw" },
		Sockets: []string{ "x11", "pulseaudio", "network" },
	},
	"tiled": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw", "xdg-pictures:rw", "xdg-templates:rw" },
		Sockets: []string{ "x11" },
	},
	"visual studio code": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw" },
		Sockets: []string{ "x11", "network" },
	},
	"waterfox": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-download:rw" },
		Sockets: []string{ "x11", "pulseaudio", "network", "dbus" },
	},
	"xonotic": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "alsa", "network" },
	},
	"yuzu": {
		Level: 2,
		Devices: []string{ "dri", "input" },
		Files:   []string{ "xdg-download:ro", "~/Games:ro", "~/Roms:ro" },
		Sockets: []string{ "x11", "alsa", "network" },
	},
}

func FromName(name string) (*permissions.AppImagePerms, error) {
	name = strings.ToLower(name)

	if p, present := profiles[name]; present {
		p.Files = helpers.CleanFiles(p.Files)
		return &p, nil
	}

	// Load in duplicate permissions based on their names as some of the same
	// AppImages may be released under different names
	aliases := map[string]string {
		"aranym mmu":          "aranym jit",
		"firefox beta":        "firefox",
		"firefox nightly":     "firefox",
		"python2.7.18":        "python",
		"python3":             "python",
		"python3.5.10":        "python",
		"python3.6.15":        "python",
		"python3.7.12":        "python",
		"python3.8.12":        "python",
		"python3.9.9":         "python",
		"python3.10.1":        "python",
		"waterfox classic":    "waterfox",
	}

	if a, present := aliases[name]; present {
		p := profiles[a]
		p.Files = helpers.CleanFiles(p.Files)
		return &p, nil
	}

	// If both tests fail, return with a level of -1
	return &permissions.AppImagePerms{ Level: -1 }, errors.New("cannot find permissions for app `" + name + "`")
}

func Profiles() map[string]permissions.AppImagePerms {
	return profiles
}
