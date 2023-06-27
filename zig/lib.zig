const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

const c = @cImport({
    @cInclude("aisap.h");
});

pub const AppImage = @import("lib/AppImage.zig").AppImage;

pub const BWrapError = error{
    GeneralError,
    InvalidSyntax,
};

fn BWrapErrorFromInt(err: c_int) BWrapError!void {
    return switch (err) {
        0 => {},
        2 => BWrapError.InvalidSyntax,

        else => BWrapError.GeneralError,
    };
}

// Compile in Bubble Wrap and call it as if it were a library
// The C symbol must first be exposed
extern fn bwrap_main(argc: c_int, argv: [*c]const [*c]const u8) c_int;
fn bwrap(allocator: *std.mem.Allocator, args: []const []const u8) !void {
    var result = try allocator.alloc([*]const u8, args.len + 1);

    // Set ARGV0 then iterate through the slice and convert it to a C char**
    result[0] = "bwrap";
    for (args, 1..) |arg, idx| {
        result[idx] = arg.ptr;
    }

    // Convert the exit code to a Zig error
    return BWrapErrorFromInt(bwrap_main(@intCast(args.len + 1), result.ptr));
}
