const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

const c = @cImport({
    @cInclude("aisap.h");
});

pub const AppImage = @import("src/AppImage.zig").AppImage;

pub const BWrapError = error{
    GeneralError,
    InvalidSyntax,
};

fn BWrapErrorFromInt(err: i32) BWrapError!void {
    switch (err) {
        1 => return BWrapError.GeneralError,
        2 => return BWrapError.InvalidSyntax,
        else => return,
    }
}

// Compile in Bubble Wrap and call it as if it were a library
// The C symbol must first be exposed
extern fn bwrap_main(argc: i32, argv: [*c]const [*c]const u8) i32;
fn bwrap(allocator: *std.mem.Allocator, args: []const []const u8) !void {
    var result = try allocator.alloc([*]const u8, args.len + 1);

    // Set ARGV0 then iterate through the slice and convert it to a C char**
    result[0] = "bwrap";
    for (args) |arg, idx| {
        result[idx + 1] = @ptrCast([*]const u8, arg.ptr);
    }

    // Convert the exit code to a Zig error
    return BWrapErrorFromInt(bwrap_main(@intCast(i32, args.len + 1), result.ptr));
}
