const std = @import("std");
const testing = std.testing;
const squashfuse = @import("squashfuse");
const SquashFs = squashfuse.SquashFs;

pub const AppImage = @This();

sqfs: *SquashFs,
kind: Kind,
allocator: std.mem.Allocator,

pub const Kind = enum {
    shimg,
    type1,
    type2,
};

pub fn open(allocator: std.mem.Allocator, file: std.fs.File) !AppImage {
    const kind = try getKind(file);
    const offset = switch (kind) {
        .type1, .type2 => try getOffsetFromElf(file),
        .shimg => try getOffsetFromShimg(file),
    };

    try file.seekTo(0);

    const sqfs = try SquashFs.open(allocator, file, .{
        .offset = offset,
    });
    errdefer sqfs.close();

    var appimage = AppImage{
        .allocator = allocator,
        .sqfs = sqfs,
        .kind = kind,
    };

    var buf: [4096]u8 = undefined;

    _ = try appimage.loadDesktopEntry(&buf);

    return appimage;
}

pub fn close(appimage: *AppImage) void {
    appimage.sqfs.close();
}

fn loadDesktopEntry(appimage: *AppImage, buf: []u8) !usize {
    var found = false;

    var root = appimage.sqfs.root();
    var it = try root.iterate();

    while (try it.next()) |entry| {
        const extension = std.fs.path.extension(entry.name);
        if (!std.mem.eql(u8, extension, ".desktop")) continue;

        found = true;

        var file = try root.openFile(entry.name, .{});
        _ = try file.reader().readAll(buf);
    }

    return 0;
}

fn getKind(file: std.fs.File) !Kind {
    try file.seekTo(0);

    var buf: [19]u8 = undefined;
    const read_len = try file.readAll(&buf);

    if (read_len < 19) return error.InvalidFormat;

    if (std.mem.eql(u8, &buf, "#!/bin/sh\n#.shImg.#")) {
        return .shimg;
    } else if (std.mem.eql(u8, buf[0..4], "\x7fELF")) {
        if (!std.mem.eql(u8, buf[8..10], "AI")) {
            return error.InvalidFromat;
        }

        return switch (buf[10]) {
            1 => .type1,
            2 => .type2,

            else => error.UnknownAppImageVersion,
        };
    }

    return error.InvalidFormat;
}

fn getOffsetFromElf(file: std.fs.File) !u64 {
    try file.seekTo(0);

    const header = try std.elf.Header.read(file);
    return header.shoff + (header.shentsize * header.shnum);
}

fn getOffsetFromShimg(file: std.fs.File) !u64 {
    try file.seekTo(0);

    var buf_reader = std.io.bufferedReader(file.reader());
    var in_stream = buf_reader.reader();

    // Small buffer needed, the `sfs_offset` line should be well below this amount
    var buf: [256]u8 = undefined;

    var line: u32 = 0;
    while (try in_stream.readUntilDelimiterOrEof(&buf, '\n')) |text| {
        // Iterated over too many lines, not shImg
        line += 1;
        if (line > 512) return error.NotShimg;

        if (text.len > 10 and std.mem.eql(u8, text[0..11], "sfs_offset=")) {
            var it = std.mem.tokenize(u8, text, "=");

            // Throw away first chunk, should equal `sfs_offset`
            _ = it.next();

            try file.seekTo(0);

            return try std.fmt.parseInt(
                u64,
                it.next() orelse return error.NotShimg,
                0,
            );
        }
    }

    return error.InvalidFormat;
}
