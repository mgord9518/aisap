package permissions

type AppImagePerms struct {
	Level        int    `json:"level"`       // How much access to system files
	Files      []string `json:"filesystem"`  // Grant permission to access files
	Devices    []string `json:"devices"`     // Access device files (eg: dri, input)
	Sockets    []string `json:"sockets"`     // Use sockets (eg: x11, pulseaudio, network)
	NoDataDir    bool   `json:"no_data_dir"` // Whether or not a data dir should be created (only
	// use if the AppImage saves ZERO data eg: 100% online or a game without
	// save files)
}
