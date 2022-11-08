const std = @import("std");
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;
const c = @cImport({
    @cInclude("aisap.h");
});

pub const AppImage = struct {
    name:   []const u8,
    path:   []const u8,
    offset:   u64,

    // The internal pointer to the C struct
    _internal: *c.aisap_AppImage,

    // TODO: Use this to replace the Go implemenation
    // Once the Zig implementation is complete,
    export fn init(ai: *AppImage, path: []u8) !AppImage {
        ai.name = path;
        return ai;
    }

    // Find the offset of the internal read-only filesystem
    // TODO: implement for shImg
    export fn offset(ai: *AppImage) !u64 {
        var f = try fs.cwd().openFile(span(ai._internal.path), .{});
        const hdr = try std.elf.Header.read(f);

        return hdr.shoff + hdr.shentsize * hdr.shnum;
    }

    export fn wrapArgs(ai: *AppImage) [][]const u8 {
        var cmd_args = [_][*]const u8 {
            "bwrap", "--setenv", "TMPDIR"
        };

        _ = ai;

        return cmd_args;
    }

};

// These wrapper will likely be removed in v1.0
// To access this information, just use the struct fields
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

fn aisap_appimage_sandbox(ai: *c.aisap_AppImage, argc: i32, args: [*c][*c] u8) i32 {
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
    const hdr = std.elf.Header.read(f) catch return 2;

    off.* = hdr.shoff + hdr.shentsize * hdr.shnum;

    return 0;
}
