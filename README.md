# aisap

<p align="center"><img src="resources/aisap.svg" width=256 height="256"/></p>

AppImage SAndboxing Project: a Golang library to help sandbox AppImages with
bwrap

VERY EARLY DEVELOPMENT! Many parts of this are subject to change and should be
expected to until it reaches a more stable form

## What is it?
aisap intends to be a simple way to implement Android/Flatpak style sandboxing
with AppImages. It does have a partial profile system, but it intends to keep
it as basic as possible, thus easier to understand what a program actually
requires to run. This is NOT a sandboxing implementation in and of itself, it
just helps take simple rules and feeds them into bwrap

It currently has a basic re-implementaion of the go-appimage API, so modifying
existing GoLang programs to include sandboxing should be fairly painless

## Bare bones basics
In order for aisap to sandbox an AppImage, it requires a basic profile, which
is a desktop entry (INI format) containing at least one of the following flags
under the `[X-AppImage-Required-Permissions]` section:
```
Files
Devices
Sockets
Share
```
These flags can be included in the AppImage's internal desktop file, another desktop entry by use of the `--profile` command flag with aisap-bin, or aisap's internal profile library, which is simply an arrary of permissions based on known AppImage's names (the `Name` desktop entry flag)

The ultimate goal is to have as many popular AppImages in [aisap's internal library](profiles/README.md) as possible, while smaller, less known ones may request their own permssions per the developer. Running programs sandboxed should mostly be seamless and feel native with the system

Unfortunately, I've been unable to find a Golang binding for squashfuse, so my current *hacky* workaround is to simply use squashfuse's binary executable, which needs to be in the same directory as any project that uses this library, or installed on the host system. Once I get sandboxing to a more stable state, I fully intend to create a proper binding of squashfuse for Golang, which would certainly benefit more than this project (unless someone else wants to knock that part out 😉)

## Usage of the aisap CLI tool
aisap CLI tool is a simple wrapper around the library to mainly serve as a
demonstration. It does *not* offer any desktop integration, but it should be
fairly easy to wrap it into shell scripts to automatically launch AppImages in.
I use it daily on my system and it generally works well, but there is no
guarentee everything will work as intended yet. If something isn't working,
send a bug report! I'll try to fix it ASAP

If a program is supported by aisap's internal library, or if the AppImage has
provided its own requirement for permissions, aisap CLI will run and sandbox
the requested AppImage using the simple syntax:
```
aisap f.appimage
```

It also includes several command line flags:
```
normal options:
  -h, --help:        display this help menu
  -l, --list-perms:  print all permissions to be granted to the app

long-only options:
  --example:     print out examples
  --level:       change the permissions level
  --add-file:    give the sandbox access to a filesystem object
  --add-device:  add a device to the sandbox (eg dri)
  --add-socket:  allow the sandbox to access another socket (eg x11)
  --rm-file:     revoke a file from the sandbox
  --rm-device:   remove access to a device
  --rm-socket:   disable a socket
  --profile:     use a profile from a desktop entry
  --version:     show the version and quit
```

## Sandboxing levels
Sandboxing levels allow for a base configuration of system files to grant by
default. For AppImages that lack an internal profile with aisap, the default
is 3 (which will likely cause the app not to work at all, so it is reccomended
to launch using the command line flags to grant access to required permissions
or to create a profile).

`Level 0` lacks sandboxing entirely, it does the exact same thing as simply launching the AppImage directly

`Level 1` is the most lenient sandbox, it gives access to almost all system and device files but still restricts home files

`Level 2` is intended to be the target for most GUI apps when creating a profile. It gives access to common system files like fonts and themes, restricts device files and home files

`Level 3` is the most strict. It only grants access to very basic system files (binaries and libraries). It should mainly be used for console applications

## API:
### NewAppImage
```
NewAppImage(src string) (*AppImage, error)
```
Re-implementation from go-appimage. The main differences being that this mounts the AppImage instead of extracting its files (because it needs to be mounted in order to sandbox anyway)

### Unmount
```
Unmount(ai *AppImage) error
```
Unmount (shockingly) unmounts the requested AppImage. Note that this is an AppImage struct, which is created by the `NewAppImage()` function. This function needs to be used before using `os.Exit()` or /tmp will be trashed with mounted AppImages

### Run
```
Run(ai *AppImage, args []string) error
```
Run executes the AppImage without any sandboxing. However, it still automatically creates a private home directory for the AppImage

### Sandbox
```
Sandbox(ai *AppImage, args []string) error
```
Sandbox takes an AppImage and sandboxes it using the permissions offered.

### GetWrapArgs
```
GetWrapArgs(ai *AppImage) []string
```
GetWrapArgs takes aisap permissions and translates them into bwrap command line flags. This can be used on its own to see what an AppImage *would* launch with, or to manually launch it

### (AppImage) ExtractFile
```
(ai AppImage) ExtractFile(path string, dest string, resolveSymlinks bool) error
```
Extract a file from the AppImage to `dest`. If `resolveSymlinks` is set to false, the raw symlink will be extracted instead of its target

### (AppImage) ExtractFileReader
```
(ai AppImage) ExtractFileRead(path string) (io.ReadCloser, error)
```
Return a reader of the requested file

### (AppImage) Thumbnail
```
(ai AppImage) Thumbnail() (io.Reader, error)
```
Attempts to extract a thumbnail from the AppImage if available. If provided in a format other than PNG (eg: SVG, XPM) it attempts to convert it to PNG before serving

### (AppImage) Type
```
(ai AppImage) Type() int
```
Return the type of AppImage. Only supports type 2 currently

### (AppImage) TempDir
```
(ai AppImage) TempDir() string
```
Returns the AppImage's temporary directory. By default, it will be `/tmp/.aisapTemp_...`

### (AppImage) MountDir
```
(ai AppImage) MountDir() string
```
Returns the AppImage's mountpoint. With aisap, all AppImage structs will be mounted on creation

### (AppImage) AddFile
```
(ai AppImage) AddFile(str string)
```
Add a single file or directory to the sandbox, appending `:rw` to the filename will give it write access. To prevent writing to the file or directory, either append nothing, or `:ro`

### (AppImage) AddFiles
```
(ai AppImage) AddFiles(s []string)
```
Like `AddFile()` but operates on a list of files or directories

### (AppImage) AddDevice
```
(ai AppImage) AddDevice(s string)
```
Give the sandbox access to a device file (eg: `dri`, `input`). Specifying the full path (`/dev/X`) is not necessary and advised against

### (AppImage) AddDevices
```
(ai AppImage) AddDevices(s []string)
```
Allow the sandbox to access more device files (eg: dri, input)

### (AppImage) AddSockets
```
(ai AppImage) AddSockets(s []string)
```
Share sockets with the sandbox (eg: x11, pulseaudio)

### (AppImage) SetPerms
```
(ai AppImage) SetPerms(entryFile string) error
```
Set permissions for an AppImage using a desktop entry containing permissions flags

### (AppImage) SetRootDir
```
(ai AppImage) SetRootDir(d string)
```
Change the directoy that the sandbox grabs system files from. This is useful if you want to hide your real system files or utilize another Linux distro's libraries for compatibility

### (AppImage) SetDataDir
```
(ai AppImage) SetDataDir(d string)
```
Change the `HOME` directory of the AppImage. By default, this is `[APPIMAGE NAME].home` in the same directory

### (AppImage) SetLevel
```
(ai AppImage) SetLevel(l int)
```
Change the sandbox's base level

