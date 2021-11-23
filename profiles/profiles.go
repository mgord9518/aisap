package profiles

// List of all profiles supported by aisal out of the box.
var Profiles = map[string]AppImagePerms{
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
			Devices: []string{ "dri" },
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
