const std = @import("std");
const warn = @import("std").debug.print;
const aisap = @import("aisap");
const AppImage = aisap.AppImage;
const squashfuse = @import("squashfuse");
const SquashFs = squashfuse.SquashFs;

pub fn main() !void {
    var ai = try AppImage.init("/home/mgord9518/.local/bin/go");
    try ai.mount();
    //    _ = ai;
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    var allocator = arena.allocator();
    var wrap_args = ai.wrapArgs(allocator);

    var sfs = try SquashFs.init("/home/mgord9518/.local/bin/go", 544156);
    var walker = try sfs.walk("");
    while (try walker.next()) |entry| {
        std.debug.print("{any}\n", .{entry.kind});
        std.debug.print("{s}\n", .{entry.path});
        std.debug.print("{s}\n", .{entry.basename});
    }

    // Testing WIP SquashFS library
    std.debug.print("SQFS\n", .{});
    std.debug.print("sfs.fd: {d}\n", .{sfs.internal.fd});
    std.debug.print("sfs.version: {d}.{d}\n", .{ sfs.version.major, sfs.version.minor });

    std.debug.print("AISAP\n", .{});
    std.debug.print("ai name: {s}\n", .{ai.name});
    std.debug.print("wrap args: {d}\n", .{wrap_args.len});

    //    try ai.sandbox(&allocator);

    var i: i32 = 0;
    while (i < 50) {
        //        printMap(wrap_args);
        //        std.debug.print("{d}\n", .{i});
        i += 1;
    }
}

fn printMap(map: []const []const u8) void {
    for (map) |row| {
        for (row) |tile| {
            warn("{c}", .{tile});
        }
        warn("\n", .{});
    }
}
