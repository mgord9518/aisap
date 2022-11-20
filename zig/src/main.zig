const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

const c = @cImport({
    @cInclude("aisap.h");
});

pub const AppImage = struct {
    name: []const u8,
    path: []const u8,

    // The internal pointer to the C struct
    _internal: *c.aisap_AppImage = undefined,

    // TODO: Use this to replace the Go implemenation
    // Once the Zig version is up to par, the Zig -> C bindings will call to
    // the parent pointer's methods. This cannot currently be done as Go
    // doesn't allow Go pointers to be passed to C
    pub fn init(path: []const u8) !AppImage {
        // Create the AppImage type for the C binding
        var c_ai = c.aisap_AppImage{
            .name      = toMut(path).ptr,
            .path      = toMut(path).ptr,
            .data_dir  = undefined,
            .temp_dir  = undefined,
            .root_dir  = undefined,
            .mount_dir = undefined,
            .md5       = undefined,
            .run_id    = undefined,
            ._index    = 0,
            ._parent   = undefined,
            .ai_type   = 2,
        };

        var ai = AppImage{
            .name = path,
            .path = path,
            ._internal = &c_ai,
        };

        c_ai._parent = &ai;

        return ai;
    }

    // Find the offset of the internal read-only filesystem
    pub fn offset(ai: *AppImage) !u64 {
        var f = try fs.cwd().openFile(span(ai._internal.path), .{});
        const hdr = try std.elf.Header.read(f);

        return hdr.shoff + hdr.shentsize * hdr.shnum;
    }

    pub fn wrapArgs(ai: *AppImage, allocator: *std.mem.Allocator) [][]const u8 {
        // Need an allocator as the size of `cmd_args` will change size

        var cmd_args: [][]const u8 = undefined;
        cmd_args = allocator.alloc([]const u8, 2) catch unreachable;
        cmd_args[0] = "test";
        cmd_args[1] = "test2";

        _ = ai;

        return cmd_args;
    }

    // This can't be finished until AppImage.wrapArgs works correctly
    pub fn sandbox(ai: *AppImage, allocator: *std.mem.Allocator) !void {
        const cmd = [_][]const u8 {
            "--ro-bind", "/", "/",
            "sh",
        };

        _ = ai;
        _ = try bwrap(allocator, &cmd);
    }
};

pub const BWrapError = error {
    GeneralError,
    InvalidSyntax,
};

fn BWrapErrorFromInt(err: i32) BWrapError!void {
    switch (err) {
        1    => return BWrapError.GeneralError,
        2    => return BWrapError.InvalidSyntax,
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
    return (BWrapErrorFromInt(
        bwrap_main(@intCast(i32, args.len + 1), result.ptr))
    );
}

export fn aisap_appimage_name(ai: *c.aisap_AppImage) [*:0]const u8 {
    return ai.name;
}

export fn aisap_appimage_md5(ai: *c.aisap_AppImage) [*:0]const u8 {
    return ai.md5;
}

export fn aisap_appimage_tempdir(ai: *c.aisap_AppImage) [*:0]const u8 {
    return ai.temp_dir;
}

export fn aisap_appimage_mountdir(ai: *c.aisap_AppImage) [*:0]const u8 {
    return ai.mount_dir;
}

export fn aisap_appimage_runid(ai: *c.aisap_AppImage) [*:0]const u8 {
    return ai.run_id;
}

export fn aisap_appimage_type(ai: *c.aisap_AppImage) i32 {
    return ai.ai_type;
}

// TODO: Re-implement wrap.go in Zig

fn aisap_appimage_sandbox(ai: *c.aisap_AppImage, argc: i32, args: [*c][*c]u8) i32 {
    //std.debug.print("{s}", .{argv[0]});
    _ = args;
    _ = argc;
    _ = ai;

    return 0;
}

// Mounts AppImage to ai.mount_dir;
//export fn aisap_appimage_mount(ai *c.aisap_AppImage) {
//}

// Get the SquashFS image offset of the AppImage
// Offset is stored in `off`, returns error code
export fn aisap_appimage_offset(ai: *c.aisap_AppImage, off: *u64) i32 {
    var f = fs.cwd().openFile(span(ai.path), .{}) catch return 1;

    // Get header
    // Buffer of 19 as it is shImg's header size.
    // Normal AppImages can be detected by reading 10 bytes
    var hdr_buf: [19]u8 = undefined;
    _ = f.read(hdr_buf[0..]) catch return 1;
    f.seekTo(0) catch return 1;

    if (std.mem.eql(u8, hdr_buf[0..], "#!/bin/sh\n#.shImg.#")) {
        // Read shImg offset
        var buf_reader = io.bufferedReader(f.reader());
        var in_stream = buf_reader.reader();

        // Small buffer needed, the `sfs_offset` line should be well below this amount
        var buf: [256]u8 = undefined;

        var line: u32 = 0;
        while (in_stream.readUntilDelimiterOrEof(&buf, '\n') catch return 3) |text| {
            // Iterated over too many lines, not shImg
            line += 1;
            if (line > 512) return 5;

            if (text.len > 10 and std.mem.eql(u8, text[0..11], "sfs_offset=")) {
                var it = std.mem.tokenize(u8, text, "=");

                // Throw away first chunk, should equal `sfs_offset`
                _ = it.next();

                off.* = std.fmt.parseInt(u64, it.next().?, 0) catch return 4;
                return 0;
            }
        }
    }

    // Read ELF offset
    const hdr = std.elf.Header.read(f) catch return 2;

    off.* = hdr.shoff + hdr.shentsize * hdr.shnum;

    return 0;
}

fn toMut(str: []const u8) []u8 {
    var buf: [256]u8 = undefined;
    var mut: []u8    = buf[0..str.len];
    std.mem.copy(u8, mut, str);

    return mut;
}
