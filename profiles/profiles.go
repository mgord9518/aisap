package profiles

import (
	"strings"

	"github.com/mgord9518/aisap/permissions"
)

// List of all profiles supported by aisal out of the box.
// Most of these have only been tested on my (Manjaro and Arch) systems, so they may not work correctly on yours
// If that is the case, please report the issue and any error messages you encounter so that I can try to fix them
var Profiles = map[string]permissions.AppImagePerms{
	// Any apps that require superuser can't be sandboxed in this way
	"balenaetcher": {
		Level: 0,
	},
	// Badlion (and others) might be able to get switched to level 2, so specify devices anyway
	// Proprietary and unofficial AppImages should be high priority to be sandboxed to the fullest extent
	"badlion client": {
		Level: 1,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	"blender": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-templates:rw", "xdg-documents:rw" },
		Sockets: []string{ "x11", "wayland" },
	},
	"deemix-gui": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-music:rw" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	"dolphin emulator": {
		Level: 2,
		Files:   []string{ "xdg-download:ro", "~/Games:ro", "~/Roms:ro" },
		Devices: []string{ "dri", "input" },
		Sockets: []string{ "x11", "wayland", "pulseaudio" },
	},
	"dust3d": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-templates:rw", "xdg-documents:rw" },
		Sockets: []string{ "x11", "wayland" },
	},
	"endless sky": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-templates:rw", "xdg-documents:rw" },
		Sockets: []string{ "x11", "wayland", "pulseaudio" },
	},
	"firefox": {
		Level: 2,
		Files:   []string{ "xdg-download:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	"firefox beta": {
		Level: 2,
		Files:   []string{ "xdg-download:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	"firefox nightly": {
		Level: 2,
		Files:   []string{ "xdg-download:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	"fontforge": {
		Level: 2,
		Files:   []string{ "xdg-documents:rw", "~/.fonts:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	// Also requires `/etc/passwd`
	"freecad conda": {
		Level: 1,
		Files:   []string{ "xdg-documents:rw", "xdg-templates:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	"geometrize": {
		Level: 2,
		Files:   []string{ "xdg-pictures:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	"gnu image manipulation program": {
		Level: 1,
		Files:   []string{ "xdg-pictures:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	"inkscape": {
		Level: 2,
		Files:   []string{ "xdg-documents:rw", "xdg-pictures:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	// Network for plugins and syncing
	"joplin": {
		Level: 2,
		Files:   []string{ "xdg-documents:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "network" },
	},
	"krita": {
		Level: 2,
		Files:   []string{ "xdg-pictures:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	"leocad": {
		Level: 2,
		Files:   []string{ "xdg-documents:rw", "xdg-templates:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	"librewolf": {
		Level: 2,
		Files:   []string{ "xdg-download:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	"linedancer": {
		Level: 3,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	"liteide": {
		Level: 2,
		Files:   []string{ "xdg-documents:rw", "~/go:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	// Fails to find SSL certs, need to further investigate to increase the
	// sandbox
	"microsoft edge": {
		Level: 1,
		Files:   []string{ "xdg-download:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
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
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	"mypaint": {
		Level: 2,
		Files:   []string{ "xdg-pictures:rw", "xdg-templates:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	// Write access given to because save files are stored in the same
	// directory as the rom
	"mgba": {
		Level: 2,
		Files:   []string{ "xdg-download:rw", "~/Games:rw", "~/Roms:rw" },
		Devices: []string{ "dri", "input" },
		Sockets: []string{ "x11", "wayland", "pulseaudio" },
	},
	// Network needed for cloud service, and can run in level 2 if given
	// `/etc/passwd`
	// TODO: Provide a fake `/etc/passwd` when running in level 2 or 3
	"onlyoffice desktop editors": {
		Level: 1,
		Files:   []string{ "xdg-documents:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	"photogimp": {
		Level: 1,
		Files:   []string{ "xdg-pictures:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	"pixsrt": {
		Level: 2,
		Files:   []string{ "xdg-pictures:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	"powder toy": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "network" },
	},
	"ppsspp": {
		Level: 2,
		Files:   []string{ "xdg-download:ro", "~/Games:ro", "~/Roms:ro" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "pulseaudio" },
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
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	// Audio for notification sounds
	"sengi": {
		Level: 1,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	// Link to device not tested
	"signal": {
		Level: 2,
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland", "network" },
	},
	"smallbasic": {
		Level: 2,
		Files:   []string{ "xdg-documents:rw" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	// I think it's an interesting idea to have a confined terminal
	// enviornment although it could also greatly hinder its usefullness
	// so I'd like to hear feedback
	// TODO: add more files but keep it isolated from the host system
	"station": {
		Level: 1,
		Devices: []string{ "dri" },
		Files:   []string{ "~/.config/nvim:ro", "~/.profile:ro",
						   "~/.bashrc:ro",      "~/.zshrc:ro",
						   "~/.viminfo:ro"},
		Sockets: []string{ "x11", "wayland" },
	},
	// Untested with real equipment but launches
	// TODO: Properly test Subsurface
	"subsurface": {
		Level: 1,
		Files:   []string{ "xdg-documents:ro" },
		Devices: []string{ "dri" },
		Sockets: []string{ "x11", "wayland" },
	},
	"supertuxkart": {
		Level: 2,
		Devices: []string{ "dri", "input" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	"supertux 2": {
		Level: 2,
		Devices: []string{ "dri", "input" },
		Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
	},
	"texstudio": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw", "xdg-templates:rw" },
		Sockets: []string{ "x11", "wayland", },
	},
	"visual studio code": {
		Level: 2,
		Devices: []string{ "dri" },
		Files:   []string{ "xdg-documents:rw" },
		Sockets: []string{ "x11", "wayland", "network" },
	},
}

func FromName(name string) *permissions.AppImagePerms {
	name = strings.ToLower(name)

	if p, present := Profiles[name]; present {
		return &p
	} else {
		p.Level = -1
		return &p
	}
}
