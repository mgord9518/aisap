package profiles

// List of all profiles supported by aisal out of the box.
// Most of these have only been tested on my (Manjaro and Arch) systems, so they may not work correctly on yours
// If that is the case, please report the issue and any error messages you encounter so that I can try to fix them
var Profiles = map[string]AppImagePerms{
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
		"dolphin emulator": 
			Level: 1,
			Files:   []string{ "xdg-download:ro", "~/Games:ro", "~/Roms:ro" },
			Devices: []string{ "dri", "input" },
			Sockets: []string{ "x11", "wayland", "pulseaudio" },
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
		// Minecraft requires access to keyring in order to launch correctly,
		// until a fix is found Minecraft will have to be run without a sandbox
		"minecraft": {
			Level: 0,
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
		"the powder toy": {
			Level: 2,
			Devices: []string{ "dri" },
			Sockets: []string{ "x11", "wayland" },
			Share:   []string{ "network" },
		},
	}
