# NAME
**aisap** - CLI tool to sandbox AppImage application bundles

# SYNOPSIS
**aisap** \[appimage] \[[option](#OPTIONS)]...

# DESCRIPTION
aisap is a tool to sandbox AppImage application bundles along with some small
secondary uses.

If used with no [options](#OPTIONS), aisap will attempt to sandbox the given
application bundle based upon permissions described in the following order:

1. Permissions set by the user via CLI flags
2. Permissions set by the user via an entry in *~/.local/share/aisap/profiles*
3. Permissions requested by the application bundle
4. Permissions via aisap's internal profile library

If no permissions can be located, aisap will refuse to run the application.

# OPTIONS
**-h**, **--help**
: Print usage information

**-l**, **--list-perms**
: Print all permissions currently granted to the application bundle

**-v**, **--verbose**
: Make output more verbose

**--example**
: Print usage examples

**--level** \[int: 0..3]
: Modify the base system permissions level

**--trust** \[bool]
: Set whether application bundle is trusted (untrusted bundles cannot be run)

**--trust-once**
: Allow running an untrusted application bundle a single time

**--root-dir** \[dir: str]
: Override the root directory exposed to the sandbox

**--data-dir** \[dir: str]
: Change the application's sandbox home location (defaults to \[AppImage].home)

**--no-data-dir**
: Make the sandbox home a tmpfs and not preserve changes

**--add-file** \[path: str]
: Give the sandbox access to a file or directory

**--add-device** \[device: str]
: Give the sandbox access to a device file (eg: input, dri)

**--add-socket** \[socket: str]
: Give the sandbox access to a socket (eg: network)

**--rm-file** \[path: str]
: Revoke access to a file from the sandbox

**--rm-device** \[device: str]
: Revoke access to a device from the sandbox

**--rm-socket** \[socket: str]
: Revoke access to a socket from the sandbox

**--extract-icon** \[dest: str]
: Extract an application's icon (may be in PNG or SVG format)

**--extract-thumbnail** \[dest: str]
: Extract an application's thumbnail preview (PNG format; should be 256x256 or
512x512)

**--profile** \[path: str]
: Set permissions for a single run based on a permissions file

**--fallback-profile** \[path: str]
: If no [profile](#DESCRIPTION) is found, this permissions file will be used as
a fallback 

**--version**
: Print aisap version (semantic)

# ENVIRONMENT
**NO_COLOR**
: Causes aisap to stop using ANSI color codes. This should be used if your
terminal doesn't have color support

# FILES
**~/.local/share/aisap/profiles**
: In this directory, individual permissions entries may be placed to override
AppImage [permissions](#DESCRIPTION). 

The permissions files are in case-sensitive INI format (such as a desktop
entry), laid out as follows:
```ini
# Section header
[X-App Permissions]

# Base permissions level
# This must be an integer between 0 and 3
Level=2

# Files/ directories that the sandbox may access
# Individual files may be separated with a semicolon (;) and currently does
# not support quoting, but I have plans to add this. XDG basedirs (such as
# xdg-download) are supported, along with using the tilde to represent your
# HOME directory. `:rw` may be appended to the filename to give read access and
# `:ro` may be appended to denote as read-only, but read-only is automatically
# implied
Files=~/add_me;xdg-download

# Device files
# Similarly to the `Files` key, individual values may be separated using a
# semicolon. These should not be prepended with `/dev`
Devices=dri;input

# Sockets
# These also may be separated with semicolons
Sockets=x11;network
```
