// Simple program for testing my squashfuse bindings and aisap implementation
// I'll probably turn this into a re-implementation of the main squashfuse
// binary program after I get everything working

const std = @import("std");
const aisap = @import("aisap");

const AppImage = aisap.AppImage;

//const test_ai_path = "/home/mgord9518/Downloads/Powder_Toy-96.2-x86_64.AppImage";
const test_ai_path = "/home/mgord9518/Downloads/Spelunky-Classic-HD-Minimal-x86-64.AppImage";

// I'll add this automatically soon, but currently, the `-s` flag must be
// supplied as it only works single-threaded
pub fn main() !void {
    var allocator = std.heap.c_allocator;

    var ai = AppImage.init(allocator, test_ai_path) catch |err| {
        std.debug.print("error: {!}\n", .{err});
        return;
    };
    defer ai.deinit();

    std.debug.print("{s}\n", .{ai.name});
    std.debug.print("{s}\n", .{ai.desktop_entry});
    std.debug.print("{s}\n", .{ai.desktop_entry.ptr});
    std.debug.print("{}\n", .{try ai.permissions(allocator)});

    var buf: [32]u8 = undefined;

    std.debug.print("{s}\n", .{try std.fmt.bufPrint(&buf, "{s}", .{ai.md5()})});
}
