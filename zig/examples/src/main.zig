const std = @import("std");
const warn = @import("std").debug.print;
const aisap = @import("aisap");
const AppImage = aisap.AppImage;

pub fn main() !void {
    // Prints to stderr (it's a shortcut based on `std.io.getStdErr()`)

    // stdout is for the actual output of your application, for example if you
    // are implementing gzip, then only the compressed bytes should be sent to
    // stdout, not any debugging messages.

    var ai = try AppImage.init("/home/mgord9518/.local/bin/go");
    //    _ = ai;
        var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
        defer arena.deinit();
        var allocator = arena.allocator();
    var wrap_args = ai.wrapArgs(&allocator);

    std.debug.print("ai name: {s}\n", .{ai.name});
    std.debug.print("wrap args: {d}\n", .{wrap_args.len});
    printMap(wrap_args);


}

fn printMap(map: []const []const u8) void {
    for (map) |row| {
        for (row) |tile| {
            warn("{c}", .{tile});
        }
        warn("\n", .{});
    }
}

test "simple test" {
    var list = std.ArrayList(i32).init(std.testing.allocator);
    defer list.deinit(); // try commenting this out and see if zig detects the memory leak!
    try list.append(42);
    try std.testing.expectEqual(@as(i32, 42), list.pop());
}
