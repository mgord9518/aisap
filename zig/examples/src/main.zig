// Simple program for testing the Zig aisap implementation
// This just dumps whatever debug information is useful for the current stage
// of development. This will eventually become an actual example
// Usage: `./zig-out/bin/example [APPIMAGE]`

const std = @import("std");
const aisap = @import("aisap");

const AppImage = aisap.AppImage;

pub fn main() !void {
    var allocator = std.heap.c_allocator;
    var args = std.process.args();

    // Skip arg[0]
    const arg0 = args.next().?;

    const arg1 = args.next() orelse {
        std.debug.print("usage: {s} [appimage]\n", .{arg0});
        return;
    };

    var ai = AppImage.init(allocator, arg1) catch |err| {
        std.debug.print("error opening application bundle: {!}\n", .{err});
        return;
    };
    defer ai.deinit();

    var md5_buf: [33]u8 = undefined;

    std.debug.print("{s}\n", .{ai.name});
    std.debug.print("{s}\n", .{ai.desktop_entry});
    std.debug.print("{}\n", .{try ai.permissions(allocator)});

    var buf: [32]u8 = undefined;

    std.debug.print("{s}\n", .{
        try std.fmt.bufPrint(&buf, "{s}", .{
            try ai.md5(&md5_buf),
        }),
    });

    try ai.mount(.{});
    //    ai.mount();
}
