const std = @import("std");
const testing = std.testing;
const squashfuse = @import("squashfuse");
const SquashFs = squashfuse.SquashFs;

pub const AppImage = @This();

sqfs: SquashFs,
kind: Kind,
allocator: std.mem.Allocator,

pub const Kind = enum {
    shimg,
    type1,
    type2,
};

pub fn open(allocator: std.mem.Allocator, path: []const u8) !AppImage {
    const cwd = std.fs.cwd();
    const file = try cwd.openFile(path, .{});

    std.debug.print("offset {d}\n", .{try offsetFromElf(file)});

    return .{
        .sqfs = undefined,
        .kind = undefined,
        .allocator = allocator,
    };
}

pub fn close(appimage: *AppImage) void {
    _ = appimage;
    return;
}

fn offsetFromElf(file: std.fs.File) !u64 {
    const header = try std.elf.Header.read(file);

    return header.shoff + (header.shentsize * header.shnum);
}

//fn offsetFromShimg(path: []const u8) u64 {}
