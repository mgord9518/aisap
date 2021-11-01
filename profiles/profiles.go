package github.com/mgord9518/aisap/profiles

// List of all profiles supported by aisal out of the box.
var Profiles = map[string]AppImagePerms{
		// TODO: add some base permissions that profiles can be built off of,
		// because many share the same requirements (and a way to inherit them)
//		"___BASIC_GUI___": {
//			Level: 2,
//			DevicePerms: []string{ "dri" },
//			SocketPerms: []string{ "x11", "wayland" },
//		}
		"LibreWolf": {
			Level: 2,
			FilePerms:   []string{ "xdg-download:rw" },
			DevicePerms: []string{ "dri" },
			SocketPerms: []string{ "x11", "wayland" },
			SharePerms:  []string{ "network" },
		},
		"The Powder Toy": {
			Level: 3,
			DevicePerms: []string{ "dri" },
			SocketPerms: []string{ "x11", "wayland" },
			SharePerms:  []string{ "network" },
		},
		// Minecraft currently must be set to level 1 because it requires a
		// valid `/etc/passwd` to launch... for some weird reason...
		// TODO: Create fake passwd file so that sandbox levels 2 and 3 can
		// launch apps that require it to be valid
		"Minecraft": {
			Level: 1,
			DevicePerms: []string{ "dri" },
			SocketPerms: []string{ "x11", "wayland" },
			SharePerms:  []string{ "network" },
		},
	}
