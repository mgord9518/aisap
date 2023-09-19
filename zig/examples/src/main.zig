// Simple program for testing the Zig aisap implementation
// This just dumps whatever debug information is useful for the current stage
// of development. This will eventually become an actual example
// Usage: `./zig-out/bin/example [APPIMAGE]`

const std = @import("std");
const aisap = @import("aisap");

const AppImage = aisap.AppImage;

pub const ParsedStruct = struct {
    perms: []AppImage.JsonPermissions,
};

fn printFilesystem(slice: []AppImage.FilesystemPermissions) void {
    std.debug.print("[", .{});

    for (slice, 0..) |element, idx| {
        if (idx != slice.len and idx != 0) {
            std.debug.print(",", .{});
        }

        std.debug.print(" {s}", .{element.src_path});
        std.debug.print(":{s}", .{if (element.writable) "rw" else "ro"});
    }

    std.debug.print(" ]\n", .{});
}

fn printSockets(slice: []AppImage.SocketPermissions) void {
    std.debug.print("[", .{});

    for (slice, 0..) |element, idx| {
        if (idx != slice.len and idx != 0) {
            std.debug.print(",", .{});
        }

        std.debug.print(" {s}", .{@tagName(element)});
    }

    std.debug.print(" ]\n", .{});
}

pub fn main() !void {
    var allocator = std.heap.page_allocator;
    var args = std.process.args();

    // Skip arg[0]
    const arg0 = args.next().?;

    const arg1 = args.next() orelse {
        std.debug.print("prints assorted debug info about an AppImage\n", .{});
        std.debug.print("usage: {s} [appimage]\n", .{arg0});
        return;
    };

    var ai = AppImage.init(allocator, arg1) catch |err| {
        std.debug.print("error opening application bundle: {!}\n", .{err});
        return;
    };
    defer ai.deinit();

    var md5_buf: [33]u8 = undefined;

    const permissions = try ai.permissions(
        allocator,
    ) orelse {
        std.debug.print("no permissions found\n", .{});
        return;
    };

    std.debug.print("permissions (from: {s}):\n", .{@tagName(permissions.origin)});
    std.debug.print("  level: {d}\n", .{permissions.level});
    std.debug.print("  filesystem: ", .{});
    if (permissions.filesystem) |filesystem| {
        printFilesystem(filesystem);
    } else {
        std.debug.print("[]\n", .{});
    }

    std.debug.print("  sockets: ", .{});
    if (permissions.sockets) |sockets| {
        printSockets(sockets);
    } else {
        std.debug.print("[]\n", .{});
    }

    std.debug.print("{s}\n", .{ai.name});
    std.debug.print("desktop: {s}\n", .{ai.desktop_entry});
    std.debug.print("type: {}\n", .{ai.kind});

    std.debug.print("md5: {s}\n", .{
        try ai.md5(&md5_buf),
    });

    try ai.mount(.{});

    const wrapArgs = try ai.wrapArgs(allocator);
    printWrapArgs(wrapArgs);

    try ai.sandbox(.{
        //.args = &[_][]const u8{"build"},
    });
}

fn printWrapArgs(args: []const []const u8) void {
    for (args) |arg| {
        std.debug.print("{s} ", .{arg});
    }
}
