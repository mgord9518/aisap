const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;
const os = std.os;

const Md5 = std.crypto.hash.Md5;

const known_folders = @import("known-folders");
const KnownFolder = known_folders.KnownFolder;

const squashfuse = @import("squashfuse");
pub const SquashFs = squashfuse.SquashFs;

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
    desktop_entry: [:0]const u8,
    image: SquashFs = undefined,
    kind: Kind,
    allocator: std.mem.Allocator,

    // This will only get populated if using the C bindings
    _internal: ?*c_AppImage = null,

    pub const Kind = enum(i3) {
        shimg = -2,

        type1 = 1,
        type2,
    };

    const JsonPermissions = struct {
        names: []const []const u8,
        level: u2,
        filesystem: ?[]const []const u8,
        devices: ?[]const []const u8,
        sockets: ?[]const []const u8,
        data_dir: bool = true,
    };

    pub const Permissions = struct {
        level: u2,
        filesystem: ?[]FilesystemPermissions,
        devices: ?[]const []const u8,
        sockets: ?[]SocketPermissions,
        data_dir: bool = true,
        origin: Origin,

        allocator: std.mem.Allocator,

        pub const Origin = enum {
            desktop_entry,
            profile_database,

            // TODO: will be based on appstream metainfo
            bundle_id,
        };

        var permissions_database: ?std.StringHashMap(Permissions) = null;

        // Parses the built-in JSON database into a HashMap
        // TODO: do this comptime
        fn initDatabase(allocator: std.mem.Allocator) !std.StringHashMap(Permissions) {
            const json_database = @embedFile("../profile_database.json");

            const parsed = try std.json.parseFromSlice(
                []JsonPermissions,
                allocator,
                json_database,
                .{},
            );

            defer parsed.deinit();

            var hash_map = std.StringHashMap(Permissions).init(allocator);

            for (parsed.value) |item| {
                for (item.names) |name| {
                    const filesystem = if (item.filesystem) |files| blk: {
                        var file_list = std.ArrayList(
                            FilesystemPermissions,
                        ).init(allocator);

                        for (files) |file| {
                            try file_list.append(try FilesystemPermissions.fromString(
                                allocator,
                                file,
                            ));
                        }

                        break :blk try file_list.toOwnedSlice();
                    } else null;

                    const sockets = if (item.sockets) |sockets| blk: {
                        var socket_list = std.ArrayList(
                            SocketPermissions,
                        ).init(allocator);

                        for (sockets) |socket| {
                            try socket_list.append(try SocketPermissions.fromString(
                                socket,
                            ));
                        }

                        break :blk try socket_list.toOwnedSlice();
                    } else null;

                    try hash_map.put(name, .{
                        .level = item.level,
                        .filesystem = filesystem,
                        .sockets = sockets,
                        .devices = null,
                        .origin = .profile_database,
                        .allocator = allocator,
                    });
                }
            }

            return hash_map;
        }

        /// Returns the permissions based on its (case insensitive) name
        /// The caller should free the returned memory with `Permissions.deinit()`
        pub fn fromName(allocator: std.mem.Allocator, name: []const u8) !?Permissions {
            if (Permissions.permissions_database == null) {
                Permissions.permissions_database = try initDatabase(allocator);
            }

            var name_buf: [256]u8 = undefined;

            const lowercase_name = std.ascii.lowerString(&name_buf, name);

            return Permissions.permissions_database.?.get(lowercase_name);
        }

        pub fn fromDesktopEntry(allocator: std.mem.Allocator, desktop_entry: []const u8) !?Permissions {
            var perms = Permissions{
                .level = 3,
                .filesystem = null,
                .sockets = null,
                .devices = null,
                .origin = .desktop_entry,
                .allocator = allocator,
            };

            // Find the permissions section of the INI file, then actually
            // parse the permissions
            var line_it = std.mem.tokenize(u8, desktop_entry, "\n");
            var in_permissions_section = false;
            var permissions_section_found = false;
            while (line_it.next()) |line| {
                if (std.mem.eql(u8, line, "[X-App Permissions]")) {
                    in_permissions_section = true;
                    permissions_section_found = true;
                    continue;
                } else if (line[0] == '[' and in_permissions_section) {
                    in_permissions_section = false;
                    break;
                }

                if (!in_permissions_section) continue;

                // Obtain the key name
                // eg `Level=3` becomes `Level`
                var key_it = std.mem.tokenize(u8, line, "=");
                const key = key_it.next() orelse "";

                var list = std.ArrayList([]const u8).init(allocator);
                var sock_list = std.ArrayList(SocketPermissions).init(allocator);
                var file_list = std.ArrayList(FilesystemPermissions).init(allocator);

                // Now get the elements
                // TODO: handle quoting
                var element_it = std.mem.tokenize(u8, key_it.next() orelse "", ";");

                // TODO: refactor
                if (std.mem.eql(u8, key, "Sockets")) {
                    while (element_it.next()) |sock| {
                        try sock_list.append(try SocketPermissions.fromString(sock));
                    }
                } else if (std.mem.eql(u8, key, "Files")) {
                    while (element_it.next()) |element| {
                        const basename = std.fs.path.basename(element);

                        var split_it = std.mem.split(u8, basename, ":");
                        const writable = std.mem.eql(u8, split_it.first(), "rw");

                        try file_list.append(.{
                            .src_path = try allocator.dupeZ(u8, element),
                            .dest_path = try allocator.dupeZ(u8, element),
                            .writable = writable,
                            .allocator = allocator,
                        });
                    }
                } else {
                    while (element_it.next()) |element| {
                        try list.append(element);
                    }
                }

                if (std.mem.eql(u8, key, "Level")) {
                    const level_slice = try list.toOwnedSlice();
                    perms.level = try std.fmt.parseInt(u2, level_slice[0], 10);
                } else if (std.mem.eql(u8, key, "Filesystem") or std.mem.eql(u8, key, "Files")) {
                    perms.filesystem = try file_list.toOwnedSlice();
                } else if (std.mem.eql(u8, key, "Devices")) {
                    perms.devices = try list.toOwnedSlice();
                } else if (std.mem.eql(u8, key, "Sockets")) {
                    perms.sockets = try sock_list.toOwnedSlice();
                } else if (std.mem.eql(u8, key, "DataDir")) {
                    perms.data_dir = std.mem.eql(u8, list.items[0], "true");
                    list.deinit();
                }
            }

            if (permissions_section_found) {
                return perms;
            }

            if (perms.filesystem) |filesystem| {
                allocator.free(filesystem);
            }
            if (perms.sockets) |sockets| {
                allocator.free(sockets);
            }
            if (perms.devices) |devices| {
                allocator.free(devices);
            }

            return null;
        }

        pub fn deinit(self: *Permissions) void {
            if (self.files) |files| {
                self.allocator.free(files);
            }
            if (self.sockets) |sockets| {
                self.allocator.free(sockets);
            }
            if (self.devices) |devices| {
                self.allocator.free(devices);
            }
        }
    };

    pub const FilesystemPermissions = struct {
        writable: bool,

        // The file's real location
        src_path: [:0]const u8,

        // Where the file will be exposed inside the sanbox
        dest_path: [:0]const u8,

        allocator: std.mem.Allocator,

        pub const XdgDirs = &[_][]const u8{};

        /// Converts xdg string to KnownFolder
        /// TODO: fix xdg-state, xdg-templates
        pub fn xdgStringToKnownFolder(path: []const u8) ?KnownFolder {
            if (std.mem.eql(u8, path, "xdg-home")) {
                return .home;
            } else if (std.mem.eql(u8, path, "xdg-documents")) {
                return .documents;
            } else if (std.mem.eql(u8, path, "xdg-pictures")) {
                return .pictures;
            } else if (std.mem.eql(u8, path, "xdg-music")) {
                return .music;
            } else if (std.mem.eql(u8, path, "xdg-videos")) {
                return .videos;
            } else if (std.mem.eql(u8, path, "xdg-desktop")) {
                return .desktop;
            } else if (std.mem.eql(u8, path, "xdg-download")) {
                return .downloads;
            } else if (std.mem.eql(u8, path, "xdg-publicshare")) {
                return .public;
            } else if (std.mem.eql(u8, path, "xdg-templates")) {
                return null;
            } else if (std.mem.eql(u8, path, "xdg-config")) {
                return .local_configuration;
            } else if (std.mem.eql(u8, path, "xdg-cache")) {
                return .cache;
            } else if (std.mem.eql(u8, path, "xdg-data")) {
                return .data;
            } else if (std.mem.eql(u8, path, "xdg-state")) {
                return null;
            }

            return null;
        }

        //        fn xdgExpand(allocator: std.mem.Allocator, path: []const u8) ![]const u8 {
        //            var top_directory: []const u8 = "";
        //
        //            const top_directory = for (path, 0..) |char, idx| blk: {
        //                if (char == '/') {
        //                    break :blk path[0..idx];
        //                }
        //            }
        //
        //            const known_folder = try xdgStringToKnownFolder(top_directory);
        //        }

        /// Converts a file string such as `~/Downloads:rw` into a FilesystemPermissions structure
        /// Unfortunately, this only requires an allocator because it'll need
        /// to handle XDG prefixes
        /// TODO: handle XDG prefixes
        pub fn fromString(allocator: std.mem.Allocator, path_string: []const u8) !FilesystemPermissions {
            const basename = std.fs.path.basename(path_string);

            var split_it = std.mem.splitBackwards(u8, basename, ":");
            const writable = std.mem.eql(u8, split_it.first(), "rw");

            const contains_writable_postfix = split_it.next() != null;

            const real_src = if (contains_writable_postfix) blk: {
                break :blk try allocator.dupeZ(u8, path_string[0 .. path_string.len - 3]);
            } else blk: {
                break :blk try allocator.dupeZ(u8, path_string);
            };

            return .{
                .src_path = real_src,
                .dest_path = real_src,
                .writable = writable,

                .allocator = allocator,
            };
        }

        // You probably shouldn't need to call this manually
        pub fn deinit(self: *FilesystemPermissions) void {
            self.allocator.free(self.src_path);
        }

        /// Creates bwrap arguments from file
        /// Caller must free returned slice if not null
        pub fn toBwrapArgs(self: *FilesystemPermissions, allocator: std.mem.Allocator) ![]const []const u8 {
            var list = std.ArrayList([]const u8).init(allocator);

            if (self.writable) {
                try list.append("--bind-try");
            } else {
                try list.append("--ro-bind-try");
            }

            try list.appendSlice(&[_][]const u8{
                self.src_path,
                self.dest_path,
            });

            return list.toOwnedSlice();
        }
    };

    pub const SocketPermissions = enum {
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

        pub fn fromString(sock: []const u8) !SocketPermissions {
            return std.meta.stringToEnum(
                SocketPermissions,
                sock,
            ) orelse AppImageError.InvalidSocket;
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
            .desktop_entry = undefined,
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
            var entry_buf = try allocator.alloc(u8, 1024 * 4);
            var inode = entry.inode();

            // When reading, save the last byte for null terminator
            const read_bytes = try inode.read(entry_buf[0 .. entry_buf.len - 2]);

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
        self.allocator.free(self.desktop_entry);
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

    pub fn permissions(ai: *AppImage, allocator: std.mem.Allocator) !?Permissions {
        var perms = try Permissions.fromDesktopEntry(allocator, ai.desktop_entry);

        if (perms == null) perms = try Permissions.fromName(allocator, ai.name);

        return perms;
    }

    extern fn aisap_appimage_wraparg_next_go(*c_AppImage, *usize) ?[*:0]const u8;

    // TODO: implement in Zig
    // This will return `![]const []const u8` once reimplemented
    pub fn wrapArgs(ai: *AppImage, allocator: std.mem.Allocator) ![][]const u8 {
        var list = std.ArrayList([]const u8).init(allocator);

        var perms = try ai.permissions(allocator);
        //        defer perms.deinit();

        std.debug.print("{}\n", .{perms});

        if (perms.filesystem) |files| {
            for (files) |*file| {
                try list.appendSlice(
                    try file.toBwrapArgs(allocator),
                );
            }
        }

        return list.toOwnedSlice();
    }

    pub fn wrapArgsZ(ai: *AppImage, allocator: std.mem.Allocator) ![*:null]?[*:0]const u8 {
        var list = std.ArrayList(?[*:0]const u8).init(allocator);

        const args_slice = try ai.wrapArgs(allocator);

        for (args_slice) |arg| {
            try list.append(try allocator.dupeZ(u8, arg));
            allocator.free(arg);
        }

        try list.append(null);

        return @ptrCast(try list.toOwnedSlice());
    }

    pub const MountOptions = struct {
        path: ?[]const u8 = null,
        foreground: bool = false,
    };

    // TODO: refactor
    fn getMountDir(ai: *AppImage, buf: []u8) ![]const u8 {
        var allocator_buf: [4096]u8 = undefined;
        var fba = std.heap.FixedBufferAllocator.init(&allocator_buf);
        const allocator = fba.allocator();

        const runtime_dir = try known_folders.getPath(allocator, .runtime) orelse unreachable;

        var md5_buf: [33]u8 = undefined;
        return try std.fmt.bufPrint(buf, "{s}/aisap/mount/{s}", .{ runtime_dir, try ai.md5(&md5_buf) });
    }

    pub fn mount(ai: *AppImage, opts: MountOptions) !void {
        var buf: [os.PATH_MAX]u8 = undefined;

        //        const cwd = fs.cwd();
        //        cwd.makePath(runtime_dir) catch |err| {
        //            if (err != os.MakeDirError.PathAlreadyExists) {}
        //        };

        const mount_dir = opts.path orelse try ai.getMountDir(&buf);

        const off = try ai.offset();

        if (opts.foreground) {
            try mountHelper.mountImage(ai.path, mount_dir, off);
        } else {
            _ = try std.Thread.spawn(
                .{},
                mountHelper.mountImage,
                .{ ai.path, mount_dir, off },
            );
        }
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
