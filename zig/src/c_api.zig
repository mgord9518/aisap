const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

const c = @cImport({
    @cInclude("aisap.h");
    @cInclude("unistd.h");
});

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

// For ABI compat with libAppImage
export fn appimage_get_md5(path: [*]u8) [*:0]const u8 {
    var ai: c.aisap_AppImage = undefined;
    _ = c.aisap_new_appimage(&ai, path);

    defer c.aisap_appimage_destroy(&ai);
    return ai.md5;
}

export fn appimage_get_type(path: [*]u8) i32 {
    var ai: c.aisap_AppImage = undefined;
    _ = c.aisap_new_appimage(&ai, path);

    defer c.aisap_appimage_destroy(&ai);
    return ai.ai_type;
}

export fn appimage_get_payload_offset(path: [*]u8) c.off_t {
    var ai: c.aisap_AppImage = undefined;
    _ = c.aisap_new_appimage(&ai, path);

    var off: u64 = 0;
    _ = aisap_appimage_offset(&ai, &off);
    var off_s = @intCast(i64, off);

    defer c.aisap_appimage_destroy(&ai);
    return off_s;
}

fn toMut(str: []const u8) []u8 {
    var buf: [256]u8 = undefined;
    var mut: []u8 = buf[0..str.len];
    std.mem.copy(u8, mut, str);

    return mut;
}
