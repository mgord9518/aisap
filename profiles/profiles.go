package profiles

// List of all profiles supported by aisal out of the box.
var Profiles = map[string]AppImagePerms{
		"LibreWolf": {
			Level: 2,
			Files:   []string{ "xdg-download:rw" },
			Devices: []string{ "dri" },
			Sockets: []string{ "x11", "wayland", "pulseaudio" },
			Share:   []string{ "network" },
		},
		"The Powder Toy": {
			Level: 2,
			Devices: []string{ "dri" },
			Sockets: []string{ "x11", "wayland" },
			Share:   []string{ "network" },
		},
		// Minecraft requires access to keyring in order to launch correctly,
		// until a fix is found Minecraft will have to be run without a sandbox
		"Minecraft": {
			Level: 0,
		},
	}
