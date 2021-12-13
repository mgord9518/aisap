package profiles

import (
	"github.com/mgord9518/aisap/permissions"
)

// List of all profiles supported by aisal out of the box.
// Most of these have only been tested on my (Manjaro and Arch) systems, so they may not work correctly on yours
// If that is the case, please report the issue and any error messages you encounter so that I can try to fix them
var Profiles = map[string]permissions.AppImagePerms{
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
		// Dolphin could be level 2 but sound breaks, need to find a solution to that
		"dolphin emulator": {
			Level: 1,
			Files:   []string{ "xdg-download:ro", "~/Games:ro", "~/Roms:ro" },
			Devices: []string{ "dri", "input" },
			Sockets: []string{ "x11", "wayland", "pulseaudio" },
		},
		"firefox": {
			Level: 2,
			Files:   []string{ "xdg-download:rw" },
			Devices: []string{ "dri" },
			Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
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
		// Network for 
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
		"librewolf": {
			Level: 2,
			Files:   []string{ "xdg-download:rw" },
			Devices: []string{ "dri" },
			Sockets: []string{ "x11", "wayland", "pulseaudio", "network" },
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
		"photogimp": {
			Level: 1,
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
		// Only partially tested (I don't have an RS acct) but title screen
		// works as intended
		"runelite": {
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
		// I think it's an interesting idea to have a confined terminal
		// enviornment although it could also greatly hinder its usefullness
		// so I'd like to hear feedback
		// TODO: add more files but keep it isolated from the host system
		"station": {
			Level: 1,
			Devices: []string{ "dri" },
			Files:   []string{ "~/.config/nvim:ro", "~/.profile:ro",
			                   "~/.bashrc:ro",      "~/.zshrc:ro" },
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
