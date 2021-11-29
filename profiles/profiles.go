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
			Sockets: []string{ "x11", "wayland", "pulseaudio" },
			Share:   []string{ "network" },
		},
		"deemix-gui": {
			Level: 2,
			Files:   []string{ "xdg-music:rw" },
			Sockets: []string{ "x11", "wayland", "pulseaudio" },
			Share:   []string{ "network" },
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
			Sockets: []string{ "x11", "wayland", "pulseaudio" },
			Share:   []string{ "network" },
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
			Sockets: []string{ "x11", "wayland", "pulseaudio" },
			Share:   []string{ "network" },
		},
		// Fails to find SSL certs, need to further investigate to increase the
		// sandbox
		"microsoft edge": {
			Level: 1,
			Files:   []string{ "xdg-download:rw" },
			Devices: []string{ "dri" },
			Sockets: []string{ "x11", "wayland", "pulseaudio" },
			Share:   []string{ "network" },
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
		// Link to device not tested
		"signal": {
			Level: 2,
			Sockets: []string{ "x11", "wayland" },
			Share:   []string{ "network" },
		},
		"supertuxkart": {
			Level: 2,
			Devices: []string{ "dri", "input" },
			Sockets: []string{ "x11", "wayland", "pulseaudio" },
			Share:   []string{ "network" },
		},
		"supertux 2": {
			Level: 2,
			Devices: []string{ "dri" },
			Sockets: []string{ "x11", "wayland", "pulseaudio" },
			Share:   []string{ "network" },
		},
		"powder toy": {
			Level: 2,
			Devices: []string{ "dri" },
			Sockets: []string{ "x11", "wayland" },
			Share:   []string{ "network" },
		},
		"visual studio code": {
			Level: 2,
			Files:   []string{ "xdg-documents:rw" },
			Sockets: []string{ "x11", "wayland" },
			Share:   []string{ "network" },
		},
	}
