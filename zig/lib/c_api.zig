const std = @import("std");
const io = std.io;
const fs = std.fs;
const span = std.mem.span;
const expect = std.testing.expect;

const Md5 = std.crypto.hash.Md5;
const ArrayList = std.ArrayList;

const c = @cImport({
    @cInclude("aisap.h");
});

const aisap = @import("appimage.zig");
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

export fn aisap_appimage_new(path: [*:0]const u8, err: *CAppImageError) c.aisap_appimage {
    const path_len = std.mem.len(path);
    return aisap_appimage_newn(path, path_len, err);
}

export fn aisap_appimage_newn(
    path: [*]const u8,
    path_len: usize,
    err: *CAppImageError,
) c.aisap_appimage {
    var allocator = std.heap.c_allocator;

    var ai: c.aisap_appimage = undefined;
    const zig_ai = allocator.create(AppImage) catch {
        err.* = .err;
        return ai;
    };

    zig_ai.* = AppImage.init(allocator, path[0..path_len]) catch {
        err.* = .err;
        return ai;
    };

    ai.path = zig_ai.path.ptr;
    ai.path_len = zig_ai.path.len;
    ai.name = zig_ai.name.ptr;
    ai.name_len = zig_ai.name.len;

    ai._zig_parent = zig_ai;

    return ai;
}

export fn aisap_appimage_mount_dir(ai: *c.aisap_appimage) ?[*:0]const u8 {
    const zig_ai = getParent(ai);

    if (zig_ai.mount_dir) |mount_dir| {
        return mount_dir;
    }

    return null;
}

// TODO: error handling
export fn aisap_appimage_mount(
    ai: *c.aisap_appimage,
    path: ?[*:0]const u8,
    err: *CAppImageError,
) void {
    const path_len = if (path) |p| std.mem.len(p) else 0;

    aisap_appimage_mountn(ai, path, path_len, err);
}

// TODO: error handling
export fn aisap_appimage_mountn(
    ai: *c.aisap_appimage,
    path: ?[*]const u8,
    path_len: usize,
    err: *CAppImageError,
) void {
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

export fn aisap_appimage_destroy(ai: *c.aisap_appimage) void {
    getParent(ai).deinit();
}

export fn aisap_appimage_md5(
    ai: *c.aisap_appimage,
    buf: [*]u8,
    buf_len: usize,
    errno: *CAppImageError,
) [*:0]const u8 {
    return (aisap.md5FromPath(
        ai.path[0..ai.path_len],
        buf[0..buf_len],
    ) catch |err| {
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

// This function typically shouldn't be called on its own. Instead,
// `aisap_appimage_sandbox` should be preferred
// Returned memory must be freed
export fn aisap_appimage_wrapargs(
    c_ai: *c.aisap_appimage,
    //perms: *c.aisap_permissions,
    err: *CAppImageError,
) [*:null]?[*:0]const u8 {
    _ = c_ai;
    _ = err;
    //const ai = getParent(c_ai);

    unreachable;
    //    return ai.wrapArgsZ(ai.allocator) catch {
    //        err.* = .err;
    //        return undefined;
    //    };
}

// TODO: Re-implement wrap.go in Zig
//extern fn bwrap_main(argc: c_int, argv: [*]const [*:0]const u8) i32;

export fn aisap_appimage_sandbox(
    ai: *c.aisap_appimage,
    argc: usize,
    argv: [*:null]const ?[*:0]const u8,
    err: *CAppImageError,
) void {
    var zig_ai = getParent(ai);

    if (argc > 0) {
        var args = zig_ai.allocator.alloc([]const u8, argc) catch |e| {
            std.debug.print("{s} ERR: {!}\n", .{ @src().fn_name, e });

            err.* = .err;
            return;
        };
        defer zig_ai.allocator.free(args);

        var i: usize = 0;
        while (i < argc) {
            const arg = argv[i] orelse {
                // TODO: argc should never be longer than the length of argv,
                // so maybe a panic would be better here?
                args.len = i;

                break;
            };

            args[i] = std.mem.sliceTo(arg, '\x00');
            i += 1;
        }

        //            zig_ai.sandbox(.{
        //                .args = args,
        //            }) catch |e| {
        //                std.debug.print("{s} ERR: {!}\n", .{ @src().fn_name, e });
        //                err.* = .err;
        //            };

        return;
    }

    //    zig_ai.sandbox(.{}) catch |e| {
    //        std.debug.print("{s} ERR: {!}\n", .{ @src().fn_name, e });
    //        err.* = .err;
    //    };
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

// libappimage API
// TODO: make adding this a conditional build flag
/// This function allocates memory on the heap, the caller is responsible
/// for freeing it
export fn appimage_get_md5(path: [*:0]const u8) [*:0]const u8 {
    const allocator = std.heap.c_allocator;

    const buf = allocator.alloc(
        u8,
        Md5.digest_length * 2 + 1,
    ) catch unreachable;

    return (aisap.md5FromPath(
        std.mem.span(path),
        buf,
    ) catch unreachable).ptr;
}

export fn appimage_get_payload_offset(path: [*:0]const u8) std.posix.off_t {
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
