# aisap permissions
Simplified permissions and what actual bwrap flags they correspond to

## Base levels
As you can see, level 1 gives access to a wide range of system files, but personal [HOME] files are still restricted. Level 1 is intended to allow some sandboxing of apps that refuse to with higher levels

Level 1:
 * `--ro-bind  /bin   /bin`
 * `--dev-bind /dev   /dev` 
 * `--ro-bind  /etc   /etc`
 * `--ro-bind  /lib   /lib`
 * `--ro-bind  /lib32 /lib32`
 * `--ro-bind  /lib64 /lib64`
 * `--ro-bind  /opt   /opt`
 * `--ro-bind  /sbin  /sbin`
 * `--ro-bind  /sys   /sys`
 * `--ro-bind  /usr   /usr`
 * `--ro-bind  [HOME]/.fonts             [SANDBOX HOME]/.fonts`
 * `--ro-bind  [HOME]/.config/fontconfig [SANDBOX HOME]/.config/fontconfig`
 * `--ro-bind  [HOME]/.config/gtk-3.0    [SANDBOX HOME]/.config/gtk-3.0`

Level 2 goes into more specifics within `/usr`, and gives no access to `/etc` by default. It should be used for a large majority of graphical applications

Level 2:
 * `--ro-bind  /bin   /bin`
 * `--ro-bind  /lib   /lib`
 * `--ro-bind  /lib32 /lib32`
 * `--ro-bind  /lib64 /lib64`
 * `--ro-bind  /opt   /opt`
 * `--ro-bind  /sbin  /sbin`
 * `--ro-bind  /sys   /sys`
 * `--ro-bind  /usr/bin   /usr/bin`
 * `--ro-bind  /usr/lib   /usr/lib`
 * `--ro-bind  /usr/lib32 /usr/lib32`
 * `--ro-bind  /usr/lib64 /usr/lib64`
 * `--ro-bind  /usr/sbin  /usr/sbin`
 * `--ro-bind  /usr/share/applications /usr/share/applications`
 * `--ro-bind  /usr/share/fontconfig   /usr/share/fontconfig`
 * `--ro-bind  /usr/share/fonts        /usr/share/fonts`
 * `--ro-bind  /usr/share/glib-2.0     /usr/share/glib-2.0`
 * `--ro-bind  /usr/share/glvnd        /usr/share/glvnd`
 * `--ro-bind  /usr/share/icons        /usr/share/icons`
 * `--ro-bind  /usr/share/libdrm       /usr/share/libdrm`
 * `--ro-bind  /usr/share/mime         /usr/share/mime`
 * `--ro-bind  /usr/share/themes       /usr/share/themes`
 * `--ro-bind  [HOME]/.fonts             [SANDBOX HOME]/.fonts`
 * `--ro-bind  [HOME]/.config/fontconfig [SANDBOX HOME]/.config/fontconfig`
 * `--ro-bind  [HOME]/.config/gtk-3.0    [SANDBOX HOME]/.config/gtk-3.0`

For minimal access, level 3 only gives access to system binaries and libraries

Level 3:
 * `--ro-bind  /bin   /bin`
 * `--ro-bind  /lib   /lib`
 * `--ro-bind  /lib32 /lib32`
 * `--ro-bind  /lib64 /lib64`
 * `--ro-bind  /opt   /opt`
 * `--ro-bind  /sbin  /sbin`
 * `--ro-bind  /usr/bin   /usr/bin`
 * `--ro-bind  /usr/lib   /usr/lib`
 * `--ro-bind  /usr/lib32 /usr/lib32`
 * `--ro-bind  /usr/lib64 /usr/lib64`
 * `--ro-bind  /usr/sbin  /usr/sbin`

For further security or to run an AppImage designed for another distro, you can use `(AppImage) SetRootDir()` to change where it pulls system files from

## Sockets
alsa:
 * `/usr/share/alsa`
 * `/etc/alsa`
 * `/etc/group`
 * `/dev/snd`

audio:
pulseaudio and alsa combined

cgroup:
same as not using `--unshare-cgroup-try` in bwrap

dbus:
 * `$XDG_RUNTIME_DIR/bus`

ipc:
same as not using `--unshare-ipc` in bwrap

network:
 * `/etc/ca-certificates`
 * `/etc/resolv.conf`
 * `/etc/ssl`
 * `/usr/share/ca-certificates`

pid:
same as not using `--unshare-pid` in bwrap

pipewire:
 * `$XDG_RUNTIME_DIR/pipewire-0`

pulseaudio:
 * `$XDG_RUNTIME_DIR/pulse`
 * `/etc/pulse`

session:
same as not using `--new-session` in bwrap

user:
same as not using `--unshare-user-try` in bwrap

uts:
same as not using `--unshare-uts` in bwrap

wayland:
 * `$XDG_RUNTIME_DIR/$WAYLAND_DISPLAY`
 * `/usr/share/x11`

x11:
 * `$XAUTHORITY`
 * `$TMPDIR/.X11-unix/X[DISPLAY]`

## Devices
dri:
 * `/sys/devices/pci000:00`
 * `/dev/nvidiactl`
 * `/dev/nvidia0`
