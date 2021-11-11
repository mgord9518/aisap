# aisap
AppImage SAndboxing Project: a Golang library to help sandbox AppImages with bwrap

## What is it?

aisap intends to be a simple way to implement Android/Flatpak style sandboxing with AppImages. It does have a partial profile system, but it intends to keep it as basic as possible, thus easier to understand what a program actually requires to run. This is NOT a sandboxing implementation in and of itself, it just helps take simple rules and feeds them into bwrap

I also plan on implementing all of the go-appimage library, so that any existing projects could add sandboxing with minimal hassle if they so desire. This is far from complete as of now

aisap itself is a Golang library, but it also provides an executable as a basic implementation

## Bare bones basics

In order for aisap to sandbox an AppImage, it requires a basic profile, which is a desktop entry (INI format) containing at least one of the following flags:
```
X-AppImage-Sandbox-Files
X-AppImage-Sandbox-Devices
X-AppImage-Sandbox-Sockets
X-AppImage-Sandbox-Share
```
These flags can be included in the AppImage's internal desktop file, another desktop entry by use of the `--profile` command flag with aisap-bin, or aisap's internal profile library, which is simply an arrary of permissions based on known AppImage's names (the `Name` desktop entry flag)

The ultimate goal is to have as many popular AppImages in aisap's internal library as possible, while smaller, less known ones may request their own permssions per the developer. Running programs sandboxed should mostly be seamless and feel native with the system

Unfortunately, I've been unable to find a Golang binding for squashfuse, so my current *hacky* workaround is to simply use squashfuse's binary executable, which needs to be in the same directory as any project that uses this library or on the host system. Once I get sandboxing to a more stable state, I fully intend to create a proper binding of squashfuse for Golang, which would certainly benefit more than this project (unless someone else wants to knock that part out 😉)

## Usage of aisap-bin

aisap-bin is a simple command line utility that serves as a proof-of-concept. I created it to demonstrate some basic usage of the API, but I also intend it to eventually be a decent program on its own. It works, and I use it on my system daily, but *DO NOT* use it for anything serious. Its very likely that there are some bugs in the way permissions are interpreted and sent to brwap.

If a program is supported by aisap's internal library, or if the AppImage has provided its own requirement for permissions, aisap-bin will run and sandbox the requested AppImage using the simple syntax:
```
aisap-bin f.appimage
```

It also includes several command line flags:
```
  -h, --help       Shows usage
  -v, --verbose    Be verbose (NEI)
  -l, --list-perms List permissions to be granted by the AppImage's profile

  --file     Add file to sandbox
  --device   Allow sandbox to access additional device files
  --socket   Allow sandbox to access additional sockets
  --share    Add share to sandbox (eg: network)

  --level    Change base level of sandbox (min: 0, max: 3)

  --profile  Manually specify an INI format permissions profile for the AppImage
```

## Sandboxing levels
Sandboxing levels allow for a base configuration of system files to grant by default. For AppImages that lack an internal profile with aisap, the default is 1.

`Level 0` lacks sandboxing entirely, it does the exact same thing as simply launching the AppImage directly

`Level 1` is the most lenient sandbox, it gives access to many system files but still restricts home files

`Level 2` is intended to be the target for most GUI apps when creating a profile. It gives access to common system files, restricts device files and home files

`Level 3` is the most strict. It only grants access to very basic system files (binaries, libraries and themes). It should mainly be used for console applications

## Basic API for aisap
NOTICE: API is under heavy development and many aspects are likely to change! This should currently be used for testing ONLY

### MountAppImage
```
MountAppImage(src string, dest string) error
```
MountAppImage mounts the requested AppImage file path (src) to the destination directory (dest). Returns error if not successful

### UnmountAppImage
```
UnmountAppImage(ai *AppImage) error
```
UnmountAppImage (shockingly) unmounts the requested AppImage. Note that this is an AppImage struct, which is created by the `NewAppImage()` function

### UnmountAppImageFile
```
UnmountAppImageFile(mntPt str) error
```
UnmountAppImageFile does the exact same thing as `UnmountAppImage()`, but works on a specified directory instead of an AppImage struct

### GetAppImageOffset
```
GetAppImageOffset(src string) (int, error)
```
GetAppImageOffset finds and returns the offset of requested AppImage source file in an int. Returns error if not successful

### GetElfSize
```
GetElfSize(src string) (int, error)
```
GetElfSize calculates the byte size of an ELF executable, returning its size. Returns error if not successful

### GetAppImageType
```
GetAppImageType(src string) (string, error)
```
GetAppImageType finds what type of AppImage a file is (if any), returning either `1` for ISO disk image AppImage, or `2` for type 2 SquashFS AppImage. Returns error if unsuccessful

### Run
```
Run(ai *AppImage, args []string) error
```
Run executes the AppImage without any sandboxing. However, it still automatically creates a private home directory for the AppImage.

### Wrap
```
Wrap(ai *AppImage, perms *profiles.AppImagePerms, args []string) error
```
Wrap takes an AppImage and sandboxes it using the permissions offered. Returns error if not successful

### GetWrapArgs
```
GetWrapArgs(perms *profiles.AppImagePerms) []string
```
GetWrapArgs takes aisap permissions and translates them into bwrap command line flags
### (AppImage) Thumbnail
```
(ai AppImage) Thumbnail() (io.Reader, error)
```
AppImage.Thumbnail() attempts to extract a thumbnail from the AppImage if available. If provided in a format other than PNG (eg: SVG, XPM) it attempts to convert it to PNG before serving
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
### (AppImage) RunId
```
(ai AppImage) RunId() string
```
Returns the AppImage's run ID. This is a random string of characters used in the mount point and sandboxed temporary directory
### (AppImage) AddFiles
```
(ai AppImage) AddFiles(s []string)
```
Give the sandbox access to specified files and directories. Every file must either be provided as an XDG standard name (eg: `xdg-download`, `xdg-home`) or a full path
### (AppImage) AddDevices
```
(ai AppImage) AddDevices(s []string)
```
Allow the sandbox to access more device files
### (AppImage) AddSockets
```
(ai AppImage) AddSockets(s []string)
```
Share sockets with the sandbox (eg: x11, pulseaudio)
### (AppImage) AddShare
```
(ai AppImage) AddShare(s []string)
```
Share other parts of your system with the sandbox (eg: network)
### (AppImage) SetPerms
```
(ai AppImage) SetPerms(entryFile string) error
```
### (AppImage) SetRootDir
```
(ai AppImage) SetRootDir(d string)
```
Change the directoy that the sandbox grabs system files from. This is useful if you want to hide your real system files or utilize another Linux distro's libraries for compatibility
### (AppImage) SetDataDir
```
(ai AppImage) SetDataDir(d string)
```
Change the `HOME` directory of the AppImage. By default, this is `[APPIMAGE NAME].home`
### (AppImage) SetTempDir
```
(ai AppImage) SetTempDir(d string)
```
Change the temporary directory of the AppImage sandbox. This is the sandbox's `/tmp`
### (AppImage) Type
```
(ai AppImage) Type() int
```
Return the type of AppImage
### (AppImage) ExtractFile
```
(ai AppImage) ExtractFile(path string, dest string, resolveSymlinks bool) error
```
Extract a file from the AppImage to `dest`. If `resolveSymlinks` is set to false, the raw symlink will be extracted instead of its target
