package aisap

// SetRootDir is used to set an alternative root directory for the Wrap function to share system files from (default: /)
func SetRootDir(dir string) {
    rootDir = dir
}

// SetDataDir is used to set an alternative data directory for the Wrap function to store the AppImage's "home" files (default: [APPIMAGE PATH].home)
func SetDataDir(dir string) {
    dataDir = dir
}

func SetTempDir(dir string) {
    tempDir = dir
}
