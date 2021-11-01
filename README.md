# aisap
AppImage SAndboxing Project: a Golang library to help sandbox AppImages with bwrap

## What is it?

aisap intends to be a simple way to implement Android/Flatpak style sandboxing with AppImages. It does have a partial profile system, but it intends to keep it as basic as possible, thus easier to understand what a program actually requires to run. This is NOT a sandboxing implementation in and of itself, it just helps take simple rules and feeds them into bwrap

I also plan on implementing all of the go-appimage library, so that any existing projects could add sandboxing with minimal hassle if they so desire. This is far from complete as of now

aisap itself is a Golang library, but it also provides an executable as a basic implementation

## Bare bones basics

In order for aisap to sandbox an AppImage, it requires a basic profile, which is a desktop entry (INI format) containing at least one of the following flags:
```
X-AppImage-File-Permissios
X-AppImage-Device-Permissions
X-AppImage-Socket-Permissions
X-AppImage-Share-Permissions
```
These flags can be included in the AppImage's internal desktop file, another desktop entry by use of the `--perm-file` command flag with aisap-bin, or aisap's internal profile library, which is simply an arrary of permissions based on known AppImage's names (the `Name` desktop entry flag)

The ultimate goal is to have as many popular AppImages in aisap's internal library as possible, while smaller, less known ones may request their own permssions per the developer. Running programs sandboxed should mostly be seamless and feel native with the system

Unfortunately, I've been unable to find a Golang binding for squashfuse, so my current *hacky* workaround is to simply use squashfuse's binary executable, which needs to be in the same directory as any project that uses this library or on the host system. Once I get sandboxing to a more stable state, I fully intend to create a proper binding of squashfuse for Golang, which would certainly benefit more than this project (unless someone else wants to knock that part out 😉)

### Usage of aisap-bin

aisap-bin is a simple command line utility that serves as a proof-of-concept. I created it to demonstrate some basic usage of the API, but I also intend it to eventually be a decent program on its own. It works, and I use it on my system daily, but *DO NOT* use it for anything serious. Its very likely that there are some bugs in the way permissions are interpreted and sent to brwap.

If a program is supported by aisap's internal library, or if the AppImage has provided its own requirement for permissions, aisap-bin will run and sandbox the requested AppImage using the simple syntax:
```
aisap-bin f.appimage
```

It also includes several command line flags:
```
  -h, --help    Shows usage
  -v, --verbose Be verbose (NEI)

  --add-file    Allow sandbox to access additional files
  --add-dev     Allow sandbox to access additional device files
  --add-soc     Allow sandbox to access additional sockets
  --add-share   Allow sandbox to access additional shares

  --level       Change base level of sandbox (min: 0, max: 3)

  --perm-file   Manually specify an INI format permissions profile for the AppImage
```

## Basic API for aisap
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
