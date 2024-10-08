const std = @import("std");
const io = std.io;
const fs = std.fs;
const time = std.time;
const mem = std.mem;
const span = std.mem.span;
const expect = std.testing.expect;
const os = std.os;
const posix = std.posix;
const parser = @import("../parser.zig");

const Md5 = std.crypto.hash.Md5;

const known_folders = @import("known-folders");
const KnownFolder = known_folders.KnownFolder;

const squashfuse = @import("squashfuse");
pub const SquashFs = squashfuse.SquashFs;

const fuse_helper = @import("mount.zig");

const config_parser = @import("profiles.zig").ProfileParser;

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
    mount_dir: ?[:0]const u8 = null,
    allocator: std.mem.Allocator,

    pub const Kind = enum(i3) {
        // Non-standard, shell script implementation built on SquashFS image
        shimg = -2,

        // Standard, built on ISO-9660 disk image
        type1 = 1,

        // Standard, built on SquashFS image
        type2 = 2,
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
        needs_deinit: bool,

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
            const json_database = @embedFile("profile_database.json");

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
                        .needs_deinit = false,
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
                .needs_deinit = true,
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
                        const basename = fs.path.basename(element);

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
            if (self.needs_deinit) {
                if (self.filesystem) |filesystem| {
                    self.allocator.free(filesystem);
                }
                if (self.sockets) |sockets| {
                    self.allocator.free(sockets);
                }
                if (self.devices) |devices| {
                    self.allocator.free(devices);
                }
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
            const basename = fs.path.basename(path_string);

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

        pub var map: ?std.AutoHashMap(
            SocketPermissions,
            []const u8,
        ) = null;

        const JsonSocket = struct {
            name: []const u8,
            flags: ?[]const []const u8 = null,
        };

        // Parses the built-in JSON database into a HashMap
        // TODO: do this comptime
        pub fn initDatabase(allocator: std.mem.Allocator) !void {
            SocketPermissions.map = std.AutoHashMap(
                SocketPermissions,
                []const u8,
            ).init(allocator);

            populateEnviron();
            const sockets_json = @embedFile("../sockets.json");

            const parsed = try std.json.parseFromSlice(
                []JsonSocket,
                allocator,
                sockets_json,
                .{},
            );
            defer parsed.deinit();

            for (parsed.value) |val| {
                std.debug.print("name:  {s}\n", .{val.name});
                try SocketPermissions.map.?.put(
                    std.meta.stringToEnum(SocketPermissions, val.name).?,
                    "teste",
                );

                if (val.flags) |flags| for (flags) |flag| {
                    var it = try parser.SyntaxIterator.init(allocator, flag);
                    while (it.next()) |token| {
                        std.debug.print("flags: {?s}\n", .{token.string});
                        allocator.free(token.string);
                    }
                };

                std.debug.print(
                    "map: {?s}\n",
                    .{SocketPermissions.map.?.get(.x11)},
                );
            }
        }

        fn populateEnviron() void {
            var buf: [4096]u8 = undefined;

            // TODO: Don't rely on $UID being present
            // TODO: Don't hardcore UID
            setDefault("XDG_RUNTIME_DIR", std.fmt.bufPrintZ(
                &buf,
                "/run/user/{s}",
                .{posix.getenv("UID") orelse "1000"},
            ) catch unreachable);
            setDefault("TMPDIR", "/tmp");
        }

        fn setDefault(
            env_var: [:0]const u8,
            default_value: [:0]const u8,
        ) void {
            _ = posix.getenv(
                env_var,
            ) orelse setenv(env_var.ptr, default_value.ptr);
        }

        pub fn fromString(sock: []const u8) !SocketPermissions {
            return std.meta.stringToEnum(
                SocketPermissions,
                sock,
            ) orelse AppImageError.InvalidSocket;
        }

        pub const SocketOptions = struct {
            uid: u32,
            runtime_dir: []const u8,
        };

        fn appendDir(
            allocator: std.mem.Allocator,
            str1: []const u8,
            str2: []const u8,
        ) ![]const u8 {
            return try std.fmt.allocPrint(
                allocator,
                "{s}/{s}",
                .{ str1, str2 },
            );
        }

        //        pub fn toBwrapArgs(allocator: std.mem.Allocator, opts: SocketOptions,) ![]const []const u8 {
        //            _ = opts;
        //
        //            var list = std.ArrayList([]const u8).init(allocator);
        //
        //
        //        }

        /// Generates bwrap args from socket. Memory is cleared on each call
        /// to the method
        pub fn toBwrapArgs(
            socket: SocketPermissions,
            allocator: std.mem.Allocator,
            opts: SocketOptions,
        ) ![]const []const u8 {
            const uid_string = try std.fmt.allocPrint(
                allocator,
                "{d}",
                .{opts.uid},
            );
            defer allocator.free(uid_string);

            return switch (socket) {
                .alsa => &[_][]const u8{
                    "--ro-bind-try", "/usr/share/alsa", "/usr/share/alsa",
                    "--ro-bind-try", "/etc/alsa",       "/etc/alsa",
                    "--ro-bind-try", "/etc/group",      "/etc/group",
                    "--dev-bind",    "/dev/snd",        "/dev/snd",
                },
                .cgroup => &[_][]const u8{},
                .dbus => blk: {
                    const runtime_bus_dir = try appendDir(
                        allocator,
                        opts.runtime_dir,
                        "bus",
                    );
                    defer allocator.free(runtime_bus_dir);

                    const dest_bus_dir = try std.fmt.allocPrint(
                        allocator,
                        "/run/user/{d}/bus",
                        .{opts.uid},
                    );
                    defer allocator.free(dest_bus_dir);

                    var list = std.ArrayList([]const u8).init(allocator);

                    try list.appendSlice(
                        &[_][]const u8{
                            "--ro-bind-try", runtime_bus_dir,         dest_bus_dir,
                            "--ro-bind-try", "/usr/share/alsa",       "/usr/share/alsa",
                            "--ro-bind-try", "/usr/share/pulseaudio", "/usr/share/pulseaudio",
                            "--ro-bind-try", "/etc/alsa",             "/etc/alsa",
                            "--ro-bind-try", "/etc/group",            "/etc/group",
                            "--ro-bind-try", "/etc/pulse",            "/etc/pulse",
                            "--dev-bind",    "/dev/snd",              "/dev/snd",
                        },
                    );

                    break :blk try list.toOwnedSlice();
                },
                .ipc => &[_][]const u8{},
                .network => allocator.dupe([]const u8, &[_][]const u8{
                    "--ro-bind-try", "/etc/ca-certificates",       "/etc/ca-certificates",
                    "--ro-bind-try", "/etc/resolv.conf",           "/etc/resolv.conf",
                    "--ro-bind-try", "/etc/ssl",                   "/etc/ssl",
                    "--ro-bind-try", "/etc/pki",                   "/etc/pki",
                    "--ro-bind-try", "/usr/share/ca-certificates", "/usr/share/ca-certificates",
                    "--share-net",
                }),
                .pid => &[_][]const u8{},
                .pipewire => blk: {
                    const runtime_pipewire_dir = try appendDir(
                        allocator,
                        opts.runtime_dir,
                        "pipewire-0",
                    );
                    defer allocator.free(runtime_pipewire_dir);

                    const dest_pipewire_dir = try std.fmt.allocPrint(
                        allocator,
                        "/run/user/{d}/pipewire-0",
                        .{opts.uid},
                    );
                    defer allocator.free(dest_pipewire_dir);

                    var list = std.ArrayList([]const u8).init(allocator);

                    try list.appendSlice(
                        &[_][]const u8{
                            "--ro-bind-try", runtime_pipewire_dir,    dest_pipewire_dir,
                            "--ro-bind-try", "/usr/share/alsa",       "/usr/share/alsa",
                            "--ro-bind-try", "/usr/share/pulseaudio", "/usr/share/pulseaudio",
                            "--ro-bind-try", "/etc/alsa",             "/etc/alsa",
                            "--ro-bind-try", "/etc/group",            "/etc/group",
                            "--ro-bind-try", "/etc/pulse",            "/etc/pulse",
                            "--dev-bind",    "/dev/snd",              "/dev/snd",
                        },
                    );

                    break :blk try list.toOwnedSlice();
                },
                .session, .user, .uts => &[_][]const u8{},
                else => &[_][]const u8{},
            };
        }
    };

    pub fn open(ai: *AppImage, path: []const u8) !SquashFs.Inode {
        var root_inode = ai.image.getRootInode();
        var it = try root_inode.walk(ai.allocator);
        while (try it.next()) |entry| {
            if (!std.mem.eql(u8, entry.path, path)) continue;

            return entry.inode();
        }

        // TODO: error
        unreachable;
    }

    pub fn init(allocator: std.mem.Allocator, path: []const u8) !AppImage {
        // Create the AppImage type for the C binding
        var ai = AppImage{
            // Name is defined when parsing desktop entry
            .name = undefined,
            .path = try allocator.dupeZ(u8, path),
            .kind = undefined,
            .desktop_entry = undefined,
            .allocator = allocator,
        };

        const off = try ai.offset();

        var appimage_file = try fs.cwd().openFile(path, .{});
        defer appimage_file.close();

        var header_buf: [19]u8 = undefined;
        _ = try appimage_file.read(header_buf[0..]);

        ai.kind = try kindFromHeaderData(&header_buf);

        // Open the SquashFS image for reading
        ai.image = try SquashFs.init(allocator, ai.path, .{ .offset = off });

        var desktop_entry_found = false;
        var root_inode = ai.image.getRootInode();

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

            // If the entry is a symlink, follow it before attempting to read
            if (inode.kind == .sym_link) {
                var path_buf: [std.fs.MAX_PATH_BYTES]u8 = undefined;
                const real_inode_path = try inode.readLink(&path_buf);

                inode = try ai.open(real_inode_path);
            }

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
        self.unmount() catch unreachable;
        self.allocator.free(self.path);
        self.allocator.free(self.name);
        self.image.deinit();
    }

    // Find the offset of the internal read-only filesystem
    pub fn offset(ai: *AppImage) !u64 {
        return offsetFromPath(ai.path);
    }

    /// Returns the bundle path's MD5
    pub fn md5(self: *const AppImage, buf: []u8) ![]const u8 {
        return try md5FromPath(self.path, buf);
    }

    /// Returns the bundle path's MD5, null-terminated
    pub fn md5Z(self: *const AppImage, buf: []u8) ![:0]const u8 {
        return try md5FromPathZ(self.path, buf);
    }

    pub fn permissions(ai: *AppImage, allocator: std.mem.Allocator) !?Permissions {
        var perms = try Permissions.fromDesktopEntry(allocator, ai.desktop_entry);

        if (perms == null) perms = try Permissions.fromName(allocator, ai.name);

        return perms;
    }

    pub const WrapArgsError = error{
        HomeNotFound,
        BundleNotMounted,
        NoPermissions,
        SandboxLevelTooLow,
    };

    // TODO: finish implementing in Zig
    // TODO: free allocated memory. Currently, this should just be allocated using an arena
    pub fn wrapArgs(
        ai: *AppImage,
        allocator: std.mem.Allocator,
        perms: Permissions,
    ) ![]const []const u8 {
        if (ai.mount_dir == null) {
            return WrapArgsError.BundleNotMounted;
        }

        if (perms.level == 0) return WrapArgsError.SandboxLevelTooLow;

        const home = try known_folders.getPath(allocator, .home) orelse {
            return WrapArgsError.HomeNotFound;
        };

        var list = std.ArrayList([]const u8).init(allocator);
        try list.append("bwrap");

        var md5_buf: [33]u8 = undefined;
        const ai_md5 = try ai.md5(&md5_buf);

        // TODO: fallback if `LOGNAME` not present
        const logname = std.posix.getenv("LOGNAME") orelse "";
        const user = try std.process.getUserInfo(logname);

        // Bwrap args for AppImages regardless of level
        try list.appendSlice(&[_][]const u8{
            "--setenv",      "TMPDIR",                             "/tmp",
            "--setenv",      "HOME",                               home,
            "--setenv",      "ARGV0",                              fs.path.basename(ai.path),

            // If these directories are symlinks, they will be resolved,
            // otherwise
            "--ro-bind-try", try resolve(allocator, "/opt"),       "/opt",
            "--ro-bind-try", try resolve(allocator, "/bin"),       "/bin",
            "--ro-bind-try", try resolve(allocator, "/sbin"),      "/sbin",
            "--ro-bind-try", try resolve(allocator, "/lib"),       "/lib",
            "--ro-bind-try", try resolve(allocator, "/lib32"),     "/lib32",
            "--ro-bind-try", try resolve(allocator, "/lib64"),     "/lib64",
            "--ro-bind-try", try resolve(allocator, "/usr/bin"),   "/usr/bin",
            "--ro-bind-try", try resolve(allocator, "/usr/sbin"),  "/usr/sbin",
            "--ro-bind-try", try resolve(allocator, "/usr/lib"),   "/usr/lib",
            "--ro-bind-try", try resolve(allocator, "/usr/lib32"), "/usr/lib32",
            "--ro-bind-try", try resolve(allocator, "/usr/lib64"), "/usr/lib64",

            "--setenv",      "APPDIR",
            try std.fmt.allocPrint(
                allocator,
                "/tmp/.mount_{s}",
                .{ai_md5},
            ),

            "--setenv",      "APPIMAGE",
            try std.fmt.allocPrint(
                allocator,
                "/app/{s}",
                .{fs.path.basename(ai.path)},
            ),

            // Set generic paths for all XDG standard dirs
            "--setenv",      "XDG_DESKTOP_DIR",
            try std.fmt.allocPrint(
                allocator,
                "{s}/Desktop",
                .{home},
            ),
            "--setenv",      "XDG_DOWNLOAD_DIR",
            try std.fmt.allocPrint(
                allocator,
                "{s}/Downloads",
                .{home},
            ),
            "--setenv",      "XDG_DOCUMENTS_DIR",
            try std.fmt.allocPrint(
                allocator,
                "{s}/Documents",
                .{home},
            ),
            "--setenv",      "XDG_MUSIC_DIR",
            try std.fmt.allocPrint(
                allocator,
                "{s}/Music",
                .{home},
            ),
            "--setenv",      "XDG_PICTURES_DIR",
            try std.fmt.allocPrint(
                allocator,
                "{s}/Pictures",
                .{home},
            ),
            "--setenv",      "XDG_VIDEOS_DIR",
            try std.fmt.allocPrint(
                allocator,
                "{s}/Videos",
                .{home},
            ),
            "--setenv",      "XDG_TEMPLATES_DIR",
            try std.fmt.allocPrint(
                allocator,
                "{s}/Templates",
                .{home},
            ),
            "--setenv",      "XDG_PUBLICSHARE_DIR",
            try std.fmt.allocPrint(
                allocator,
                "{s}/Share",
                .{home},
            ),
            "--setenv",      "XDG_DATA_HOME",
            try std.fmt.allocPrint(
                allocator,
                "{s}/.local/share",
                .{home},
            ),
            "--setenv",      "XDG_CONFIG_HOME",
            try std.fmt.allocPrint(
                allocator,
                "{s}/.config",
                .{home},
            ),
            "--setenv",      "XDG_CACHE_HOME",
            try std.fmt.allocPrint(
                allocator,
                "{s}/.cache",
                .{home},
            ),
            "--setenv",      "XDG_STATE_HOME",
            try std.fmt.allocPrint(
                allocator,
                "{s}/.local/state",
                .{home},
            ),
            "--setenv",      "XDG_RUNTIME_DIR",
            try std.fmt.allocPrint(
                allocator,
                "/run/user/{d}",
                .{user.uid},
            ),
            "--perms",    "0700", //
            "--dir",
            try std.fmt.allocPrint(
                allocator,
                "/run/user/{d}",
                .{user.uid},
            ),
            "--dev",      "/dev",
            "--proc",     "/proc",
            "--dir",      "/app",

            "--bind",     ai.path,
            try std.fmt.allocPrint(
                allocator,
                "/app/{s}",
                .{fs.path.basename(ai.path)},
            ),

            // TODO: fallback if no cache
            "--bind-try",
            try std.fmt.allocPrint(
                allocator,
                "{s}/appimage/{s}",
                .{
                    (try known_folders.getPath(allocator, .cache)).?,
                    ai_md5,
                },
            ),
            try std.fmt.allocPrint(
                allocator,
                "{s}/.cache",
                .{home},
            ),

            "--die-with-parent",
        });

        const config = (try known_folders.getPath(allocator, .local_configuration)).?;

        // Per-level arguments
        try list.appendSlice(switch (perms.level) {
            0 => unreachable,
            1 => &[_][]const u8{
                "--dev-bind",    "/dev",         "/dev",
                "--ro-bind",     "/sys",         "/sys",
                "--ro-bind-try", "/usr",         "/usr",
                "--ro-bind-try", "/etc",         "/etc",
                "--ro-bind-try", "/run/systemd", "/run/systemd",
                // TODO: do this a better way
                "--ro-bind-try",
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/.fonts",
                    .{home},
                ),
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/.fonts",
                    .{home},
                ),
                "--ro-bind-try",
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/fontconfig",
                    .{config},
                ),
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/fontconfig",
                    .{config},
                ),
                "--ro-bind-try",
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/gtk-3.0",
                    .{config},
                ),
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/gtk-3.0",
                    .{config},
                ),
                "--ro-bind-try",
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/kdeglobals",
                    .{config},
                ),
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/kdeglobals",
                    .{config},
                ),
                "--ro-bind-try",
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/lxde/lxde.conf",
                    .{config},
                ),
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/lxde/lxde.conf",
                    .{config},
                ),
            },
            2 => &[_][]const u8{
                "--ro-bind-try", "/etc/fonts",              "/etc/fonts",
                "--ro-bind-try", "/etc/ld.so.cache",        "/etc/ld.so.cache",
                "--ro-bind-try", "/etc/mime.types",         "/etc/mime.types",
                "--ro-bind-try", "/usr/share/fontconfig",   "/usr/share/fontconfig",
                "--ro-bind-try", "/usr/share/fonts",        "/usr/share/fonts",
                "--ro-bind-try", "/usr/share/icons",        "/usr/share/icons",
                "--ro-bind-try", "/usr/share/applications", "/usr/share/applications",
                "--ro-bind-try", "/usr/share/mime",         "/usr/share/mime",
                "--ro-bind-try", "/usr/share/libdrm",       "/usr/share/libdrm",
                "--ro-bind-try", "/usr/share/glvnd",        "/usr/share/glvnd",
                "--ro-bind-try", "/usr/share/glib-2.0",     "/usr/share/glib-2.0",
                "--ro-bind-try", "/usr/share/terminfo",     "/usr/share/terminfo",
                "--ro-bind-try",
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/.fonts",
                    .{home},
                ),
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/.fonts",
                    .{home},
                ),
                "--ro-bind-try",
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/fontconfig",
                    .{config},
                ),
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/fontconfig",
                    .{config},
                ),
                "--ro-bind-try",
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/gtk-3.0",
                    .{config},
                ),
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/gtk-3.0",
                    .{config},
                ),
                "--ro-bind-try",
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/kdeglobals",
                    .{config},
                ),
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/kdeglobals",
                    .{config},
                ),
                "--ro-bind-try",
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/lxde/lxde.conf",
                    .{config},
                ),
                try std.fmt.allocPrint(
                    allocator,
                    "{s}/lxde/lxde.conf",
                    .{config},
                ),
            },
            3 => &[_][]const u8{},
        });

        if (perms.data_dir) {
            //                    try list.appendSlice(
            //                        "--bind",
            //                        ai.
            //                    );
        } else {
            try list.appendSlice(&[_][]const u8{
                "--tmpfs",
                home,
            });
        }

        if (perms.filesystem) |files| {
            for (files) |*file| {
                try list.appendSlice(
                    try file.toBwrapArgs(allocator),
                );
            }
        }

        if (perms.sockets) |sockets| {
            for (sockets) |socket| {
                const socket_slice = try socket.toBwrapArgs(allocator, .{
                    // TODO: DO NOT HARD CODE
                    .runtime_dir = "/run/user/1000",
                    .uid = 1000,
                });
                defer allocator.free(socket_slice);

                try list.appendSlice(socket_slice);
            }
        }

        if (ai.mount_dir) |mount_dir| {
            try list.appendSlice(&[_][]const u8{
                "--bind",
                mount_dir,
                try std.fmt.allocPrint(
                    allocator,
                    "/tmp/.mount_{s}",
                    .{
                        ai_md5,
                    },
                ),
            });
        }

        try list.appendSlice(&[_][]const u8{
            "--",
            try std.fmt.allocPrint(
                allocator,
                "/tmp/.mount_{s}/AppRun",
                .{ai_md5},
            ),
        });

        return list.toOwnedSlice();
    }

    /// Returns a `char**` for use in C
    pub fn wrapArgsZ(ai: *AppImage, allocator: std.mem.Allocator, perms: Permissions) ![*:null]?[*:0]const u8 {
        var list = std.ArrayList(?[*:0]const u8).init(allocator);

        var arena = std.heap.ArenaAllocator.init(allocator);
        const arena_allocator = arena.allocator();
        defer arena.deinit();

        const args_slice = try ai.wrapArgs(arena_allocator, perms);

        for (args_slice) |arg| {
            try list.append(try allocator.dupeZ(u8, arg));
            //allocator.free(arg);
        }

        try list.append(null);

        return @ptrCast(try list.toOwnedSlice());
    }

    pub const MountOptions = struct {
        path: ?[]const u8 = null,
        foreground: bool = false,
    };

    fn generateMountDirPath(ai: *AppImage, allocator: std.mem.Allocator) ![:0]const u8 {
        const runtime_dir = try known_folders.getPath(
            allocator,
            .runtime,
        ) orelse unreachable;

        var md5_buf: [33]u8 = undefined;

        return try std.fmt.allocPrintZ(
            allocator,
            "{s}/aisap/mount/{s}",
            .{
                runtime_dir,
                try ai.md5(&md5_buf),
            },
        );
    }

    pub fn mount(ai: *AppImage, opts: MountOptions) !void {
        ai.mount_dir = if (opts.path) |path| blk: {
            break :blk try ai.allocator.dupeZ(
                u8,
                path,
            );
        } else blk: {
            break :blk try ai.generateMountDirPath(
                ai.allocator,
            );
        };

        errdefer {
            ai.allocator.free(ai.mount_dir.?);
            ai.mount_dir = null;
        }

        const cwd = fs.cwd();

        // ai.mount_dir defined above, this shouldn't ever be null
        cwd.makePath(ai.mount_dir.?) catch |err| {
            if (err != error.PathAlreadyExists) {}
        };

        const off = try ai.offset();

        if (opts.foreground) {
            @panic("not yet implemented");
            //try fuse_helper.mountImage(ai.path, ai.mount_dir.?, off);
        } else {
            // TODO: fix squashfuse-zig crash here
            if (false) {
                try fuse_helper.mountImage(
                    ai.path,
                    ai.mount_dir.?,
                    off,
                );
            } else {
                const offset_string = try std.fmt.allocPrint(
                    ai.allocator,
                    "-ooffset={d}",
                    .{off},
                );
                defer ai.allocator.free(offset_string);

                var proc = std.process.Child.init(&[_][]const u8{
                    "squashfuse",
                    offset_string,
                    ai.path,
                    ai.mount_dir.?,
                }, ai.allocator);

                _ = try proc.spawnAndWait();
            }
        }
    }

    pub fn unmount(ai: *AppImage) !void {
        if (ai.mount_dir) |mount_dir| {
            var umount = std.process.Child.init(&[_][]const u8{
                "fusermount",
                "-u",
                mount_dir,
            }, ai.allocator);

            _ = try umount.spawnAndWait();

            ai.allocator.free(mount_dir);
            ai.mount_dir = null;
        }
    }

    pub const SandboxOptions = struct {
        args: ?[]const []const u8 = null,
    };

    // This can't be finished until AppImage.wrapArgs works correctly
    pub fn sandbox(ai: *AppImage, perms: Permissions, opts: SandboxOptions) !void {
        var arena = std.heap.ArenaAllocator.init(ai.allocator);
        const arena_allocator = arena.allocator();
        defer arena.deinit();

        const wrap_args = try ai.wrapArgs(arena_allocator, perms);

        if (opts.args) |args| {
            var list = std.ArrayList([]const u8).init(arena_allocator);
            defer list.deinit();

            try list.appendSlice(wrap_args);
            try list.appendSlice(args);

            try bwrap(arena_allocator, list.items);
        } else {
            try bwrap(arena_allocator, wrap_args);
        }
    }
};

// Requires libc to be linked for now
extern fn setenv(key: [*:0]const u8, value: [*:0]const u8) c_int;

// If path is a symlink, return its target. Otherwise, just return the path
fn resolve(allocator: std.mem.Allocator, path: []const u8) ![]const u8 {
    const cwd = fs.cwd();
    var file = cwd.openFile(path, .{}) catch return path;
    const stat = try file.stat();

    if (stat.kind != .sym_link) return path;

    const buf = try allocator.alloc(u8, std.fs.MAX_PATH_BYTES);
    return std.posix.readlink(path, buf);
}

pub fn md5FromPath(path: []const u8, buf: []u8) ![]const u8 {
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
    return std.fmt.bufPrint(buf, "{x}", .{
        std.fmt.fmtSliceHexLower(&md5_buf),
    }) catch unreachable;
}

pub fn md5FromPathZ(path: []const u8, buf: []u8) ![:0]const u8 {
    _ = try md5FromPath(path, buf[0 .. buf.len - 1]);
    buf[buf.len - 1] = '\x00';
}

pub fn offsetFromPath(path: []const u8) !u64 {
    var f = try fs.cwd().openFile(path, .{});
    defer f.close();

    // Get header
    // Buffer of 19 as it is shImg's header size.
    // Normal AppImages can be detected by reading 10 bytes
    var header_buf: [19]u8 = undefined;
    _ = try f.read(header_buf[0..]);
    try f.seekTo(0);

    const kind = try kindFromHeaderData(&header_buf);

    if (kind == .shimg) {
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

    if (kind == .type1 or kind == .type2) {
        // Read ELF offset
        const hdr = try std.elf.Header.read(f);

        return hdr.shoff + hdr.shentsize * hdr.shnum;
    }

    unreachable;
}

pub const AppImageHeaderError = error{
    NotEnoughData,
    InvalidHeader,
    UnknownAppImageVersion,
};

/// Returns the bundle's type using header information.
/// The header_data must be at least 19 bytes
pub fn kindFromHeaderData(header_data: []const u8) !AppImage.Kind {
    if (header_data.len < 4) {
        return AppImageHeaderError.NotEnoughData;
    }

    if (header_data.len >= 19 and std.mem.eql(u8, header_data[0..19], "#!/bin/sh\n#.shImg.#")) {
        return .shimg;
    } else if (std.mem.eql(u8, header_data[0..4], "\x7fELF")) {
        // ELF file, now we need to check if it has the special AppImage
        // magic bytes

        if (header_data.len < 11) {
            return AppImageHeaderError.NotEnoughData;
        }

        if (std.mem.eql(u8, header_data[8..10], "AI")) {
            // Immediately following the "AI" bytes is the AppImage version
            return switch (header_data[10]) {
                1 => .type1,
                2 => .type2,

                else => AppImageHeaderError.UnknownAppImageVersion,
            };
        }
    }

    return AppImageHeaderError.InvalidHeader;
}

// TODO: eventually build bwrap as a library
// This shouldn't typically be called on its own. Instead, `AppImage.sandbox`
// should be preferred
pub fn bwrap(allocator: std.mem.Allocator, args: []const []const u8) !void {
    const stdout = std.io.getStdOut();
    const stderr = std.io.getStdErr();

    var child = std.process.Child.init(args, allocator);
    child.stdout = stdout;
    child.stderr = stderr;
    _ = try child.spawnAndWait();
}

// TODO: get bundle's supported architecture list
