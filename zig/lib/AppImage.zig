const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

// TODO: figure out how to add this package correctly
const squashfs = @import("squashfuse-zig/lib.zig");
pub const SquashFs = squashfs.SquashFs;

const c = @cImport({
    @cInclude("aisap.h");
});

pub const AppImageError = error{
    Error, // Generic error
    NoDesktopEntry,
    InvalidDesktopEntry,
    InvalidSocket,
};

pub const AppImage = struct {
    name: []const u8,
    path: []const u8,
    desktop_entry: [:0]const u8 = undefined,
    image: SquashFs = undefined,

    // The internal pointer to the C struct
    _internal: *c.aisap_appimage = undefined,

    pub const Permissions = struct {
        level: u2,
        files: ?[]const []const u8,
        devices: ?[]const []const u8,
        sockets: ?[]Socket,
        data_dir: bool = true,
    };

    pub const FilePermissions = struct {
        path: []const u8,
        read_only: bool,
    };

    pub const Socket = enum {
        alsa,
        audio,
        cgroup,
        dbus,
        ipc,
        network,
        pid,
        pipewire,
        pulseaudio,
        session,
        user,
        uts,
        wayland,
        x11,

        pub fn fromString(sock: []const u8) !Socket {
            if (std.mem.eql(u8, sock, "alsa")) {
                return .alsa;
            } else if (std.mem.eql(u8, sock, "audio")) {
                return .audio;
            } else if (std.mem.eql(u8, sock, "cgroup")) {
                return .cgroup;
            } else if (std.mem.eql(u8, sock, "ipc")) {
                return .ipc;
            } else if (std.mem.eql(u8, sock, "network")) {
                return .network;
            } else if (std.mem.eql(u8, sock, "pid")) {
                return .pid;
            } else if (std.mem.eql(u8, sock, "pipewire")) {
                return .pipewire;
            } else if (std.mem.eql(u8, sock, "pulseaudio")) {
                return .pulseaudio;
            } else if (std.mem.eql(u8, sock, "session")) {
                return .session;
            } else if (std.mem.eql(u8, sock, "user")) {
                return .user;
            } else if (std.mem.eql(u8, sock, "uts")) {
                return .uts;
            } else if (std.mem.eql(u8, sock, "wayland")) {
                return .wayland;
            } else if (std.mem.eql(u8, sock, "x11")) {
                return .x11;
            }

            return AppImageError.InvalidSocket;
        }
    };

    // TODO: Use this to replace the Go implemenation
    // Once the Zig version is up to par, the Zig -> C bindings will call to
    // the parent pointer's methods. This cannot currently be done as Go
    // doesn't allow Go pointers to be passed to C
    pub fn init(allocator: std.mem.Allocator, path: []const u8) !AppImage {
        // Create the AppImage type for the C binding
        var c_ai = c.aisap_appimage{
            .name = path.ptr,
            .path = path.ptr,
            .data_dir = undefined,
            .temp_dir = undefined,
            .root_dir = undefined,
            .mount_dir = undefined,
            .md5 = undefined,
            .run_id = undefined,
            ._index = 0,
            ._parent = undefined,
            .ai_type = 2,
        };

        var ai = AppImage{
            .name = undefined,
            .path = path,
            ._internal = &c_ai,
        };

        c_ai._parent = &ai;

        const off = try ai.offset();

        // Open the SquashFS image for reading
        ai.image = try SquashFs.init(allocator, ai.path, off);

        var desktop_entry_found = false;
        var root_inode = try ai.image.getInode(ai.image.internal.sb.root_inode);
        var it = try root_inode.iterate();
        while (try it.next()) |entry| {
            var split_it = std.mem.splitBackwards(u8, entry.name, ".");

            // Skip any files not ending in `.desktop`
            const extension = split_it.first();
            if (!std.mem.eql(u8, extension, "desktop")) continue;

            // Also skip any files without an extension
            if (split_it.next() == null) continue;

            // Read the first 4KiB of the desktop entry, it really should be a
            // lot smaller than this, but just in case.
            var entry_buf: [1024 * 4]u8 = undefined;
            var inode = entry.inode();

            // TODO: error checking buffer size
            const read_bytes = try inode.read(&entry_buf, 0);

            // Append null byte for use in C
            entry_buf[read_bytes] = '\x00';

            ai.desktop_entry = entry_buf[0..read_bytes :0];

            desktop_entry_found = true;

            break;
        }

        if (!desktop_entry_found) return AppImageError.NoDesktopEntry;

        var line_it = std.mem.tokenize(u8, ai.desktop_entry, "\n");
        var in_desktop_section = false;
        while (line_it.next()) |line| {
            if (std.mem.eql(u8, line, "[Desktop Entry]")) {
                in_desktop_section = true;
                continue;
            }

            var key_it = std.mem.tokenize(u8, line, "=");
            const key = key_it.next() orelse "";

            if (std.mem.eql(u8, key, "Name")) {
                ai.name = key_it.next() orelse "";
            }
        }

        return ai;
    }

    pub fn deinit(self: *AppImage) void {
        self.image.deinit();
    }

    // Find the offset of the internal read-only filesystem
    pub fn offset(ai: *AppImage) !u64 {
        var f = try fs.cwd().openFile(ai.path, .{});
        const hdr = try std.elf.Header.read(f);

        return hdr.shoff + hdr.shentsize * hdr.shnum;
    }

    pub fn wrapArgs(ai: *AppImage, allocator: std.mem.Allocator) [][]const u8 {
        // Need an allocator as the size of `cmd_args` will change size

        //var cmd_args: [][]const u8 = undefined;
        var cmd_args = allocator.alloc([]const u8, 2) catch unreachable;
        cmd_args[0] = "test";
        cmd_args[1] = "test2";

        _ = ai;

        return cmd_args;
    }

    pub fn permissions(ai: *AppImage, allocator: std.mem.Allocator) !Permissions {
        var perms = Permissions{
            .level = 3,
            .files = null,
            .sockets = null,
            .devices = null,
        };

        // Find the permissions section of the INI file, then actually
        // parse the permissions
        var line_it = std.mem.tokenize(u8, ai.desktop_entry, "\n");
        var in_permissions_section = false;
        while (line_it.next()) |line| {
            if (std.mem.eql(u8, line, "[X-App Permissions]")) {
                in_permissions_section = true;
                continue;
            }

            if (!in_permissions_section) continue;

            // Obtain the key name
            // eg `Level=3` becomes `Level`
            var key_it = std.mem.tokenize(u8, line, "=");
            const key = key_it.next() orelse "";

            var list = std.ArrayList([]const u8).init(allocator);
            var sock_list = std.ArrayList(Socket).init(allocator);

            // Now get the elements
            // TODO: handle quoting
            var element_it = std.mem.tokenize(u8, key_it.next() orelse "", ";");

            if (!std.mem.eql(u8, key, "Sockets")) {
                while (element_it.next()) |element| {
                    try list.append(element);
                }
            } else {
                while (element_it.next()) |sock| {
                    try sock_list.append(try Socket.fromString(sock));
                }
            }

            if (std.mem.eql(u8, key, "Level")) {
                const level_slice = try list.toOwnedSlice();
                perms.level = try std.fmt.parseInt(u2, level_slice[0], 10);
            } else if (std.mem.eql(u8, key, "Files")) {
                perms.files = try list.toOwnedSlice();
            } else if (std.mem.eql(u8, key, "Devices")) {
                perms.devices = try list.toOwnedSlice();
            } else if (std.mem.eql(u8, key, "Sockets")) {
                perms.sockets = try sock_list.toOwnedSlice();
            } else if (std.mem.eql(u8, key, "DataDir")) {
                perms.data_dir = std.mem.eql(u8, list.items[0], "true");
                list.deinit();
            }
        }

        return perms;
    }

    pub fn freePermissions(self: *AppImage) void {
        self.allocator.free(self.permissions.files);
        self.allocator.free(self.permissions.devices);
        self.allocator.free(self.permissions.sockets);
    }

    pub fn mount(ai: *AppImage) !void {
        _ = ai;
    }

    // This can't be finished until AppImage.wrapArgs works correctly
    //    pub fn sandbox(ai: *AppImage, allocator: *std.mem.Allocator) !void {
    //        const cmd = [_][]const u8 {
    //            "--ro-bind", "/", "/",
    //            "sh",
    //        };
    //
    //        _ = ai;
    //        _ = try bwrap(allocator, &cmd);
    //    }
};
