const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

const Md5 = std.crypto.hash.Md5;
const ArrayList = std.ArrayList;

const c = aisap.c;

const aisap = @import("AppImage.zig");
const AppImage = aisap.AppImage;

const CAppImageError = enum(u8) {
    ok = 0,
    err, // Generic error
    invalid_magic,
    no_desktop_entry,
    invalid_desktop_entry,
    invalid_socket,
    no_space_left,
};

export fn aisap_appimage_new(path: [*:0]const u8, err: *CAppImageError) aisap.c_AppImage {
    const path_len = std.mem.len(path);
    return aisap_appimage_newn(path, path_len, err);
}

export fn aisap_appimage_newn(path: [*]const u8, path_len: usize, err: *CAppImageError) aisap.c_AppImage {
    var allocator = std.heap.c_allocator;

    var ai: aisap.c_AppImage = undefined;
    var zig_ai = allocator.create(AppImage) catch unreachable;
    zig_ai.* = AppImage.init(allocator, path[0..path_len]) catch {
        err.* = .err;
        return ai;
    };

    ai.path = zig_ai.path.ptr;
    ai.path_len = zig_ai.path.len;
    ai.name = zig_ai.name.ptr;
    ai.name_len = zig_ai.name.len;

    zig_ai._internal = &ai;
    ai._zig_parent = zig_ai;

    // Init Go AppImage to access its functions until they're replaced
    c.aisap_appimage_init_go(&ai, path, @ptrCast(err));

    return ai;
}

// TODO: error handling
export fn aisap_appimage_mount(ai: *aisap.c_AppImage, path: ?[*:0]const u8, err: *CAppImageError) void {
    const path_len = if (path) |p| std.mem.len(p) else 0;

    aisap_appimage_mountn(ai, path, path_len, err);
}

// TODO: error handling
export fn aisap_appimage_mountn(ai: *aisap.c_AppImage, path: ?[*]const u8, path_len: usize, err: *CAppImageError) void {
    _ = err;

    if (path) |p| {
        getParent(ai).mount(.{
            .path = p[0..path_len],
        }) catch {
            @panic("mount error with path");
        };

        return;
    }

    getParent(ai).mount(.{}) catch @panic("mount error no path");
}

export fn aisap_appimage_destroy(ai: *aisap.c_AppImage) void {
    c.aisap_appimage_destroy_go(ai);
    getParent(ai).deinit();
}

export fn aisap_appimage_md5(ai: *aisap.c_AppImage, buf: [*]u8, buf_len: usize, errno: *CAppImageError) [*:0]const u8 {
    return (aisap.md5FromPath(ai.path[0..ai.path_len], buf[0..buf_len]) catch |err| {
        errno.* = switch (err) {
            // This should be the only error ever given from this function
            aisap.AppImageError.NoSpaceLeft => .no_space_left,

            else => .err,
        };

        unreachable;
    }).ptr;
}

//export fn aisap_appimage_tempdir(ai: *aisap.c_AppImage) [*:0]const u8 {
//    return ai.temp_dir;
//}
//
//export fn aisap_appimage_mountdir(ai: *aisap.c_AppImage) [*:0]const u8 {
//    return ai.mount_dir;
//}
//
//export fn aisap_appimage_runid(ai: *aisap.c_AppImage) [*:0]const u8 {
//    return ai.run_id;
//}
//

// This exposes an `exported` Go function that I've chosen not to include in
// the header file. This is only meant to be used to obtain the wrap arguments
// from Go as I've been unable to convert a []string to a char**

// This will be used until AppImage.WrapArgs can be completely re-implemented
// in Zig
extern fn aisap_appimage_wraparg_next_go(*aisap.c_AppImage, *i32) ?[*:0]const u8;

// Returned memory must be freed
export fn aisap_appimage_wrapargs(ai: *c.aisap_appimage, err: *CAppImageError) [*:null]?[*:0]const u8 {
    return getParent(ai).wrapArgsZ(std.heap.c_allocator) catch {
        err.* = .err;
        unreachable;
    };
}

// TODO: Re-implement wrap.go in Zig

extern fn bwrap_main(argc: i32, argv: [*c]const [*c]const u8) i32;

fn aisap_appimage_sandbox(ai: *c.aisap_appimage, argc: i32, args: [*c]const [*c]const u8) i32 {
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
    while (aisap_appimage_wraparg_next_go(ai, &len)) |arg| {
        var str: []const u8 = undefined;
        str.len = @intCast(len);
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

/// Get the SquashFS image offset of the AppImage
/// Offset is stored in `off`, returns error code
export fn aisap_appimage_offset(ai: *c.aisap_appimage, errno: *CAppImageError) usize {
    errno.* = .ok;
    const off = getParent(ai).offset() catch |err| {
        errno.* = switch (err) {
            aisap.AppImageError.InvalidMagic => .invalid_magic,
            aisap.AppImageError.NoDesktopEntry => .no_desktop_entry,
            aisap.AppImageError.InvalidDesktopEntry => .invalid_desktop_entry,
            aisap.AppImageError.InvalidSocket => .invalid_socket,

            else => .err,
        };

        return 0;
    };

    return off;
}

/// This function allocates memory on the heap, the caller is responsible
/// for freeing it
export fn appimage_get_md5(path: [*:0]const u8) [*:0]const u8 {
    var buf = std.heap.page_allocator.alloc(u8, Md5.digest_length * 2 + 1) catch unreachable;
    return (aisap.md5FromPath(std.mem.span(path), buf) catch unreachable).ptr;
}

export fn appimage_get_payload_offset(path: [*:0]const u8) std.os.off_t {
    // TODO: handle this error
    return @intCast(
        aisap.offsetFromPath(std.mem.span(path)) catch unreachable,
    );
}

fn getParent(ai: *c.aisap_appimage) *AppImage {
    return @as(*AppImage, @ptrCast(@alignCast(ai._zig_parent.?)));
}

//export fn appimage_get_type(path: [*c]u8) i32 {
//    var ai: aisap.c_AppImage = undefined;
//    _ = c.aisap_new_appimage(&ai, path);
//
//    defer c.aisap_appimage_destroy(&ai);
//    return ai.ai_type;
//}
//
//export fn appimage_get_payload_offset(path: [*c]u8) c.off_t {
//    var ai: c.aisap_appimage = undefined;
//    _ = c.aisap_new_appimage(&ai, path);
//
//    var off: u64 = 0;
//    _ = aisap_appimage_offset(&ai, &off);
//    var off_s: c.off_t = @intCast(off);
//
//    defer c.aisap_appimage_destroy(&ai);
//    return off_s;
//}
