const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

// TODO: figure out how to add this package correctly
const squashfs = @import("squashfuse-zig/src/main.zig");
pub const SquashFs = squashfs.SquashFs;

const c = @cImport({
    @cInclude("aisap.h");
});

pub const AppImageError = error{
    Error, // Generic error
    NoDesktopEntry,
    InvalidDesktopEntry,
};

pub const AppImage = struct {
    name: []const u8 = undefined,
    path: []const u8,
    desktop_entry: []const u8 = undefined,
    image: SquashFs = undefined,

    // The internal pointer to the C struct
    _internal: *c.aisap_AppImage = undefined,

    pub const Permissions = struct {
        level: u4 = 0,
        files: []const []const u8 = undefined,
        devices: []const []const u8 = undefined,
        sockets: []const []const u8 = undefined,
        data_dir: bool = true,
    };

    // TODO: Use this to replace the Go implemenation
    // Once the Zig version is up to par, the Zig -> C bindings will call to
    // the parent pointer's methods. This cannot currently be done as Go
    // doesn't allow Go pointers to be passed to C
    pub fn init(path: []const u8) !AppImage {
        // Create the AppImage type for the C binding
        var c_ai = c.aisap_AppImage{
            .name = toMut(path).ptr,
            .path = toMut(path).ptr,
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
            .path = path,
            ._internal = &c_ai,
        };

        c_ai._parent = &ai;

        const off = try ai.offset();

        // Open the SquashFS image for reading
        ai.image = try SquashFs.init(ai.path, off);

        var desktop_entry_found = false;
        var walker = try ai.image.walk("");
        while (try walker.next()) |entry| {
            var it = std.mem.splitBackwards(u8, entry.path, ".");

            // Skip any files not ending in `.desktop`
            // Check for both null-terminated and non-null terminated names as
            // I haven't settled on how it should be done with my squashfuse
            // bindings
            const extension = it.first();
            if (!std.mem.eql(u8, extension, "desktop\x00") and !std.mem.eql(u8, extension, "desktop")) continue;

            // Also skip any files without an extension
            if (it.next() == null) continue;

            // Read the first 4KiB of the desktop entry, it really should be a
            // lot smaller than this, but just in case.
            var buf: [1024 * 4]u8 = undefined;
            var inode = try ai.image.getInode(entry.id);
            const read_bytes = try ai.image.readRange(&inode, &buf, 0);
            ai.desktop_entry = buf[0..@intCast(usize, read_bytes)];
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

    pub fn deinit(ai: *AppImage) void {
        ai.image.deinit();
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
        var perms = Permissions{};

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

            // Now get the elements
            // TODO: handle quoting
            var element_it = std.mem.tokenize(u8, key_it.next() orelse "", ";");
            while (element_it.next()) |element| {
                try list.append(element);
            }

            if (std.mem.eql(u8, key, "Level")) {
                const level_slice = try list.toOwnedSlice();
                perms.level = try std.fmt.parseInt(u4, level_slice[0], 10);
            } else if (std.mem.eql(u8, key, "Files")) {
                perms.sockets = try list.toOwnedSlice();
            } else if (std.mem.eql(u8, key, "Devices")) {
                perms.devices = try list.toOwnedSlice();
            } else if (std.mem.eql(u8, key, "Sockets")) {
                perms.sockets = try list.toOwnedSlice();
            }
        }

        return perms;
    }

    pub fn mount(ai: *AppImage) !void {
        //        var sqfs = Squash{};
        _ = ai;
        //        const err = sqfs.lookup("test.sfs");
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

fn toMut(str: []const u8) []u8 {
    var buf: [256]u8 = undefined;
    var mut: []u8 = buf[0..str.len];
    std.mem.copy(u8, mut, str);

    return mut;
}
