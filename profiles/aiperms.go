package github.com/mgord9518/aisap/profiles

type AppImagePerms struct {
	Level	      int
	FilePerms   []string
	DevicePerms []string
	SocketPerms []string
	SharePerms  []string
}
