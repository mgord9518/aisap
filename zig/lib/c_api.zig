const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

const Md5 = std.crypto.hash.Md5;
const ArrayList = std.ArrayList;

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

// This exposes an `exported` Go function that I've chosen not to include in
// the header file. This is only meant to be used to obtain the wrap arguments
// from Go as I've been unable to convert a []string to a char**

// This will be used until AppImage.WrapArgs can be completely re-implemented
// in Zig
extern fn aisap_appimage_wraparg_next(*c.aisap_AppImage, *i32) ?[*:0]const u8;

export fn aisap_appimage_wrapargs(ai: *c.aisap_AppImage) [*:0]const u8 {
    var it: i32 = undefined;
    //while (aisap_appimage_wraparg_next(ai, &it)) |arg| {
    //    std.debug.print("{s} ({d}) ", .{ arg, it });
    //}

    _ = aisap_appimage_sandbox(ai, 0, null);

    const ret = aisap_appimage_wraparg_next(ai, &it);
    if (ret != null) {
        return ret.?;
    }

    return "failed";
}

// TODO: Re-implement wrap.go in Zig

extern fn bwrap_main(argc: i32, argv: [*c]const [*c]const u8) i32;

fn aisap_appimage_sandbox(ai: *c.aisap_AppImage, argc: i32, args: [*c]const [*c]const u8) i32 {
    var buf: [10000]u8 = undefined;
    var fba = std.heap.FixedBufferAllocator.init(&buf);
    var allocator = fba.allocator();

    _ = argc;
    _ = args;

    // Build char** from the aisap-Go `WrapArgs` method. This will be replaced
    // once I can re-implement it in Zig
    var list = ArrayList([]const u8).init(allocator);
    var len: i32 = undefined;
    defer list.deinit();

    // Since this is just bwrap's main() function renamed and built into a lib,
    // argv[0] should be set to `bwrap`
    var it: i32 = 0;

    list.append("bwrap") catch return 3;
    while (aisap_appimage_wraparg_next(ai, &len)) |arg| {
        var str: []const u8 = undefined;
        str.len = @intCast(usize, len);
        str.ptr = arg;

        list.append(str) catch return 3;
        it += 1;
    }

    for (list.items) |str| {
        std.debug.print("{s} {d}", .{ str, it });
    }

    // TODO: add args to command before executing
    _ = args;
    _ = argc;

    _ = std.ChildProcess.exec(.{ .allocator = allocator, .argv = list.items }) catch return 127;

    return 0;
}

//fn aisap_appimage_sandbox(ai: *c.aisap_AppImage, argc: i32, args: [*c]const [*c]const u8) i32 {
//    var buf: [10000]u8 = undefined;
//    var fba = std.heap.FixedBufferAllocator.init(&buf);
//    var allocator = fba.allocator();
//
//    // Build char** from the aisap-Go `WrapArgs` method. This will be replaced
//    // once I can re-implement it in Zig
//    var list = ArrayList([*c]const u8).init(allocator);
//    var len: i32 = undefined;
//
//    // Since this is just bwrap's main() function renamed and built into a lib,
//    // argv[0] should be set to `bwrap`
////    list.append("bwrap") catch return 3;
//    var it: i32 = 0;
//    while (aisap_appimage_wraparg_next(ai, &len)) |arg| {
//        list.append(arg) catch return 3;
//        it += 1;
//    }
//
//    //    for (list.items) |str| {
//    //        std.debug.print("{s} {d}", .{ str, it });
//    //    }
//
//    //std.debug.print("{d}", .{bwrap_main(it + 1, list.items.ptr)});
//    std.debug.print("{d}", .{std.os.execvpeZ("bwrap", list.items.pt)});
//
//    //std.debug.print("{s}", .{argv[0]});
//    _ = args;
//    _ = argc;
//    //    _ = ai;
//
//    return 0;
//}

// Mounts AppImage to ai.mount_dir;
//export fn aisap_appimage_mount(ai *c.aisap_AppImage) {
//}

/// Get the SquashFS image offset of the AppImage
/// Offset is stored in `off`, returns error code
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

/// This function doesn't actually require opening an AppImage, just calculates
/// the MD5 using its path
export fn appimage_get_md5(path: [*:0]const u8) [*c]const u8 {
    var print_buf: [512]u8 = undefined;

    var buf: [Md5.digest_length]u8 = undefined;

    // Generate the MD5
    var h = Md5.init(.{});
    h.update(std.fmt.bufPrint(&print_buf, "file://{s}", .{path}) catch return null);
    h.final(&buf);

    // Format as hexadecimal instead of raw bytes
    const ret = std.fmt.bufPrint(&print_buf, "{x}", .{std.fmt.fmtSliceHexLower(&buf)}) catch unreachable;

    // Append null char for C
    print_buf[ret.len] = 0;

    return ret.ptr;
}

export fn appimage_get_type(path: [*c]u8) i32 {
    var ai: c.aisap_AppImage = undefined;
    _ = c.aisap_new_appimage(&ai, path);

    defer c.aisap_appimage_destroy(&ai);
    return ai.ai_type;
}

export fn appimage_get_payload_offset(path: [*c]u8) c.off_t {
    var ai: c.aisap_AppImage = undefined;
    _ = c.aisap_new_appimage(&ai, path);

    var off: u64 = 0;
    _ = aisap_appimage_offset(&ai, &off);
    var off_s = @intCast(c.off_t, off);

    defer c.aisap_appimage_destroy(&ai);
    return off_s;
}
