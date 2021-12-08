package permissions

type AppImagePerms struct {
	Level     int    // How much access to system files
	Files   []string // Grant permission to access files
	Devices []string // Access device files (eg: dri, input)
	Sockets []string // Use sockets (eg: x11, pulseaudio, network)
}
