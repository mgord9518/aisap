# NAME

**aisap** - Go package for sandboxing AppImage application bundles through
bwrap

# SYNOPSIS

Importing:

```go
import(
    aisap "github.com/mgord9518/aisap"
)
```

Opening AppImage bundles:

```go
// Creates an AppImage struct from a source path
aisap.NewAppImage(src: string) (AppImage, error)
```

AppImage methods:

```go
// Unmounts (if applicable) and cleans up the AppImage object
Destroy() error

// Mounts the AppImage at a given destination.
// If no destination is given, the AppImage will be mounted to a temporary
// directory in '$XDG_RUNTIME_DIR/aisap'
Mount(dest ...string) error by

// If AppImage.Perms.Level > 0, the AppImage will be sandboxed with
// AppImage.Perms, if 0, it will run unsandboxed. args are passed directly to
// the AppImage
Run(args \[]string) error

// Identical to AppImage.Run, except AppImage.Perms.Level equaling 0 is an
// error condition
Sandbox(args \[]string) error

// Generates bwrap arguments for sandboxing based on the AppImage's permissions
WrapArgs() ([]string, error)

// Extracts a file (path) from the AppImage's SquashFS image to dest. If
// resolveSymlinks is true and the path is a symlink, the file it points to
// will be extracted instead
ExtractFile(path string, dest string, resolveSymlinks bool)

// Identical to AppImage.ExtractFile, except returns an io.ReadSeeker interface
// instead of writing to a file. resolveSymlinks is not present in this method
// as it would make no sense
ExtractFileReader(path string) (io.ReadCloser, error)

// Returns an io.Reader to the thumbnail of the application bundle if available
Thumbnail() (io.Reader, error)

// Returns the type of the AppImage, currently only supports type 2 and shImg
// (-2). PRs for supporting type 1 and possibly other (well-defined) unofficial
// AppImage implementations are welcome
Type()

// Returns the TempDir of the AppImage, which only exists if the AppImage is
// mounted, otherwise returns an empty string
TempDir()

// Returns the directory the AppImage is mounted to. If not mounted, it will
// return an empty string
MountDir()

// Change the root directory exposed to the sandbox. This could be useful for
// running an AppImage designed for another Linux distro
SetRootDir(root string)
```

AppImagePerms methods:

```go
// Set the base permission level, which denotes what system-level files the
// application should be able to read
SetLevel(level int) error

// Grant the sandbox access to one or more files. The entire filepaths will
// be exposed to the sandbox
AddFiles(files ...string)

// Grant the sandbox access to one or more device files. They should **not**
// be prepended with '/dev'
AddDevices(devices ...string)

// Grant the sandbox access to one or more sockets
AddSockets(sockets ...Socket) error

// Revoke access to one more files from the sandbox
RemoveFiles(files ...string)

// Revoke access to one or more devices from the sandbox
RemoveSockets(sockets ...Socket)

// Revoke access to one or more devices from the sandbox
RemoveDevices(devices ...string)
```
