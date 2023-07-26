const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

const Md5 = std.crypto.hash.Md5;

const squashfs = @import("squashfuse");
pub const SquashFs = squashfs.SquashFs;

// TODO: mounting in Zig
const mountHelper = @import("mount.zig");

pub const c = @cImport({
    @cInclude("aisap.h");
});

pub const c_AppImage = c.aisap_appimage;

pub const AppImageError = error{
    Error, // Generic error
    InvalidMagic,
    NoDesktopEntry,
    InvalidDesktopEntry,
    InvalidSocket,
    NoSpaceLeft,
};

pub const AppImage = struct {
    name: [:0]const u8,
    path: [:0]const u8,
    desktop_entry: [:0]const u8 = undefined,
    image: SquashFs = undefined,
    kind: Kind,
    allocator: std.mem.Allocator,

    // The internal pointer to the C struct
    _internal: ?*c_AppImage = null,

    pub const Kind = enum(i3) {
        shimg = -2,

        type1 = 1,
        type2,
    };

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
            return std.meta.stringToEnum(Socket, sock) orelse AppImageError.InvalidSocket;
        }
    };

    // TODO: Use this to replace the Go implemenation
    // Once the Zig version is up to par, the Zig -> C bindings will call to
    // the parent pointer's methods. This cannot currently be done as Go
    // doesn't allow Go pointers to be passed to C
    pub fn init(allocator: std.mem.Allocator, path: []const u8) !AppImage {
        // Create the AppImage type for the C binding
        var ai = AppImage{
            // Name is defined when parsing desktop entry
            .name = undefined,
            .path = try allocator.dupeZ(u8, path),
            .kind = .type2,
            .allocator = allocator,
        };

        const off = try ai.offset();

        // Open the SquashFS image for reading
        ai.image = try SquashFs.init(allocator, ai.path, .{ .offset = off });

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
            const read_bytes = try inode.read(&entry_buf);

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

            // TODO: guess name based on filename if none set in desktop entry
            if (std.mem.eql(u8, key, "Name")) {
                ai.name = try allocator.dupeZ(u8, key_it.next() orelse "");

                break;
            }
        }

        return ai;
    }

    // TODO
    pub fn deinit(self: *AppImage) void {
        _ = self;
        //self.image.deinit();
        //self.allocator.free(self.path);
        //self.allocator.free(self.name);
        //self.freePermissions();
    }

    // Find the offset of the internal read-only filesystem
    pub fn offset(ai: *AppImage) !u64 {
        return offsetFromPath(ai.path);
    }

    pub fn md5(self: *const AppImage, buf: []u8) ![:0]const u8 {
        return try md5FromPath(self.path, buf);
    }

    // TODO
    pub fn wrapArgs_old(ai: *AppImage, allocator: std.mem.Allocator) [][]const u8 {
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

            // TODO: refactor
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

    //    pub fn freePermissions(self: *AppImage) void {
    //        if (self.permissions.files) {
    //            self.allocator.free(self.permissions.files);
    //        }
    //        if (self.permissions.devices) {
    //            self.allocator.free(self.permissions.devices);
    //        }
    //        if (self.perimssions.sockets) {
    //            self.allocator.free(self.permissions.sockets);
    //        }
    //    }

    extern fn aisap_appimage_wraparg_next_go(*c_AppImage, *i32) ?[*:0]const u8;

    // TODO: implement in Zig
    // This will return `![]const []const u8` once reimplemented
    // Currently returns `![*:null]?[*:0]const u8` for easier C interop
    pub fn wrapArgs(ai: *AppImage, allocator: std.mem.Allocator) ![*:null]?[*:0]const u8 {
        var wrapargs_list = std.ArrayList(?[*:0]const u8).init(allocator);

        var arg_len: c_int = undefined;

        while (aisap_appimage_wraparg_next_go(ai._internal.?, &arg_len)) |arg| {
            try wrapargs_list.append(arg);
        }

        try wrapargs_list.append(null);

        return @ptrCast(wrapargs_list.items.ptr);
    }

    pub const MountOptions = struct {
        path: ?[]const u8 = null,
    };

    pub fn mount(ai: *AppImage, opts: MountOptions) !void {
        // TODO: proper temp dir
        //        std.debug.print("AppImage.zig test\n", .{});
        const mount_dir = opts.path orelse "/tmp/mountTemp";

        //        std.debug.print("AppImage.zig test2 {s}\n", .{mount_dir});
        try mountHelper.mountImage(ai.path, mount_dir, try ai.offset());
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

pub fn md5FromPath(path: []const u8, buf: []u8) ![:0]const u8 {
    if (buf.len < Md5.digest_length * 2 + 1) {
        return AppImageError.NoSpaceLeft;
    }

    var md5_buf: [Md5.digest_length]u8 = undefined;

    // Generate the MD5
    var h = Md5.init(.{});
    h.update("file://");
    h.update(path);
    h.final(&md5_buf);

    // Format as hexadecimal instead of raw bytes
    return std.fmt.bufPrintZ(buf, "{x}", .{
        std.fmt.fmtSliceHexLower(&md5_buf),
    }) catch unreachable;
}

pub fn offsetFromPath(path: []const u8) !u64 {
    var f = try fs.cwd().openFile(path, .{});
    defer f.close();

    // Get header
    // Buffer of 19 as it is shImg's header size.
    // Normal AppImages can be detected by reading 10 bytes
    var hdr_buf: [19]u8 = undefined;
    _ = try f.read(hdr_buf[0..]);
    try f.seekTo(0);

    // TODO: function to detect type
    if (std.mem.eql(u8, hdr_buf[0..], "#!/bin/sh\n#.shImg.#")) {
        // Read shImg offset
        var buf_reader = io.bufferedReader(f.reader());
        var in_stream = buf_reader.reader();

        // Small buffer needed, the `sfs_offset` line should be well below this amount
        var buf: [256]u8 = undefined;

        var line: u32 = 0;
        while (try in_stream.readUntilDelimiterOrEof(&buf, '\n')) |text| {
            // Iterated over too many lines, not shImg
            line += 1;
            if (line > 512) return AppImageError.InvalidMagic;

            if (text.len > 10 and std.mem.eql(u8, text[0..11], "sfs_offset=")) {
                var it = std.mem.tokenize(u8, text, "=");

                // Throw away first chunk, should equal `sfs_offset`
                _ = it.next();

                return try std.fmt.parseInt(u64, it.next().?, 0);
            }
        }
    }

    // Read ELF offset
    const hdr = try std.elf.Header.read(f);

    return hdr.shoff + hdr.shentsize * hdr.shnum;
}
