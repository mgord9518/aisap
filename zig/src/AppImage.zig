const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;
const squashfs = @import("squashfs");
const Squash = squashfs.SquashFs;

const c = @cImport({
    @cInclude("aisap.h");
});

pub const AppImage = struct {
    name: []const u8,
    path: []const u8,

    // The internal pointer to the C struct
    _internal: *c.aisap_AppImage = undefined,

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
            .name = path,
            .path = path,
            ._internal = &c_ai,
        };

        c_ai._parent = &ai;

        return ai;
    }

    // Find the offset of the internal read-only filesystem
    pub fn offset(ai: *AppImage) !u64 {
        var f = try fs.cwd().openFile(span(ai._internal.path), .{});
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

    pub fn mount(ai: *AppImage) !void {
        //        var sqfs = Squash{};
        _ = ai;
        //        const err = sqfs.lookup("test.sfs");
        //        std.debug.print("test {d}\n", .{err});
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
