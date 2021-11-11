package profiles

type AppImagePerms struct {
	Level     int    // How much access to system files
	Files   []string // Grant permission to access files
	Devices []string // Access device files (eg: `/dev/dri`)
	Sockets []string // Use sockets (eg: x11)
	Share   []string // Share from host (eg: network)
}
