// Simple program for testing my squashfuse bindings and aisap implementation

const std = @import("std");
const warn = @import("std").debug.print;
const aisap = @import("aisap");

const AppImage = aisap.AppImage;
const SquashFs = @import("squashfuse").SquashFs;

// Been using my AppImage build of Go to test it, obviously if anyone else
// wants to use this pre-alpha test you'll need to change the path
const test_ai_path = "/home/mgord9518/Git/yabg/YABG-0.0.1-x86_64.AppImage";

pub fn main() !void {
    var ai = try AppImage.init(test_ai_path);
    defer ai.deinit();

    try ai.mount();

    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    var allocator = arena.allocator();
    var wrap_args = ai.wrapArgs(allocator);

    const perms = try ai.permissions(std.heap.page_allocator);
    std.debug.print("{d}\n", .{perms.level});
    std.debug.print("{s}\n", .{perms.files});
    std.debug.print("{s}\n", .{perms.sockets});
    std.debug.print("{s}\n", .{perms.devices});

    // var sfs = try SquashFs.init(test_ai_path, 544156);
    //var walker = try sfs.walk("");

    // Iterate over the AppImage's internal SquashFS image
    // This is to test my squashfuse bindings
    //    while (try walker.next()) |entry| {
    //        std.debug.print("{any}\n", .{entry.kind});
    //        std.debug.print("{s} ", .{entry.path});
    //        std.debug.print("{s}\n", .{entry.basename});
    //
    //        if (entry.kind == .File and std.mem.eql(u8, entry.path, ".DirIcon")) {
    //            //            var inode = sfs.getInode(entry.id) catch unreachable;
    //
    //            // Read 8MiB at a time (fast buffer size for modern hardware)
    //            var buf: [8 * 1024 * 1024]u8 = undefined;
    //
    //            try sfs.extract(&buf, entry, "/tmp/testfile");
    //
    //            std.debug.print("Extracted .DirIcon to `/tmp/testfile`\n", .{});
    //
    //            break;
    //        }
    //    }

    // Testing WIP SquashFS library
    //    std.debug.print("SQFS\n", .{});
    //    std.debug.print("sfs.fd: {d}\n", .{sfs.internal.fd});
    //    std.debug.print("sfs.version: {d}.{d}\n", .{ sfs.version.major, sfs.version.minor });

    std.debug.print("AISAP\n", .{});
    std.debug.print("ai name: {s}\n", .{ai.name});
    std.debug.print("wrap args: {d}\n", .{wrap_args.len});
}
